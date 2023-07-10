package start

import (
	"fmt"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	"go-outpost/internal/config"
	"go-outpost/internal/http-server/handlers/event"
	"go-outpost/internal/http-server/handlers/job"
	"go-outpost/internal/http-server/handlers/user/balance"
	"go-outpost/internal/http-server/model"
	resp "go-outpost/internal/lib/api/response"
	"go-outpost/internal/lib/logger/sl"
	"go-outpost/internal/repository"
	"golang.org/x/exp/slog"
	"net/http"
	"time"
)

type RouletteStart struct {
	log            *slog.Logger
	rouletteRep    repository.RouletteRepository
	rouletteBetRep repository.RouletteBetRepository
	cache          *cache.Cache
	pusher         *event.PusherEvent
	rouletteRoller *RouletteRoller
	balance        balance.Interface
}

type Roulette interface {
	Roll(roulette *model.Roulette) (winColor string, winNumber int)
}

func NewRouletteStart(
	log *slog.Logger,
	rouletteRep repository.RouletteRepository,
	rouletteBetRep repository.RouletteBetRepository,
	pusherClient *event.PusherEvent,
	rouletteRoller *RouletteRoller) *RouletteStart {
	return &RouletteStart{
		log:            log,
		rouletteRep:    rouletteRep,
		rouletteBetRep: rouletteBetRep,
		cache:          cache.New(5*time.Minute, 10*time.Minute),
		pusher:         pusherClient,
		rouletteRoller: rouletteRoller,
	}
}

func (s *RouletteStart) New() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.roulette.start.New"

		var (
			err                   error
			log                   *slog.Logger
			roulette              *model.Roulette
			winColorAndNumberData *RouletteWinColorAndNumberData
			rouletteID            int64
		)

		log = s.log.With(
			slog.String("op", op),
		)

		round := s.getRoundFromCacheOrDB()

		roulette = &model.Roulette{
			UUID:  uuid.New(),
			Round: round,
		}

		rouletteID, err = s.rouletteRep.SaveRoulette(*roulette)
		if err != nil {
			log.Error("failed to save roulette")

			render.JSON(w, r, resp.Error("failed to save roulette", http.StatusInternalServerError))

			return
		}

		log.Info("roulette created")

		roulette, err = s.rouletteRep.GetRouletteByID(rouletteID)
		if err != nil {
			log.Error("failed to get roulette")

			render.JSON(w, r, resp.Error("failed to get roulette", http.StatusInternalServerError))

			return
		}

		/*autoBets := getAutoBetsFromCache()
		if len(autoBets) != 0 {
			job.Dispatch(&RouletteBetReplicateJob{rouletteID: rouletteID}, 0)
		}*/

		s.updateCacheRound(round + 1)

		err = s.sendNewRoundEvent(roulette)
		if err != nil {
			log.Error("failed to send new round event")

			render.JSON(w, r, resp.Error("failed to send new round event", http.StatusInternalServerError))

			return
		}

		log.Info("new round event sent")

		winColorAndNumberData, err = s.rouletteRoller.Roll(roulette)
		if err != nil {
			log.Error("failed to roll roulette")

			render.JSON(w, r, resp.Error("failed to roll roulette", http.StatusInternalServerError))

			return
		}

		log.Info("roulette rolled",
			slog.Any("win_color", winColorAndNumberData.Color),
			slog.Any("win_number", winColorAndNumberData.Number))

		if err = s.handleWinners(rouletteID, winColorAndNumberData.Color); err != nil {
			log.Error("failed to handle winners")

			render.JSON(w, r, resp.Error("failed to handle winners", http.StatusInternalServerError))

			return
		}

		job.Dispatch(&RouletteUpdatePlayedAtJob{rouletteID: rouletteID}, 15*time.Second)
		job.Dispatch(&RouletteWinnerJob{winColor: winColorAndNumberData.Color, winNumber: winColorAndNumberData.Number}, 15*time.Second)
	}
}

func (s *RouletteStart) getRoundFromCacheOrDB() int64 {
	// round exists in cache then return it if not get from db and set in cache and iterate it
	round := s.getRoundFromCache()

	if round != 0 {
		return round
	}

	round, err := s.rouletteRep.GetLastRound()
	if err != nil {
		return 0
	}

	s.updateCacheRound(round)

	return round
}

func (s *RouletteStart) getRoundFromCache() int64 {
	round, found := s.cache.Get("current_round")
	if found {
		return round.(int64)
	}

	return 0
}

func (s *RouletteStart) updateCacheRound(round int64) {
	s.cache.Set("current_round", round, cache.DefaultExpiration)

	return
}

func (s *RouletteStart) sendNewRoundEvent(roulette *model.Roulette) error {
	data := map[string]interface{}{
		"uuid":       roulette.UUID.String(),
		"created_at": roulette.CreatedAt.Format(time.RFC3339),
		"round":      roulette.Round,
	}

	return s.pusher.TriggerEvent("balance-channel", "outcome-event", data)
}

func (s *RouletteStart) handleWinners(rouletteID int64, color config.Color) error {
	const op = "handlers.roulette.start.handleWinners"

	var (
		err        error
		bets       []model.RouletteBet
		multiplier int
	)

	bets, err = s.rouletteBetRep.GetBetsByRouletteIDAndColor(rouletteID, color)
	if err != nil {
		s.log.Error("failed to get winners by roulette id", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	multiplier = s.GetMultiplierByColor(color)

	for _, winner := range bets {
		if err = s.balance.Income(winner.UserID, winner.Amount*multiplier, config.Roulette); err != nil {
			s.log.Error("failed to income user balance", sl.Err(err))

			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}

func (s *RouletteStart) GetMultiplierByColor(color config.Color) int {
	colorConfig, ok := config.RouletteWheelConfig.Colors[color]
	if !ok {
		return 0
	}

	return colorConfig.Multiplier
}

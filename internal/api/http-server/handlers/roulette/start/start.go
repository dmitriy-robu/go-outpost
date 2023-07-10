package start

import (
	"fmt"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	"go-outpost/internal/api/config"
	"go-outpost/internal/api/http-server/handlers/event"
	"go-outpost/internal/api/http-server/handlers/job"
	"go-outpost/internal/api/http-server/handlers/user/balance"
	"go-outpost/internal/api/http-server/model"
	"go-outpost/internal/api/repository"
	resp "go-outpost/internal/lib/api/response"
	"go-outpost/internal/lib/logger/sl"
	"golang.org/x/exp/slog"
	"net/http"
	"time"
)

type RouletteStart struct {
	log            *slog.Logger
	rouletteRep    repository.RouletteRepository
	rouletteBetRep repository.RouletteBetRepository
	cache          *cache.Cache
	event          *event.PusherEvent
	rouletteRoller *RouletteRoller
	balance        balance.Interface
}

func NewRouletteStart(
	log *slog.Logger,
	rouletteRep repository.RouletteRepository,
	rouletteBetRep repository.RouletteBetRepository,
	eventClient *event.PusherEvent,
	rouletteRoller *RouletteRoller) *RouletteStart {
	return &RouletteStart{
		log:            log,
		rouletteRep:    rouletteRep,
		rouletteBetRep: rouletteBetRep,
		cache:          cache.New(5*time.Minute, 10*time.Minute),
		event:          eventClient,
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
			log.Error("failed to save roulette", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to save roulette", http.StatusInternalServerError))

			return
		}

		log.Info("roulette created", sl.String("roulette_id", fmt.Sprintf("%d", rouletteID)))

		roulette, err = s.rouletteRep.GetRouletteByID(rouletteID)
		if err != nil {
			log.Error("failed to get roulette", sl.Err(err))

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
			log.Error("failed to send new round event", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to send new round event", http.StatusInternalServerError))

			return
		}

		log.Info("new round event sent")

		winColorAndNumberData, err = s.rouletteRoller.Roll(roulette)
		if err != nil {
			log.Error("failed to roll roulette", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to roll roulette", http.StatusInternalServerError))

			return
		}

		log.Info("roulette rolled",
			slog.Any("win_color", winColorAndNumberData.Color),
			slog.Any("win_number", winColorAndNumberData.Number))

		if err = s.handleWinners(rouletteID, winColorAndNumberData.Color); err != nil {
			log.Error("failed to handle winners", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to handle winners", http.StatusInternalServerError))

			return
		}

		log.Info("winners handled")

		eventMessage := event.Message{
			Channel: "roulette",
			Event:   "test",
			Data: map[string]interface{}{
				"uuid":       roulette.UUID.String(),
				"created_at": roulette.CreatedAt,
				"round":      roulette.Round,
			},
		}
		job.Dispatch(&job.SendEventJob{EventMessage: eventMessage, Event: s.event}, 15*time.Second)

		job.Dispatch(&RouletteStartJob{RouletteStart: s, RouletteID: rouletteID}, 15*time.Second)
	}
}

func (s *RouletteStart) UpdateRouletteUpdateAt(rouletteID int64) {
	const op = "handlers.roulette.start.UpdateRouletteUpdateAt"

	var (
		err      error
		log      *slog.Logger
		roulette *model.Roulette
	)

	log = s.log.With(
		slog.String("op", op),
	)

	now := time.Now()

	roulette = &model.Roulette{
		ID:       rouletteID,
		PlayedAt: &now,
	}

	err = s.rouletteRep.UpdateRoulettePlayedAt(roulette)
	if err != nil {
		log.Error("failed to update roulette", sl.Err(err))

		return
	}

	log.Info("roulette updated", sl.Any("roulette", roulette))

	return
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
		"created_at": roulette.CreatedAt,
		"round":      roulette.Round,
	}

	message := event.Message{
		Channel: "roulette",
		Event:   "start",
		Data:    data,
	}

	return s.event.TriggerEvent(message)
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

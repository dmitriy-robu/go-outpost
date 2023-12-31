package start

import (
	"database/sql"
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
	transaction    repository.Transaction
}

func NewRouletteStart(
	log *slog.Logger,
	rouletteRep repository.RouletteRepository,
	rouletteBetRep repository.RouletteBetRepository,
	eventClient *event.PusherEvent,
	rouletteRoller *RouletteRoller,
	balance balance.Interface,
	transaction repository.Transaction) *RouletteStart {
	return &RouletteStart{
		log:            log,
		rouletteRep:    rouletteRep,
		rouletteBetRep: rouletteBetRep,
		cache:          cache.New(5*time.Minute, 10*time.Minute),
		event:          eventClient,
		rouletteRoller: rouletteRoller,
		balance:        balance,
		transaction:    transaction,
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
			tx                    *sql.Tx
			round                 int64
		)

		log = s.log.With(
			slog.String("op", op),
		)

		tx, err = s.transaction.StartTransaction()
		if err != nil {
			log.Error("failed to start transaction", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to start transaction", http.StatusInternalServerError))

			return
		}
		defer func() {
			if r := recover(); r != nil {
				if err = tx.Rollback(); err != nil {
					log.Error("failed to rollback transaction", sl.Err(err))
				}
			}
		}()

		round = s.getRoundFromCacheOrDB()

		roulette = &model.Roulette{
			UUID:  uuid.New(),
			Round: round,
		}

		rouletteID, err = s.rouletteRep.SaveRoulette(*roulette)
		if err != nil {
			log.Error("failed to save roulette", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to save roulette", http.StatusInternalServerError))

			if err = tx.Rollback(); err != nil {
				log.Error("failed to rollback transaction", sl.Err(err))
			}

			return
		}

		s.updateCacheRound(round + 1)

		log.Info("roulette created", sl.String("roulette_id", fmt.Sprintf("%d", rouletteID)))

		roulette, err = s.rouletteRep.GetRouletteByID(rouletteID)
		if err != nil {
			log.Error("failed to get roulette", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to get roulette", http.StatusInternalServerError))

			if err = tx.Rollback(); err != nil {
				log.Error("failed to rollback transaction", sl.Err(err))
			}

			return
		}

		/*autoBets := getAutoBetsFromCache()
		if len(autoBets) != 0 {
			job.Dispatch(&RouletteBetReplicateJob{rouletteID: rouletteID}, 0)
		}*/

		err = s.sendNewRoundEvent(roulette)
		if err != nil {
			log.Error("failed to send new round event", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to send new round event", http.StatusInternalServerError))

			if err = tx.Rollback(); err != nil {
				log.Error("failed to rollback transaction", sl.Err(err))
			}

			return
		}

		log.Info("new round event sent")

		winColorAndNumberData, err = s.rouletteRoller.Roll(roulette)
		if err != nil {
			log.Error("failed to roll roulette", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to roll roulette", http.StatusInternalServerError))

			if err = tx.Rollback(); err != nil {
				log.Error("failed to rollback transaction", sl.Err(err))
			}

			return
		}

		log.Info("roulette rolled",
			slog.Any("win_color", winColorAndNumberData.Color),
			slog.Any("win_number", winColorAndNumberData.Number))

		if err = s.handleWinners(rouletteID, winColorAndNumberData.Color); err != nil {
			log.Error("failed to handle winners", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to handle winners", http.StatusInternalServerError))

			if err = tx.Rollback(); err != nil {
				log.Error("failed to rollback transaction", sl.Err(err))
			}

			return
		}

		log.Info("winners handled")

		delay := 15 * time.Second

		eventMessage := event.Message{
			Channel: "roulette",
			Event:   "winner",
			Data: map[string]interface{}{
				"color":  winColorAndNumberData.Color,
				"number": winColorAndNumberData.Number,
			},
		}

		job.Dispatch(&job.SendEventJob{EventMessage: eventMessage, Event: s.event}, delay)

		job.Dispatch(&RouletteStartJob{RouletteStart: s, RouletteID: rouletteID}, delay)

		if err = tx.Commit(); err != nil {
			log.Error("failed to commit transaction", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to commit transaction", http.StatusInternalServerError))

			return
		}
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

	log.Info("roulette updated")

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
		return 1
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
		roulette   *model.Roulette
	)

	roulette, err = s.rouletteRep.GetPreviousRouletteID(rouletteID)
	if err != nil {
		s.log.Error("failed to get previous roulette id", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	s.log.Info("previous roulette", sl.Any("rouletteID", roulette.ID))

	bets, err = s.rouletteBetRep.GetBetsByRouletteIDAndColor(roulette.ID, color)
	if err != nil {
		s.log.Error("failed to get winners by roulette id", sl.Err(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	multiplier = s.getMultiplierByColor(color)

	s.log.Info("multiplier", sl.Any("multiplier", multiplier))

	if len(bets) == 0 {
		s.log.Info("No bets found")

		return nil
	}

	for _, winner := range bets {
		if err = s.balance.Income(winner.UserID, winner.Amount*multiplier, config.Roulette); err != nil {
			s.log.Error("failed to update user balance", sl.Err(err))

			return fmt.Errorf("%s: %w", op, err)
		}

		s.log.Info("user balance updated", sl.Any("user_id", winner.UserID))
	}

	return nil
}

func (s *RouletteStart) getMultiplierByColor(color config.Color) int {
	colorConfig, ok := config.RouletteWheelConfig.Colors[color]
	if !ok {
		return 0
	}

	return colorConfig.Multiplier
}

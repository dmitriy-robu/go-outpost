package start

import (
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	"go-outpost/internal/http-server/handlers/event"
	"go-outpost/internal/http-server/handlers/job"
	"go-outpost/internal/http-server/model"
	resp "go-outpost/internal/lib/api/response"
	"go-outpost/internal/repository"
	"golang.org/x/exp/slog"
	"net/http"
	"time"
)

type RouletteStart struct {
	log            *slog.Logger
	rouletteRep    repository.RouletteRepository
	cache          *cache.Cache
	pusher         *event.PusherEvent
	rouletteRoller *RouletteRoller
}

type Roulette interface {
	Roll(roulette *model.Roulette) (winColor string, winNumber int)
}

func NewRouletteStart(
	log *slog.Logger,
	rouletteRep repository.RouletteRepository,
	pusherClient *event.PusherEvent,
	rouletteRoller *RouletteRoller) *RouletteStart {
	return &RouletteStart{
		log:            log,
		rouletteRep:    rouletteRep,
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

		winColorAndNumberData, err = s.rouletteRoller.Roll(roulette)
		if err != nil {
			log.Error("failed to roll roulette")

			render.JSON(w, r, resp.Error("failed to roll roulette", http.StatusInternalServerError))

			return
		}

		job.Dispatch(&RouletteUpdatePlayedAtJob{rouletteID: rouletteID}, 15*time.Second)
		job.Dispatch(&RouletteWinnerJob{winColor: winColorAndNumberData.Color, winNumber: winColorAndNumberData.Number}, 15*time.Second)
	}
}

func (s *RouletteStart) getRoundFromCacheOrDB() int64 {
	/*	round, err := s.getRoundFromCache()
		if err != nil {
			round, err = s.rouletteRep.GetLastRound()
			if err != nil {
				return 0
			}

			s.updateCacheRound(round)
		}

		return round*/
	return 0
}

func (s *RouletteStart) getRoundFromCache() int64 {
	round, found := s.cache.Get("current_round")
	if found {
		return round.(int64)
	}

	return 0
}

func (s *RouletteStart) updateCacheRound(round int64) {
	s.cache.Set("current_round", round, 0)
}

func (s *RouletteStart) sendNewRoundEvent(roulette *model.Roulette) error {
	data := map[string]interface{}{
		"uuid":       roulette.UUID.String(),
		"created_at": roulette.CreatedAt.Format(time.RFC3339),
		"round":      roulette.Round,
	}

	return s.pusher.TriggerEvent("balance-channel", "outcome-event", data)
}

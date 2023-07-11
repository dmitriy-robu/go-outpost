package place_bet

import (
	"database/sql"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"go-outpost/internal/api/config"
	"go-outpost/internal/api/http-server/handlers/user/balance"
	"go-outpost/internal/api/http-server/model"
	"go-outpost/internal/api/repository"
	resp "go-outpost/internal/lib/api/response"
	"go-outpost/internal/lib/converter"
	"go-outpost/internal/lib/logger/sl"
	"golang.org/x/exp/slog"
	"net/http"
)

type Request struct {
	BetRequest []BetRequest `json:"bets" validate:"required,min=1,max=2"`
	UserUUID   string       `json:"user_uuid" validate:"required"`
}

type BetRequest struct {
	Color  config.Color `json:"color" validate:"required"`
	Amount float64      `json:"amount" validate:"required,min=0.01"`
}

type Response struct {
	resp.Response
}

type BetCounter interface {
	CountBetsByRouletteAndUser(rouletteID int64, userID int64) (int, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=BetSaver
type BetSaver interface {
	SaveBet(bet model.RouletteBet) (int64, error)
	BetCounter
}

type Bet struct {
	log         *slog.Logger
	validator   *validator.Validate
	rouletteRep repository.RouletteRepository
	betSaver    BetSaver
	userRep     repository.UserRepository
	balance     balance.Interface
	transaction repository.Transaction
}

func NewBet(
	log *slog.Logger,
	rouletteRep repository.RouletteRepository,
	betSaver BetSaver,
	userRep repository.UserRepository,
	balance balance.Interface,
	transaction repository.Transaction) *Bet {
	return &Bet{
		log:         log,
		validator:   validator.New(),
		rouletteRep: rouletteRep,
		betSaver:    betSaver,
		userRep:     userRep,
		balance:     balance,
		transaction: transaction,
	}
}

func (b *Bet) New() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.bet.save.New"

		var (
			err             error
			req             Request
			log             *slog.Logger
			uuidStr         string
			roulette        *model.Roulette
			user            *model.User
			betCount        int
			rouletteBet     model.RouletteBet
			id              int64
			convertedAmount int
			userBalance     *model.UserBalance
			tx              *sql.Tx
		)

		log = b.log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		tx, err = b.transaction.StartTransaction()
		if err != nil {
			log.Error("failed to start transaction", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to start transaction", http.StatusInternalServerError))

			return
		}
		defer func() {
			if r := recover(); r != nil {
				if err = b.transaction.RollbackTransaction(tx); err != nil {
					log.Error("failed to rollback transaction", sl.Err(err))

					return
				}
			}
		}()

		if err = render.DecodeJSON(r.Body, &req); err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to decode request body", http.StatusBadRequest))

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err = b.validator.Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			log.Error("invalid request", sl.Err(err))

			render.JSON(w, r, resp.ValidationError(validateErr))

			return
		}

		uuidStr = chi.URLParam(r, "uuid")

		roulette, err = b.rouletteRep.FindRouletteByUUID(uuidStr)
		if err != nil {
			log.Error("failed to find start", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to find start", http.StatusNotFound))

			return
		}

		log.Info("roulette found", slog.Any("roulette", roulette))

		user, err = b.userRep.FindUserByUUID(req.UserUUID)
		if err != nil {
			log.Error("failed to find user", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to find user", http.StatusNotFound))

			return
		}

		log.Info("user found", slog.Any("user", user))

		userBalance, err = b.userRep.FindUserBalanceByID(user.ID)
		if err != nil {
			log.Error("failed to find user balance", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to find user balance", http.StatusInternalServerError))

			return
		}

		log.Info("user balance found", slog.Any("user_balance", userBalance))

		if userBalance.Balance < 0 {
			log.Error("user has no balance", sl.Err(err))

			render.JSON(w, r, resp.Error("user has no balance", http.StatusNotFound))

			return
		}

		log.Info("user has balance", slog.Any("user_balance", userBalance))

		totalBetAmount := 0.0
		for _, bet := range req.BetRequest {
			totalBetAmount += bet.Amount
		}

		convertedAmount = converter.ConvertAmountFloatToInt(totalBetAmount)

		if userBalance.Balance < convertedAmount {
			log.Error("user has insufficient balance", sl.Err(err))

			render.JSON(w, r, resp.Error("user has insufficient balance", http.StatusNotFound))

			return
		}

		log.Info("user has sufficient balance", slog.Any("user_balance", userBalance))

		betCount, err = b.betSaver.CountBetsByRouletteAndUser(roulette.ID, user.ID)
		if err != nil {
			log.Error("failed to count bets", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to count bets", http.StatusInternalServerError))

			return
		}

		log.Info("bet count", slog.Any("bet_count", betCount))

		if betCount+len(req.BetRequest) > 2 {
			log.Info("user has already placed 2 bets on this start",
				slog.Any("user_id", user.ID),
				slog.Any("roulette_id", roulette.ID))

			render.JSON(w, r, resp.Error(
				"user is trying to place more than 2 bets on this start",
				http.StatusInternalServerError))

			return
		}

		log.Info("user has not placed 2 bets on this start")

		if err = b.balance.Outcome(user.ID, convertedAmount, config.Roulette); err != nil {
			log.Error("failed to update user balance", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to update user balance", http.StatusInternalServerError))

			if err = b.transaction.RollbackTransaction(tx); err != nil {
				log.Error("failed to rollback transaction", sl.Err(err))

				return
			}

			return
		}

		log.Info("user balance updated", slog.Any("user_balance", userBalance))

		for _, bet := range req.BetRequest {
			rouletteBet = model.RouletteBet{
				Color:      bet.Color,
				Amount:     converter.ConvertAmountFloatToInt(bet.Amount),
				RouletteID: roulette.ID,
				UserID:     user.ID,
			}

			id, err = b.betSaver.SaveBet(rouletteBet)
			if err != nil {
				log.Error("failed to save bet", sl.Err(err))

				render.JSON(w, r, resp.Error("failed to save bet", http.StatusInternalServerError))

				if err = b.transaction.RollbackTransaction(tx); err != nil {
					log.Error("failed to rollback transaction", sl.Err(err))

					return
				}

				return
			}

			log.Info("bet saved", slog.Any("id", id))
		}

		if err = b.transaction.CommitTransaction(tx); err != nil {
			log.Error("failed to commit transaction", sl.Err(err))

			render.JSON(w, r, resp.Error("failed to commit transaction", http.StatusInternalServerError))

			return
		}

		responseOK(w, r)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
	})
}

package balance

import (
	"fmt"
	"go-outpost/internal/config"
	"go-outpost/internal/http-server/handlers/event"
	"go-outpost/internal/http-server/model"
	"go-outpost/internal/lib/converter"
	"go-outpost/internal/repository"
	"golang.org/x/exp/slog"
	"strconv"
)

type Balance struct {
	userRep repository.UserRepository
	log     *slog.Logger
	pusher  *event.PusherEvent
}

type Interface interface {
	Income(userID int64, amount int, game config.Game) error
	Outcome(userID int64, amount int, game config.Game) error
}

func NewBalance(
	userRep repository.UserRepository,
	log *slog.Logger,
	pusherClient *event.PusherEvent) *Balance {
	return &Balance{
		userRep: userRep,
		log:     log,
		pusher:  pusherClient,
	}
}

func (b *Balance) Income(userID int64, amount int, game config.Game) error {
	const op = "handlers.user.balance.Income"

	var (
		err         error
		user        *model.User
		userBalance *model.UserBalance
		message     event.Message
	)

	if err = b.userRep.IncomeToUserBalance(userID, amount); err != nil {
		b.log.Error("failed to income to user balance")

		return fmt.Errorf("%s: %w", op, err)
	}

	b.log.Info("user balance updated")

	if err = b.userRep.CreateUserBalanceTransaction(userID, amount, config.Income, game); err != nil {
		b.log.Error("failed to create user balance transaction")

		return fmt.Errorf("%s: %w", op, err)
	}

	b.log.Info("user balance transaction created")

	user, err = b.userRep.GetUserByID(userID)
	if err != nil {
		b.log.Error("failed to find user by id")

		return fmt.Errorf("%s: %w", op, err)
	}

	b.log.Info("user found")

	userBalance, err = b.userRep.FindUserBalanceByID(user.ID)
	if err != nil {
		b.log.Error("failed to find user balance by id")

		return fmt.Errorf("%s: %w", op, err)
	}

	b.log.Info("user balance found")

	message = event.Message{
		Channel: "balance-channel",
		Event:   "income-event",
		Data: map[string]interface{}{
			"user_uuid":      user.UUID,
			"amount":         strconv.Itoa(amount),
			"operation_type": config.Income,
			"module":         config.Roulette,
			"balance":        converter.ConvertAmountIntToSting(userBalance.Balance),
		},
	}

	return b.pusher.TriggerEvent(message)

}

func (b *Balance) Outcome(userID int64, amount int, game config.Game) error {
	const op = "handlers.user.balance.Outcome"

	var (
		err         error
		user        *model.User
		userBalance *model.UserBalance
		message     event.Message
	)

	if err = b.userRep.OutcomeFromUserBalance(userID, amount); err != nil {
		b.log.Error("failed to outcome from user balance")

		return fmt.Errorf("%s: %w", op, err)
	}

	b.log.Info("user balance updated")

	if err = b.userRep.CreateUserBalanceTransaction(userID, amount, config.Outcome, game); err != nil {
		b.log.Error("failed to create user balance transaction")

		return fmt.Errorf("%s: %w", op, err)
	}

	b.log.Info("user balance transaction created")

	user, err = b.userRep.GetUserByID(userID)
	if err != nil {
		b.log.Error("failed to find user by id")

		return fmt.Errorf("%s: %w", op, err)
	}

	b.log.Info("user found")

	userBalance, err = b.userRep.FindUserBalanceByID(user.ID)
	if err != nil {
		b.log.Error("failed to find user balance by id")

		return fmt.Errorf("%s: %w", op, err)
	}

	b.log.Info("user balance found")

	message = event.Message{
		Channel: "balance-channel",
		Event:   "outcome-event",
		Data: map[string]interface{}{
			"user_uuid":      user.UUID,
			"amount":         strconv.Itoa(amount),
			"operation_type": config.Outcome,
			"module":         config.Roulette,
			"balance":        converter.ConvertAmountIntToSting(userBalance.Balance),
		},
	}

	return b.pusher.TriggerEvent(message)
}

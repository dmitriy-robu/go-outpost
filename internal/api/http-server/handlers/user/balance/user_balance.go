package balance

import (
	"fmt"
	config2 "go-outpost/internal/api/config"
	"go-outpost/internal/api/http-server/handlers/event"
	model2 "go-outpost/internal/api/http-server/model"
	"go-outpost/internal/api/repository"
	"go-outpost/internal/lib/converter"
	"golang.org/x/exp/slog"
	"strconv"
)

type Balance struct {
	userRep repository.UserRepository
	log     *slog.Logger
	pusher  *event.PusherEvent
}

type Interface interface {
	Income(userID int64, amount int, game config2.Game) error
	Outcome(userID int64, amount int, game config2.Game) error
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

func (b *Balance) Income(userID int64, amount int, game config2.Game) error {
	const op = "handlers.user.balance.Income"

	var (
		err         error
		user        *model2.User
		userBalance *model2.UserBalance
		message     event.Message
	)

	if err = b.userRep.IncomeToUserBalance(userID, amount); err != nil {
		b.log.Error("failed to income to user balance")

		return fmt.Errorf("%s: %w", op, err)
	}

	b.log.Info("user balance updated")

	if err = b.userRep.CreateUserBalanceTransaction(userID, amount, config2.Income, game); err != nil {
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
			"operation_type": config2.Income,
			"module":         config2.Roulette,
			"balance":        converter.ConvertAmountIntToSting(userBalance.Balance),
		},
	}

	return b.pusher.TriggerEvent(message)

}

func (b *Balance) Outcome(userID int64, amount int, game config2.Game) error {
	const op = "handlers.user.balance.Outcome"

	var (
		err         error
		user        *model2.User
		userBalance *model2.UserBalance
		message     event.Message
	)

	if err = b.userRep.OutcomeFromUserBalance(userID, amount); err != nil {
		b.log.Error("failed to outcome from user balance")

		return fmt.Errorf("%s: %w", op, err)
	}

	b.log.Info("user balance updated")

	if err = b.userRep.CreateUserBalanceTransaction(userID, amount, config2.Outcome, game); err != nil {
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
			"operation_type": config2.Outcome,
			"module":         config2.Roulette,
			"balance":        converter.ConvertAmountIntToSting(userBalance.Balance),
		},
	}

	return b.pusher.TriggerEvent(message)
}

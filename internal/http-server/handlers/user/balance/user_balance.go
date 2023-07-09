package balance

import (
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
	err := b.userRep.IncomeToUserBalance(userID, amount)
	if err != nil {
		b.log.Error("failed to income to user balance")

		return err
	}

	err = b.userRep.CreateUserBalanceTransaction(userID, amount, config.Income, game)
	if err != nil {
		b.log.Error("failed to create user balance transaction")

		return err
	}

	return nil
}

func (b *Balance) Outcome(userID int64, amount int, game config.Game) error {
	var (
		err  error
		user *model.User
	)

	if err = b.userRep.OutcomeFromUserBalance(userID, amount); err != nil {
		b.log.Error("failed to outcome from user balance")

		return err
	}

	if err = b.userRep.CreateUserBalanceTransaction(userID, amount, config.Outcome, game); err != nil {
		b.log.Error("failed to create user balance transaction")

		return err
	}

	user, err = b.userRep.GetUserByID(userID)

	userBalance, err := b.userRep.FindUserBalanceByID(user.ID)
	if err != nil {
		b.log.Error("failed to find user balance by id")

		return err
	}

	data := map[string]interface{}{
		"user_uuid":      user.UUID,
		"amount":         strconv.Itoa(amount),
		"operation_type": string(config.Outcome),
		"module":         string(config.Roulette),
		"balance":        converter.ConvertAmountIntToSting(userBalance.Balance),
	}

	return b.pusher.TriggerEvent("balance-channel", "outcome-event", data)
}

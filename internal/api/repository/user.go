package repository

import (
	"database/sql"
	"fmt"
	config "go-outpost/internal/api/config"
	"go-outpost/internal/api/http-server/handlers/mysql"
	model "go-outpost/internal/api/http-server/model"
	"time"
)

type UserRepository struct {
	dbhandler mysql.Handler
}

func NewUserRepository(dbhandler mysql.Handler) *UserRepository {
	return &UserRepository{dbhandler: dbhandler}
}

func (repo *UserRepository) FindUserByUUID(uuid string) (*model.User, error) {
	const query = "SELECT id FROM users WHERE uuid = ?"
	row, err := repo.dbhandler.PrepareAndQueryRow(query, uuid)
	if err != nil {
		return nil, err
	}

	user := &model.User{}

	err = row.Scan(&user.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return user, nil
}

func (repo *UserRepository) FindUserBalanceByID(userID int64) (*model.UserBalance, error) {
	const op = "repository.user.FindUserBalanceByID"

	const query = "SELECT balance FROM user_balances WHERE user_id = ?"
	row, err := repo.dbhandler.PrepareAndQueryRow(query, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	userBalance := &model.UserBalance{}

	err = row.Scan(&userBalance.Balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return userBalance, nil
}

func (repo *UserRepository) OutcomeFromUserBalance(userID int64, amount int) error {
	const op = "repository.user.OutcomeFromUserBalance"

	now := time.Now()

	const query = "UPDATE user_balances SET balance = balance - ?, updated_at = ? WHERE id = ?"
	_, err := repo.dbhandler.PrepareAndExecute(query, amount, now, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (repo *UserRepository) IncomeToUserBalance(userID int64, amount int) error {
	const op = "repository.user.IncomeToUserBalance"

	now := time.Now()

	const query = "UPDATE user_balances SET balance = balance + ?, updated_at = ? WHERE id = ?"
	_, err := repo.dbhandler.PrepareAndExecute(query, amount, now, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (repo *UserRepository) CreateUserBalanceTransaction(
	userID int64,
	amount int,
	balanceType config.BalanceType,
	game config.Game,
) error {
	const op = "repository.user.CreateUserBalanceTransaction"

	now := time.Now()

	const query = "INSERT INTO user_balance_transactions(user_id, value, type, module, created_at, updated_at) " +
		"VALUES(?, ?, ?, ?, ?, ?)"
	_, err := repo.dbhandler.PrepareAndExecute(query, userID, amount, balanceType, game, now, now)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (repo *UserRepository) GetUserByID(userID int64) (*model.User, error) {
	const op = "repository.user.GetUserByID"

	const query = "SELECT id, uuid FROM users WHERE id = ?"
	row, err := repo.dbhandler.PrepareAndQueryRow(query, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	user := &model.User{}

	err = row.Scan(&user.ID, &user.UUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (repo *UserRepository) StartTransaction() (*sql.Tx, error) {
	const op = "repository.user.StartTransaction"

	tx, err := repo.dbhandler.StartTransaction()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return tx, nil
}

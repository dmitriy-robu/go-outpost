package repository

import (
	"fmt"
	"go-outpost/internal/http-server/handlers/mysql"
	"go-outpost/internal/http-server/model"
	"time"
)

type BetRepository struct {
	dbhandler mysql.Handler
}

func NewBetRepository(dbhandler mysql.Handler) *BetRepository {
	return &BetRepository{dbhandler: dbhandler}
}

func (repo *BetRepository) SaveBet(bet model.RouletteBet) (int64, error) {
	const op = "repository.bet.SaveBet"

	now := time.Now()

	res, err := repo.dbhandler.PrepareAndExecute(
		"INSERT INTO roulette_bets(color, amount, roulette_id, user_id, created_at, updated_at) "+
			"VALUES(?, ?, ?, ?, ?, ?)",
		bet.Color, bet.Amount, bet.RouletteID, bet.UserID, now, now)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (repo *BetRepository) CountBetsByRouletteAndUser(rouletteID int64, userID int64) (int, error) {
	const op = "repository.bet.CountBetsByRouletteAndUser"

	const query = "SELECT COUNT(*) FROM roulette_bets WHERE roulette_id = ? AND user_id = ?"
	row, err := repo.dbhandler.PrepareAndQueryRow(query, rouletteID, userID)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	var count int

	err = row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return count, nil
}

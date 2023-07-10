package repository

import (
	"fmt"
	"go-outpost/internal/config"
	"go-outpost/internal/http-server/handlers/mysql"
	"go-outpost/internal/http-server/model"
	"time"
)

type RouletteBetRepository struct {
	dbhandler mysql.Handler
}

func NewBetRepository(dbhandler mysql.Handler) *RouletteBetRepository {
	return &RouletteBetRepository{dbhandler: dbhandler}
}

func (repo *RouletteBetRepository) SaveBet(bet model.RouletteBet) (int64, error) {
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

func (repo *RouletteBetRepository) CountBetsByRouletteAndUser(rouletteID int64, userID int64) (int, error) {
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

func (repo *RouletteBetRepository) GetBetsByRouletteIDAndColor(
	rouletteID int64,
	color config.Color,
) ([]model.RouletteBet, error) {
	const op = "repository.bet.GetBetsByRouletteIDAndColor"

	const query = "SELECT id, color, amount, roulette_id, user_id, created_at, updated_at " +
		"FROM roulette_bets WHERE roulette_id = ? AND color = ?"

	rows, err := repo.dbhandler.PrepareAndQuery(query, rouletteID, color)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	bets := make([]model.RouletteBet, 0)

	for rows.Next() {
		var bet model.RouletteBet

		err = rows.Scan(&bet.ID, &bet.Color, &bet.Amount, &bet.RouletteID, &bet.UserID, &bet.CreatedAt, &bet.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		bets = append(bets, bet)
	}

	return bets, nil
}

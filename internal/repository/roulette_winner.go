package repository

import (
	"fmt"
	"go-outpost/internal/config"
	"go-outpost/internal/http-server/handlers/mysql"
	"go-outpost/internal/http-server/model"
	"time"
)

type RouletteWinnerRepository struct {
	dbhandler mysql.Handler
}

func NewRouletteWinnerRepository(dbhandler mysql.Handler) *RouletteWinnerRepository {
	return &RouletteWinnerRepository{dbhandler: dbhandler}
}

func (repo *RouletteWinnerRepository) SaveWin(roulette *model.Roulette, color config.Color, number int) error {
	const op = "repository.roulette_winner.SaveWin"

	now := time.Now()

	_, err := repo.dbhandler.PrepareAndExecute(
		"INSERT INTO roulette_wins(color, roulette_id, number, created_at, updated_at) "+
			"VALUES(?, ?, ?, ?, ?)",
		color, roulette.ID, number, now, now)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (repo *RouletteWinnerRepository) GetNumberByColor(color config.Color) (int, error) {
	switch color {
	case config.Red:
		return 1, nil
	case config.Black:
		return 2, nil
	case config.Green:
		return 3, nil
	}

	return 0, fmt.Errorf("invalid color: %s", color)
}

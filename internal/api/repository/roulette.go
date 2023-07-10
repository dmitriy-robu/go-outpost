package repository

import (
	"database/sql"
	"fmt"
	"go-outpost/internal/api/http-server/handlers/mysql"
	"go-outpost/internal/api/http-server/model"
	"time"
)

type RouletteRepository struct {
	dbhandler mysql.Handler
}

func NewRouletteRepository(dbhandler mysql.Handler) *RouletteRepository {
	return &RouletteRepository{dbhandler: dbhandler}
}

func (repo *RouletteRepository) FindRouletteByUUID(uuid string) (*model.Roulette, error) {
	const op = "repository.roulette.FindRouletteByUUID"

	const query = "SELECT id,round,played_at FROM roulettes WHERE uuid = ?"

	row, err := repo.dbhandler.PrepareAndQueryRow(query, uuid)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	roulette := &model.Roulette{}

	err = row.Scan(&roulette.ID, &roulette.Round, &roulette.PlayedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return roulette, nil
}

func (repo *RouletteRepository) SaveRoulette(roulette model.Roulette) (int64, error) {
	const op = "repository.roulette.SaveRoulette"

	const query = "INSERT INTO roulettes(uuid, round, created_at, updated_at) VALUES(?, ?, ?, ?)"

	now := time.Now()

	res, err := repo.dbhandler.PrepareAndExecute(query, roulette.UUID, roulette.Round, now, now)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (repo *RouletteRepository) GetLastRound() (int64, error) {
	const op = "repository.roulette.GetLastRound"

	const query = "SELECT round FROM roulettes ORDER BY round DESC LIMIT 1"

	row, err := repo.dbhandler.PrepareAndQueryRow(query)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	var round int64

	err = row.Scan(&round)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return round, nil
}

func (repo *RouletteRepository) GetRouletteByID(id int64) (*model.Roulette, error) {
	const op = "repository.roulette.GetRouletteByID"

	const query = "SELECT id,uuid,round,played_at FROM roulettes WHERE id = ?"

	row, err := repo.dbhandler.PrepareAndQueryRow(query, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	roulette := &model.Roulette{}

	err = row.Scan(&roulette.ID, &roulette.UUID, &roulette.Round, &roulette.PlayedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return roulette, nil
}

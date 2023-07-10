package repository

import (
	"fmt"
	"go-outpost/internal/api/http-server/handlers/mysql"
	model "go-outpost/internal/api/http-server/model"
)

type ProvablyFairRepository struct {
	dbhandler mysql.Handler
}

func NewProvablyFairRepository(dbhandler mysql.Handler) *ProvablyFairRepository {
	return &ProvablyFairRepository{dbhandler: dbhandler}
}

func (repo *ProvablyFairRepository) SaveProvablyFair(provablyFair model.ProvablyFair) error {
	const op = "repository.provably_fair.SaveProvablyFair"

	const query = "INSERT INTO provably_fairs(game_draw_id," +
		" client_seed," +
		" server_seed," +
		" resulted_hash," +
		" resulted_random_number," +
		" min," +
		" max," +
		" nonce," +
		" created_at," +
		" updated_at) " +
		"VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	_, err := repo.dbhandler.PrepareAndExecute(query,
		provablyFair.GameDrawID,
		provablyFair.ClientSeed,
		provablyFair.ServerSeed,
		provablyFair.ResultedHash,
		provablyFair.ResultedRandomNumber,
		provablyFair.Min,
		provablyFair.Max,
		provablyFair.Nonce,
		provablyFair.CreatedAt,
		provablyFair.UpdatedAt)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (repo *ProvablyFairRepository) SaveGameDraw(gameDraw model.GameDraw) (int64, error) {
	const op = "repository.provably_fair.SaveGameDraw"

	const query = "INSERT INTO game_draws(game_id," +
		" game," +
		" created_at," +
		" updated_at) " +
		"VALUES(?, ?, ?, ?)"
	res, err := repo.dbhandler.PrepareAndExecute(query,
		gameDraw.GameID,
		gameDraw.Game,
		gameDraw.CreatedAt,
		gameDraw.UpdatedAt)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, _ := res.LastInsertId()

	return id, nil
}

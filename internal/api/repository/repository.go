package repository

import (
	"database/sql"
	"fmt"
	"go-outpost/internal/api/http-server/handlers/mysql"
)

type Repository struct {
	dbhandler mysql.Handler
}

func NewRepository(dbhandler mysql.Handler) *Repository {
	return &Repository{dbhandler: dbhandler}
}

func (repo *Repository) StartTransaction() (*sql.Tx, error) {
	const op = "repository.StartTransaction"

	tx, err := repo.dbhandler.StartTransaction()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return tx, nil
}

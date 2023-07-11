package repository

import (
	"database/sql"
	"fmt"
	"go-outpost/internal/api/http-server/handlers/mysql"
)

type Transaction struct {
	dbhandler mysql.Handler
}

func NewRepository(dbhandler mysql.Handler) *Transaction {
	return &Transaction{dbhandler: dbhandler}
}

func (tr *Transaction) StartTransaction() (*sql.Tx, error) {
	const op = "repository.StartTransaction"

	tx, err := tr.dbhandler.StartTransaction()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return tx, nil
}

func (tr *Transaction) RollbackTransaction(tx *sql.Tx) error {
	const op = "repository.RollbackTransaction"

	err := tx.Rollback()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (tr *Transaction) CommitTransaction(tx *sql.Tx) error {
	const op = "repository.CommitTransaction"

	err := tx.Commit()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

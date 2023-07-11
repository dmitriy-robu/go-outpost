package repository

import (
	"database/sql"
	"fmt"
	"go-outpost/internal/api/http-server/handlers/mysql"
)

type Transaction struct {
	dbhandler mysql.Handler
}

func NewTransaction(dbhandler mysql.Handler) *Transaction {
	return &Transaction{dbhandler: dbhandler}
}

func (tr *Transaction) StartTransaction() (*sql.Tx, error) {
	const op = "repository.transaction.StartTransaction"

	tx, err := tr.dbhandler.StartTransaction()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return tx, nil
}

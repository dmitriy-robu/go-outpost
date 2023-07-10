package mysql

import "database/sql"

type DB interface {
	Execute(statement string) (sql.Result, error)
	Query(statement string) (*sql.Rows, error)
}

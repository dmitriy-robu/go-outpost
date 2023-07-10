package mysql

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

type Handler struct {
	Conn *sql.DB
}

func New(conn *sql.DB) *Handler {
	return &Handler{Conn: conn}
}

func (handler *Handler) Execute(statement *sql.Stmt, args ...interface{}) (sql.Result, error) {
	const op = "mysql.mysql.Execute"

	result, err := statement.Exec(args)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return result, nil
}

func (handler *Handler) Query(statement string) (*sql.Rows, error) {
	const op = "mysql.mysql.Query"

	rows, err := handler.Conn.Query(statement)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return rows, nil
}

func (handler *Handler) Prepare(statement string) (*sql.Stmt, error) {
	const op = "mysql.mysql.Prepare"

	stmt, err := handler.Conn.Prepare(statement)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return stmt, nil
}

func (handler *Handler) PrepareAndExecute(statement string, args ...interface{}) (sql.Result, error) {
	const op = "mysql.mysql.PrepareAndExecute"

	stmt, err := handler.Conn.Prepare(statement)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	result, err := stmt.Exec(args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return result, nil
}

func (handler *Handler) PrepareAndQueryRow(statement string, args ...interface{}) (*sql.Row, error) {
	const op = "mysql.mysql.PrepareAndQueryRow"

	stmt, err := handler.Conn.Prepare(statement)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRow(args...)

	return row, nil
}

func (handler *Handler) PrepareAndQuery(statement string, args ...interface{}) (*sql.Rows, error) {
	const op = "mysql.mysql.PrepareAndQueryRow"

	stmt, err := handler.Conn.Prepare(statement)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	row, err := stmt.Query(args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return row, nil
}

func (handler *Handler) StartTransaction() (*sql.Tx, error) {
	const op = "mysql.mysql.StartTransaction"

	tx, err := handler.Conn.Begin()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return tx, nil
}

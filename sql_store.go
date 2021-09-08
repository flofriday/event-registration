package eventregistration

import (
	"database/sql"
	"errors"
)

type SqlStore struct {
	db            *sql.DB
	insertStmt    *sql.Stmt
	queryUuidStmt *sql.Stmt
	queryAllStmt  *sql.Stmt
	statisticStmt *sql.Stmt
}

func NewSqlStore(filename string) (*SqlStore, error) {
	return nil, errors.New("Not implemented")
}

func (s *SqlStore) add(user User) (*User, error) {
	return nil, errors.New("Not implemented")
}

func (s *SqlStore) getByUuid(uuid string) (*User, error) {
	return nil, errors.New("Not implemented")
}

func (s *SqlStore) getAll(uuid string) (*User, error) {
	return nil, errors.New("Not implemented")
}

func (s *SqlStore) getStatistics(uuid string) (*int, error) {
	return nil, errors.New("Not implemented")
}

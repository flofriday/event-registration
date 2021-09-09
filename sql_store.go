package main

import (
	"database/sql"
	_ "embed"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed create.sql
var createQuery string

type SqlStore struct {
	db            *sql.DB
	insertStmt    *sql.Stmt
	queryUuidStmt *sql.Stmt
	queryAllStmt  *sql.Stmt
}

func NewSqlStore(filename string) (*SqlStore, error) {

	// Setup the sqlite database
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(createQuery)
	if err != nil {
		return nil, err
	}

	// Setup the statements
	insertQuery := "INSERT INTO user (uuid, firstname, lastname, email, phone, createdat) VALUES( ?, ?, ?, ?, ?, ?)"
	insertStmt, err := db.Prepare(insertQuery)
	if err != nil {
		return nil, err
	}

	byUuidQuery := "SELECT uuid, firstname, lastname, email, phone, createdat FROM user WHERE uuid=?"
	queryUuidStmt, err := db.Prepare(byUuidQuery)
	if err != nil {
		return nil, err
	}

	getAllQuery := "SELECT uuid, firstname, lastname, email, phone, createdat FROM user"
	queryAllStmt, err := db.Prepare(getAllQuery)
	if err != nil {
		return nil, err
	}

	// Assemble the store
	store := SqlStore{
		db:            db,
		insertStmt:    insertStmt,
		queryUuidStmt: queryUuidStmt,
		queryAllStmt:  queryAllStmt,
	}
	return &store, nil
}

func (s *SqlStore) add(user User) error {
	_, err := s.insertStmt.Exec(user.UUID, user.FirstName, user.LastName, user.Email, user.Phone, user.CreatedAt.UnixNano())
	return err
}

func (s *SqlStore) getByUuid(uuid string) (*User, error) {
	var user User
	row := s.queryUuidStmt.QueryRow(uuid)

	var nanoseconds int64
	err := row.Scan(&user.UUID, &user.FirstName, &user.LastName, &user.Email, &user.Phone, &nanoseconds)
	if err != nil {
		return nil, err
	}

	user.CreatedAt = time.Unix(0, nanoseconds)
	return &user, nil
}

func (s *SqlStore) getAll() ([]User, error) {
	rows, err := s.queryAllStmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]User, 0)
	for rows.Next() {
		var user User
		var nanoseconds int64
		err := rows.Scan(&user.UUID, &user.FirstName, &user.LastName, &user.Email, &user.Phone, &nanoseconds)
		if err != nil {
			return nil, err
		}
		user.CreatedAt = time.Unix(0, nanoseconds)
		users = append(users, user)
	}

	return users, nil
}

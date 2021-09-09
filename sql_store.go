package main

import (
	"database/sql"
	_ "embed"
	"errors"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed create.sql
var createQuery string

type SqlStore struct {
	db             *sql.DB
	insertStmt     *sql.Stmt
	queryUuidStmt  *sql.Stmt
	deleteUuidStmt *sql.Stmt
	queryAllStmt   *sql.Stmt
	queryLastStmt  *sql.Stmt
	queryCountStmt *sql.Stmt
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

	byUuidDelete := "DELETE FROM user WHERE uuid=?"
	deleteUuidStmt, err := db.Prepare(byUuidDelete)
	if err != nil {
		return nil, err
	}

	getLastQuery := "SELECT uuid, firstname, lastname, email, phone, createdat FROM user ORDER BY createdat DESC LIMIT ?"
	queryLastStmt, err := db.Prepare(getLastQuery)
	if err != nil {
		return nil, err
	}

	countQuery := "SELECT COUNT(*) FROM user"
	queryCountStmt, err := db.Prepare(countQuery)
	if err != nil {
		return nil, err
	}

	// Assemble the store
	store := SqlStore{
		db:             db,
		insertStmt:     insertStmt,
		queryUuidStmt:  queryUuidStmt,
		deleteUuidStmt: deleteUuidStmt,
		queryAllStmt:   queryAllStmt,
		queryLastStmt:  queryLastStmt,
		queryCountStmt: queryCountStmt,
	}
	return &store, nil
}

func (s *SqlStore) Add(user User) error {
	_, err := s.insertStmt.Exec(user.UUID, user.FirstName, user.LastName, user.Email, user.Phone, user.CreatedAt.UnixNano())
	return err
}

func (s *SqlStore) GetByUuid(uuid string) (*User, error) {
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

func (s *SqlStore) DeleteByUuid(uuid string) error {
	res, err := s.deleteUuidStmt.Exec(uuid)
	if err != nil {
		return err
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected != 1 {
		return errors.New("Nothing deleted, maybe user doesn't exist anymore?")
	}

	return nil
}

func (s *SqlStore) GetAll() ([]User, error) {
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

func (s *SqlStore) GetLastN(n int) ([]User, error) {
	rows, err := s.queryLastStmt.Query(n)
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

func (s *SqlStore) Count() (int, error) {
	var count int
	row := s.queryCountStmt.QueryRow()
	err := row.Scan(&count)
	return count, err
}

package db

import (
	"database/sql"
	"hpdipto/flowterm/utils"

	_ "modernc.org/sqlite"
)

type DB struct {
	conn *sql.DB
}

func New(conn *sql.DB) *DB {
	return &DB{conn: conn}
}

func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		utils.LogError(db.Close())
		return nil, err
	}

	return db, nil
}

package db

import (
	"database/sql"
	"github.com/hpdipto/flowterm/utils"
	"strings"

	_ "modernc.org/sqlite"
)

type DB struct {
	conn *sql.DB
}

func New(conn *sql.DB) *DB {
	return &DB{conn: conn}
}

func dsn(path string) string {
	if strings.Contains(path, "?") {
		return path
	}

	if !strings.HasPrefix(path, "file:") {
		path = "file:" + path
	}

	return path + "?_pragma=foreign_keys(1)"
}

func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dsn(path))
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		utils.LogError(db.Close())
		return nil, err
	}

	return db, nil
}

package db

import (
	"database/sql"
	"embed"
	"io"
	"log"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrations embed.FS

func Migrate(conn *sql.DB) error {
	goose.SetBaseFS(migrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}

	goose.SetLogger(log.New(io.Discard, "", 0))

	return goose.Up(conn, "migrations")
}

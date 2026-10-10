package main

import (
	"hpdipto/flowterm/db"
	"hpdipto/flowterm/tui"
	"hpdipto/flowterm/utils"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	dbName := os.Getenv("DB_NAME")
	conn, err := db.Open(dbName)
	if err != nil {
		panic(err)
	}
	defer utils.LogDeferredError(conn.Close)

	database := db.New(conn)
	utils.PanicOnError(db.Migrate(conn))

	tui.InitTUi(database)
}

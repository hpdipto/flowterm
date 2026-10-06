package main

import (
	"hpdipto/flowterm/db"
	"hpdipto/flowterm/utils"
	"os"

	"github.com/joho/godotenv"
)

var database *db.DB

type config struct {
	focusSessionName string
	focusMinute      int
	breakSessionName string
	breakMinute      int
}

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

	database = db.New(conn)

	c := config{
		focusSessionName: "Focus",
		focusMinute:      25,
		breakSessionName: "Break",
		breakMinute:      5,
	}
	session, err := createSessionFromConfig(database, c)
	if err != nil {
		panic(err)
	}

	InitTUi(session.ID, c)
}

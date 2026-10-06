package main

import (
	"hpdipto/flowterm/db"
	"os"

	"github.com/joho/godotenv"
)

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
	defer conn.Close()

	database := db.New(conn)

	c := config{
		focusSessionName: "Focus",
		focusMinute:      25,
		breakSessionName: "Break",
		breakMinute:      5,
	}
	if err := createSessionFromConfig(database, c); err != nil {
		panic(err)
	}

	InitTUi(c)
}

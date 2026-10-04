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

	DB_NAME := os.Getenv("DB_NAME")
	db, err := db.Open(DB_NAME)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	c := config{
		focusSessionName: "Focus",
		focusMinute:      25,
		breakSessionName: "Break",
		breakMinute:      5,
	}
	InitTUi(c)
}

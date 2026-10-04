package main

import "hpdipto/flowterm/db"

type config struct {
	focusSessionName string
	focusMinute      int
	breakSessionName string
	breakMinute      int
}

func main() {
	db, err := db.Open("flowterm.db")
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

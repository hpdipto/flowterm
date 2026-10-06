package main

import "hpdipto/flowterm/db"

func createSessionFromConfig(database *db.DB, c config) error {
	session := &db.Session{
		FocusSessionName:     c.focusSessionName,
		BreakSessionName:     c.breakSessionName,
		FocusSessionDuration: c.focusMinute,
		BreakSessionDuration: c.breakMinute,
		Interrupted:          false,
	}

	return database.CreateSession(session)
}

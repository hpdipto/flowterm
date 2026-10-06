package main

import "hpdipto/flowterm/db"

func createSessionFromConfig(database *db.DB, c config) (*db.Session, error) {
	session := &db.Session{
		FocusSessionName:     c.focusSessionName,
		BreakSessionName:     c.breakSessionName,
		FocusSessionDuration: c.focusMinute,
		BreakSessionDuration: c.breakMinute,
		Interrupted:          false,
	}

	if err := database.CreateSession(session); err != nil {
		return nil, err
	}
	return session, nil
}

func closeSession(database *db.DB, sessionID int64, interrupted bool) error {
	return database.UpdateSession(sessionID, db.SessionUpdate{
		Interrupted: &interrupted,
	})
}

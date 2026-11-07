package db

import "github.com/hpdipto/flowterm/config"

func CreateSessionFromConfig(database *DB, c config.Config) (*Session, error) {
	session := &Session{
		FocusSessionName:     c.FocusSessionName,
		BreakSessionName:     c.BreakSessionName,
		FocusSessionDuration: c.FocusMinute,
		BreakSessionDuration: c.BreakMinute,
		Interrupted:          false,
	}

	if err := database.CreateSession(session); err != nil {
		return nil, err
	}

	for _, tagName := range c.Tags {
		tag, err := database.GetOrCreateTag(tagName)
		if err != nil {
			return nil, err
		}
		if err := database.AddTagToSession(session.ID, tag.ID); err != nil {
			return nil, err
		}
	}

	return session, nil
}

func CloseSession(database *DB, sessionID int64, interrupted bool) error {
	return database.UpdateSession(sessionID, SessionUpdate{
		Interrupted: &interrupted,
	})
}

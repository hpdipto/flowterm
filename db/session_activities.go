package db

import (
	"time"

	"hpdipto/flowterm/utils"
)

func (d *DB) CreateSessionActivity(a *SessionActivity) error {
	now := time.Now()

	result, err := d.conn.Exec(`
		INSERT INTO session_activities (
			session_id,
			app,
			started_at,
			ended_at,
			created_at,
			updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		a.SessionID,
		a.App,
		a.StartedAt,
		a.EndedAt,
		now,
		now,
	)
	if err != nil {
		return err
	}

	a.ID, err = result.LastInsertId()
	if err != nil {
		return err
	}

	a.CreatedAt = now
	a.UpdatedAt = now

	return nil
}

func (d *DB) ListSessionActivities(sessionID int64) ([]SessionActivity, error) {
	rows, err := d.conn.Query(`
		SELECT
			id,
			session_id,
			app,
			started_at,
			ended_at,
			created_at,
			updated_at
		FROM session_activities
		WHERE session_id = ?
		ORDER BY started_at
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer utils.LogDeferredError(rows.Close)

	var activities []SessionActivity

	for rows.Next() {
		var a SessionActivity

		if err := rows.Scan(
			&a.ID,
			&a.SessionID,
			&a.App,
			&a.StartedAt,
			&a.EndedAt,
			&a.CreatedAt,
			&a.UpdatedAt,
		); err != nil {
			return nil, err
		}

		activities = append(activities, a)
	}

	return activities, rows.Err()
}

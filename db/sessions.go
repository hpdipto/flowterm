package db

import (
	"fmt"
	"hpdipto/flowterm/utils"
	"strings"
	"time"
)

type SessionUpdate struct {
	FocusSessionName     *string
	BreakSessionName     *string
	FocusSessionDuration *int
	BreakSessionDuration *int
	Interrupted          *bool
}

func (d *DB) CreateSession(s *Session) error {
	now := time.Now()

	result, err := d.conn.Exec(`
		INSERT INTO sessions (
			focus_session_name,
			break_session_name,
			focus_session_duration,
			break_session_duration,
			interrupted,
			created_at,
			updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		s.FocusSessionName,
		s.BreakSessionName,
		s.FocusSessionDuration,
		s.BreakSessionDuration,
		s.Interrupted,
		Timestamp(now),
		Timestamp(now),
	)
	if err != nil {
		return err
	}

	s.ID, err = result.LastInsertId()
	if err != nil {
		return err
	}

	s.CreatedAt = now
	s.UpdatedAt = now

	return nil
}

func (d *DB) GetSession(id int64) (*Session, error) {
	var s Session

	err := d.conn.QueryRow(`
		SELECT
			id,
			focus_session_name,
			break_session_name,
			focus_session_duration,
			break_session_duration,
			interrupted,
			created_at,
			updated_at
		FROM sessions
		WHERE id = ?
	`, id).Scan(
		&s.ID,
		&s.FocusSessionName,
		&s.BreakSessionName,
		&s.FocusSessionDuration,
		&s.BreakSessionDuration,
		&s.Interrupted,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (d *DB) UpdateSession(id int64, update SessionUpdate) error {
	set := []string{}
	args := []any{}

	if update.FocusSessionName != nil {
		set = append(set, "focus_session_name = ?")
		args = append(args, *update.FocusSessionName)
	}

	if update.BreakSessionName != nil {
		set = append(set, "break_session_name = ?")
		args = append(args, *update.BreakSessionName)
	}

	if update.FocusSessionDuration != nil {
		set = append(set, "focus_session_duration = ?")
		args = append(args, *update.FocusSessionDuration)
	}

	if update.BreakSessionDuration != nil {
		set = append(set, "break_session_duration = ?")
		args = append(args, *update.BreakSessionDuration)
	}

	if update.Interrupted != nil {
		set = append(set, "interrupted = ?")
		args = append(args, *update.Interrupted)
	}

	if len(set) == 0 {
		return nil
	}

	set = append(set, "updated_at = ?")
	args = append(args, Timestamp(time.Now()))

	args = append(args, id)

	query := fmt.Sprintf(
		"UPDATE sessions SET %s WHERE id = ?",
		strings.Join(set, ", "),
	)

	_, err := d.conn.Exec(query, args...)
	return err
}

func (d *DB) DeleteSession(id int64) error {
	_, err := d.conn.Exec(`
		DELETE FROM sessions
		WHERE id = ?
	`, id)

	return err
}

func (d *DB) ListSessions() ([]Session, error) {
	rows, err := d.conn.Query(`
		SELECT
			id,
			focus_session_name,
			break_session_name,
			focus_session_duration,
			break_session_duration,
			interrupted,
			created_at,
			updated_at
		FROM sessions
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer utils.LogDeferredError(rows.Close)

	var sessions []Session

	for rows.Next() {
		var s Session

		if err := rows.Scan(
			&s.ID,
			&s.FocusSessionName,
			&s.BreakSessionName,
			&s.FocusSessionDuration,
			&s.BreakSessionDuration,
			&s.Interrupted,
			&s.CreatedAt,
			&s.UpdatedAt,
		); err != nil {
			return nil, err
		}

		sessions = append(sessions, s)
	}

	return sessions, rows.Err()
}

func (d *DB) MarkAbandonedSessions() error {
	_, err := d.conn.Exec(`
      UPDATE sessions
      SET interrupted = 1, updated_at = ?
      WHERE interrupted = 0 AND updated_at = created_at
   `,
		Timestamp(time.Now()),
	)

	return err
}

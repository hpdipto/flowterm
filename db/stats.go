package db

import (
	"hpdipto/flowterm/utils"
	"time"
)

type SessionStat struct {
	FocusSessionName     string
	FocusSessionDuration int
	BreakSessionName     string
	BreakSessionDuration int
	Interrupted          bool
	Tags                 string
}

func (d *DB) ListSessionStats(since time.Time, all bool) ([]SessionStat, error) {
	query := `
		SELECT
			s.focus_session_name,
			s.focus_session_duration,
			s.break_session_name,
			s.break_session_duration,
			s.interrupted,
			COALESCE(group_concat(t.name, ','), '')
		FROM sessions s
		LEFT JOIN session_tags st ON st.session_id = s.id
		LEFT JOIN tags t ON t.id = st.tag_id
		WHERE s.created_at >= ?
	`

	args := []any{since}

	if !all {
		query += " AND s.interrupted = 0"
	}

	query += `
		GROUP BY s.id
		ORDER BY s.created_at DESC
	`

	rows, err := d.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer utils.LogDeferredError(rows.Close)

	var stats []SessionStat
	for rows.Next() {
		var s SessionStat

		if err := rows.Scan(
			&s.FocusSessionName,
			&s.FocusSessionDuration,
			&s.BreakSessionName,
			&s.BreakSessionDuration,
			&s.Interrupted,
			&s.Tags,
		); err != nil {
			return nil, err
		}

		stats = append(stats, s)
	}

	return stats, rows.Err()
}

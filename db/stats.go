package db

import (
	"hpdipto/flowterm/utils"
	"sort"
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

type ActivityStat struct {
	App      string
	Duration time.Duration
}

func (d *DB) ListActivityStats(since time.Time, all bool) ([]ActivityStat, error) {
	query := `
		SELECT a.app, a.started_at, a.ended_at
		FROM session_activities a
		JOIN sessions s ON s.id = a.session_id
		WHERE s.created_at >= ?
	`

	args := []any{since}

	if !all {
		query += " AND s.interrupted = 0"
	}

	rows, err := d.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer utils.LogDeferredError(rows.Close)

	totals := map[string]time.Duration{}

	for rows.Next() {
		var app string
		var start, end time.Time

		if err := rows.Scan(&app, &start, &end); err != nil {
			return nil, err
		}

		totals[app] += end.Sub(start)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	stats := make([]ActivityStat, 0, len(totals))
	for app, duration := range totals {
		stats = append(stats, ActivityStat{App: app, Duration: duration})
	}

	sort.Slice(stats, func(i, j int) bool { return stats[i].Duration > stats[j].Duration })

	return stats, nil
}

package db

import (
	"hpdipto/flowterm/utils"
	"sort"
	"time"
)

type SessionStat struct {
	ID                   int64
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
			s.id,
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
			&s.ID,
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
	Name     string
	Duration time.Duration
}

type ActivityStats struct {
	Focus []ActivityStat
	Break []ActivityStat
}

func (d *DB) ListSessionActivityStats(since time.Time, all bool) (map[int64]ActivityStats, error) {
	query := `
		SELECT a.session_id, a.app, a.app_name, a.phase, a.started_at, a.ended_at
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

	totals := map[int64]map[Phase]map[string]time.Duration{}
	names := map[int64]map[Phase]map[string]string{}

	for rows.Next() {
		var sessionID int64
		var app, name string
		var phase Phase
		var start, end time.Time

		if err := rows.Scan(&sessionID, &app, &name, &phase, &start, &end); err != nil {
			return nil, err
		}

		if phase != PhaseBreak {
			phase = PhaseFocus
		}

		if totals[sessionID] == nil {
			totals[sessionID] = map[Phase]map[string]time.Duration{}
			names[sessionID] = map[Phase]map[string]string{}
		}

		if totals[sessionID][phase] == nil {
			totals[sessionID][phase] = map[string]time.Duration{}
			names[sessionID][phase] = map[string]string{}
		}

		totals[sessionID][phase][app] += end.Sub(start)

		if name != "" && names[sessionID][phase][app] == "" {
			names[sessionID][phase][app] = name
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := make(map[int64]ActivityStats, len(totals))
	for sessionID, phases := range totals {
		result[sessionID] = ActivityStats{
			Focus: buildActivityStats(phases[PhaseFocus], names[sessionID][PhaseFocus]),
			Break: buildActivityStats(phases[PhaseBreak], names[sessionID][PhaseBreak]),
		}
	}

	return result, nil
}

func buildActivityStats(totals map[string]time.Duration, names map[string]string) []ActivityStat {
	stats := make([]ActivityStat, 0, len(totals))

	for app, duration := range totals {
		name := names[app]
		if name == "" {
			name = app
		}

		stats = append(stats, ActivityStat{App: app, Name: name, Duration: duration})
	}

	sort.Slice(stats, func(i, j int) bool { return stats[i].Duration > stats[j].Duration })

	return stats
}

package db_test

import (
	"hpdipto/flowterm/db"
	"hpdipto/flowterm/internal/testsupport"
	"testing"
	"time"
)

func TestTimestampFormat(t *testing.T) {
	value := time.Date(2026, 10, 7, 0, 13, 37, 182038000, time.FixedZone("", 6*60*60))

	if got := db.Timestamp(value); got != "2026-10-07T00:13:37.182038+06:00" {
		t.Errorf("Timestamp = %q", got)
	}
}

func TestStoredTimestampsAreRFC3339(t *testing.T) {
	database, conn := testsupport.OpenDB(t)

	session := &db.Session{
		FocusSessionName:     "focus",
		BreakSessionName:     "break",
		FocusSessionDuration: 25,
		BreakSessionDuration: 5,
	}
	if err := database.CreateSession(session); err != nil {
		t.Fatalf("create session: %v", err)
	}

	tag, err := database.GetOrCreateTag("alpha")
	if err != nil {
		t.Fatalf("create tag: %v", err)
	}

	if err := database.AddTagToSession(session.ID, tag.ID); err != nil {
		t.Fatalf("add tag: %v", err)
	}

	activity := &db.SessionActivity{
		SessionID: session.ID,
		App:       "com.apple.Safari",
		AppName:   "Safari",
		Phase:     db.PhaseFocus,
		StartedAt: session.CreatedAt,
		EndedAt:   session.CreatedAt.Add(time.Minute),
	}
	if err := database.CreateSessionActivity(activity); err != nil {
		t.Fatalf("create activity: %v", err)
	}

	values := map[string]string{}

	for name, query := range map[string]string{
		"sessions.created_at":           "SELECT CAST(created_at AS TEXT) FROM sessions WHERE id = ?",
		"sessions.updated_at":           "SELECT CAST(updated_at AS TEXT) FROM sessions WHERE id = ?",
		"tags.created_at":               "SELECT CAST(created_at AS TEXT) FROM tags WHERE id = ?",
		"tags.updated_at":               "SELECT CAST(updated_at AS TEXT) FROM tags WHERE id = ?",
		"session_tags.created_at":       "SELECT CAST(created_at AS TEXT) FROM session_tags WHERE session_id = ?",
		"session_activities.started_at": "SELECT CAST(started_at AS TEXT) FROM session_activities WHERE id = ?",
		"session_activities.ended_at":   "SELECT CAST(ended_at AS TEXT) FROM session_activities WHERE id = ?",
	} {
		var id int64 = session.ID
		if name == "tags.created_at" || name == "tags.updated_at" {
			id = tag.ID
		} else if name == "session_activities.started_at" || name == "session_activities.ended_at" {
			id = activity.ID
		}

		var value string
		if err := conn.QueryRow(query, id).Scan(&value); err != nil {
			t.Fatalf("read %s: %v", name, err)
		}

		values[name] = value
	}

	for name, value := range values {
		if _, err := time.Parse(time.RFC3339Nano, value); err != nil {
			t.Errorf("%s = %q, want RFC 3339", name, value)
		}
	}
}

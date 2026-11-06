package db_test

import (
	"database/sql"
	"errors"
	"hpdipto/flowterm/db"
	"hpdipto/flowterm/internal/testsupport"
	"hpdipto/flowterm/utils"
	"path/filepath"
	"testing"
	"time"
)

func TestCreateSession(t *testing.T) {
	database, _ := testsupport.OpenDB(t)

	s := &db.Session{
		FocusSessionName:     "focus",
		BreakSessionName:     "break",
		FocusSessionDuration: 25,
		BreakSessionDuration: 5,
		Interrupted:          true,
	}

	before := time.Now()
	if err := database.CreateSession(s); err != nil {
		t.Fatalf("create session: %v", err)
	}
	after := time.Now()

	if s.ID == 0 {
		t.Error("ID = 0, want assigned id")
	}
	if s.CreatedAt.Before(before) || s.CreatedAt.After(after) {
		t.Errorf("CreatedAt = %v, want between %v and %v", s.CreatedAt, before, after)
	}
	if !s.CreatedAt.Equal(s.UpdatedAt) {
		t.Errorf("UpdatedAt = %v, want CreatedAt = %v", s.UpdatedAt, s.CreatedAt)
	}

	got, err := database.GetSession(s.ID)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}

	if got.FocusSessionName != s.FocusSessionName {
		t.Errorf("FocusSessionName = %q, want %q", got.FocusSessionName, s.FocusSessionName)
	}
	if got.BreakSessionName != s.BreakSessionName {
		t.Errorf("BreakSessionName = %q, want %q", got.BreakSessionName, s.BreakSessionName)
	}
	if got.FocusSessionDuration != s.FocusSessionDuration {
		t.Errorf("FocusSessionDuration = %d, want %d", got.FocusSessionDuration, s.FocusSessionDuration)
	}
	if got.BreakSessionDuration != s.BreakSessionDuration {
		t.Errorf("BreakSessionDuration = %d, want %d", got.BreakSessionDuration, s.BreakSessionDuration)
	}
	if got.Interrupted != s.Interrupted {
		t.Errorf("Interrupted = %t, want %t", got.Interrupted, s.Interrupted)
	}

	// Timestamps survive the round trip; compare with second precision since
	// DATETIME storage drops sub-second detail.
	if got.CreatedAt.Unix() != s.CreatedAt.Unix() {
		t.Errorf("stored CreatedAt = %v, want %v", got.CreatedAt, s.CreatedAt)
	}
	if got.UpdatedAt.Unix() != s.UpdatedAt.Unix() {
		t.Errorf("stored UpdatedAt = %v, want %v", got.UpdatedAt, s.UpdatedAt)
	}
}

func TestCreateSessionMultipleSessionsGetDistinctIDs(t *testing.T) {
	database, _ := testsupport.OpenDB(t)

	first := &db.Session{FocusSessionName: "focus"}
	second := &db.Session{FocusSessionName: "focus"}

	if err := database.CreateSession(first); err != nil {
		t.Fatalf("create first session: %v", err)
	}
	if err := database.CreateSession(second); err != nil {
		t.Fatalf("create second session: %v", err)
	}

	if first.ID == second.ID {
		t.Errorf("both sessions got ID %d, want distinct ids", first.ID)
	}
}

func TestGetSessionReturnsOnlyMatchingRow(t *testing.T) {
	database, _ := testsupport.OpenDB(t)

	first := &db.Session{FocusSessionName: "first", FocusSessionDuration: 10}
	second := &db.Session{FocusSessionName: "second", FocusSessionDuration: 20}
	for _, s := range []*db.Session{first, second} {
		if err := database.CreateSession(s); err != nil {
			t.Fatalf("create session: %v", err)
		}
	}

	got, err := database.GetSession(second.ID)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}

	if got.ID != second.ID {
		t.Errorf("ID = %d, want %d", got.ID, second.ID)
	}
	if got.FocusSessionName != second.FocusSessionName {
		t.Errorf("FocusSessionName = %q, want %q", got.FocusSessionName, second.FocusSessionName)
	}
	if got.FocusSessionDuration != second.FocusSessionDuration {
		t.Errorf("FocusSessionDuration = %d, want %d", got.FocusSessionDuration, second.FocusSessionDuration)
	}
}

func TestGetSessionMissingID(t *testing.T) {
	database, _ := testsupport.OpenDB(t)

	s := &db.Session{FocusSessionName: "focus"}
	if err := database.CreateSession(s); err != nil {
		t.Fatalf("create session: %v", err)
	}

	got, err := database.GetSession(s.ID + 1)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("err = %v, want %v", err, sql.ErrNoRows)
	}
	if got != nil {
		t.Errorf("session = %+v, want nil", got)
	}
}

func TestUpdateSessionPartialUpdateLeavesOtherFields(t *testing.T) {
	database, _ := testsupport.OpenDB(t)

	s := &db.Session{
		FocusSessionName:     "focus",
		BreakSessionName:     "break",
		FocusSessionDuration: 25,
		BreakSessionDuration: 5,
	}
	if err := database.CreateSession(s); err != nil {
		t.Fatalf("create session: %v", err)
	}

	focusDuration := 50
	interrupted := true
	if err := database.UpdateSession(s.ID, db.SessionUpdate{
		FocusSessionDuration: &focusDuration,
		Interrupted:          &interrupted,
	}); err != nil {
		t.Fatalf("update session: %v", err)
	}

	got, err := database.GetSession(s.ID)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}

	if got.FocusSessionDuration != focusDuration {
		t.Errorf("FocusSessionDuration = %d, want %d", got.FocusSessionDuration, focusDuration)
	}
	if !got.Interrupted {
		t.Error("Interrupted = false, want true")
	}
	if got.FocusSessionName != "focus" {
		t.Errorf("FocusSessionName = %q, want %q", got.FocusSessionName, "focus")
	}
	if got.BreakSessionName != "break" {
		t.Errorf("BreakSessionName = %q, want %q", got.BreakSessionName, "break")
	}
	if got.BreakSessionDuration != 5 {
		t.Errorf("BreakSessionDuration = %d, want 5", got.BreakSessionDuration)
	}
}

func TestUpdateSessionEmptyUpdateIsNoop(t *testing.T) {
	database, conn := testsupport.OpenDB(t)

	s := &db.Session{
		FocusSessionName:     "focus",
		BreakSessionName:     "break",
		FocusSessionDuration: 25,
		BreakSessionDuration: 5,
	}
	if err := database.CreateSession(s); err != nil {
		t.Fatalf("create session: %v", err)
	}

	// Pin updated_at to a known past value, so a no-op is distinguishable from
	// a write of time.Now().
	seeded := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	if _, err := conn.Exec(
		"UPDATE sessions SET updated_at = ? WHERE id = ?",
		db.Timestamp(seeded),
		s.ID,
	); err != nil {
		t.Fatalf("seed updated_at: %v", err)
	}

	if err := database.UpdateSession(s.ID, db.SessionUpdate{}); err != nil {
		t.Fatalf("update session: %v", err)
	}

	got, err := database.GetSession(s.ID)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}

	if got.FocusSessionName != s.FocusSessionName ||
		got.BreakSessionName != s.BreakSessionName ||
		got.FocusSessionDuration != s.FocusSessionDuration ||
		got.BreakSessionDuration != s.BreakSessionDuration ||
		got.Interrupted != s.Interrupted {
		t.Errorf("session = %+v, want unchanged %+v", got, s)
	}
	if got.UpdatedAt.Unix() != seeded.Unix() {
		t.Errorf("UpdatedAt = %v, want unchanged %v", got.UpdatedAt, seeded)
	}
}

func TestUpdateSessionWritesEveryField(t *testing.T) {
	database, _ := testsupport.OpenDB(t)

	s := &db.Session{
		FocusSessionName:     "focus",
		BreakSessionName:     "break",
		FocusSessionDuration: 25,
		BreakSessionDuration: 5,
	}
	if err := database.CreateSession(s); err != nil {
		t.Fatalf("create session: %v", err)
	}

	focusName := "deep work"
	breakName := "rest"
	focusDuration := 30
	breakDuration := 10
	interrupted := true
	if err := database.UpdateSession(s.ID, db.SessionUpdate{
		FocusSessionName:     &focusName,
		BreakSessionName:     &breakName,
		FocusSessionDuration: &focusDuration,
		BreakSessionDuration: &breakDuration,
		Interrupted:          &interrupted,
	}); err != nil {
		t.Fatalf("update session: %v", err)
	}

	got, err := database.GetSession(s.ID)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}

	if got.FocusSessionName != focusName {
		t.Errorf("FocusSessionName = %q, want %q", got.FocusSessionName, focusName)
	}
	if got.BreakSessionName != breakName {
		t.Errorf("BreakSessionName = %q, want %q", got.BreakSessionName, breakName)
	}
	if got.FocusSessionDuration != focusDuration {
		t.Errorf("FocusSessionDuration = %d, want %d", got.FocusSessionDuration, focusDuration)
	}
	if got.BreakSessionDuration != breakDuration {
		t.Errorf("BreakSessionDuration = %d, want %d", got.BreakSessionDuration, breakDuration)
	}
	if !got.Interrupted {
		t.Error("Interrupted = false, want true")
	}
	if got.CreatedAt.Unix() != s.CreatedAt.Unix() {
		t.Errorf("CreatedAt = %v, want unchanged %v", got.CreatedAt, s.CreatedAt)
	}
	if got.UpdatedAt.Before(got.CreatedAt) {
		t.Errorf("UpdatedAt = %v, before CreatedAt = %v", got.UpdatedAt, got.CreatedAt)
	}
}

func TestDeleteSessionRemovesOnlyTargetRow(t *testing.T) {
	database, _ := testsupport.OpenDB(t)

	first := &db.Session{FocusSessionName: "first"}
	second := &db.Session{FocusSessionName: "second"}
	for _, s := range []*db.Session{first, second} {
		if err := database.CreateSession(s); err != nil {
			t.Fatalf("create session: %v", err)
		}
	}

	if err := database.DeleteSession(first.ID); err != nil {
		t.Fatalf("delete session: %v", err)
	}

	if _, err := database.GetSession(first.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("get deleted session: err = %v, want %v", err, sql.ErrNoRows)
	}

	got, err := database.GetSession(second.ID)
	if err != nil {
		t.Fatalf("get remaining session: %v", err)
	}
	if got.FocusSessionName != second.FocusSessionName {
		t.Errorf("FocusSessionName = %q, want %q", got.FocusSessionName, second.FocusSessionName)
	}
}

func TestDeleteSessionMissingIDKeepsOtherRows(t *testing.T) {
	database, _ := testsupport.OpenDB(t)

	s := &db.Session{FocusSessionName: "focus"}
	if err := database.CreateSession(s); err != nil {
		t.Fatalf("create session: %v", err)
	}

	// Current behavior: deleting an absent id is a silent no-op.
	if err := database.DeleteSession(s.ID + 100); err != nil {
		t.Errorf("delete missing session: %v, want nil", err)
	}

	got, err := database.GetSession(s.ID)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if got.FocusSessionName != s.FocusSessionName {
		t.Errorf("FocusSessionName = %q, want %q", got.FocusSessionName, s.FocusSessionName)
	}
}

func TestListSessionsNewestFirst(t *testing.T) {
	database, conn := testsupport.OpenDB(t)

	names := []string{"oldest", "middle", "newest"}
	sessions := make([]*db.Session, 0, len(names))
	for _, name := range names {
		s := &db.Session{
			FocusSessionName:     name,
			BreakSessionName:     "break-" + name,
			FocusSessionDuration: 10,
		}
		if err := database.CreateSession(s); err != nil {
			t.Fatalf("create session %q: %v", name, err)
		}
		sessions = append(sessions, s)
	}

	// Seed distinct created_at values, so the ordering assertion cannot depend
	// on tie-breaking between rows created in the same instant.
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i, s := range sessions {
		if _, err := conn.Exec(
			"UPDATE sessions SET created_at = ? WHERE id = ?",
			db.Timestamp(base.Add(time.Duration(i)*time.Minute)),
			s.ID,
		); err != nil {
			t.Fatalf("seed created_at: %v", err)
		}
	}

	got, err := database.ListSessions()
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}

	if len(got) != len(names) {
		t.Fatalf("session count = %d, want %d", len(got), len(names))
	}

	for i, name := range []string{"newest", "middle", "oldest"} {
		if got[i].FocusSessionName != name {
			t.Errorf("session[%d].FocusSessionName = %q, want %q", i, got[i].FocusSessionName, name)
		}
		if got[i].BreakSessionName != "break-"+name {
			t.Errorf("session[%d].BreakSessionName = %q, want %q", i, got[i].BreakSessionName, "break-"+name)
		}
		if got[i].FocusSessionDuration != 10 {
			t.Errorf("session[%d].FocusSessionDuration = %d, want 10", i, got[i].FocusSessionDuration)
		}
	}
}

func TestListSessionsEmpty(t *testing.T) {
	database, _ := testsupport.OpenDB(t)

	got, err := database.ListSessions()
	if err != nil {
		t.Fatalf("list sessions: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("session count = %d, want 0", len(got))
	}
}

func TestListSessionsMissingTable(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { utils.LogError(conn.Close()) })

	sessions, err := db.New(conn).ListSessions()
	if err == nil {
		t.Fatal("expected error when sessions table is missing")
	}
	if sessions != nil {
		t.Errorf("sessions = %+v, want nil", sessions)
	}
}

func TestListSessionsScanError(t *testing.T) {
	database, conn := testsupport.OpenDB(t)

	// focus_session_duration holds TEXT, so scanning that column into an int
	// fails.
	now := time.Now()
	if _, err := conn.Exec(`
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
	`, "focus", "break", "not-a-number", 5, false, db.Timestamp(now), db.Timestamp(now)); err != nil {
		t.Fatalf("insert malformed row: %v", err)
	}

	sessions, err := database.ListSessions()
	if err == nil {
		t.Fatal("expected error for unparsable focus_session_duration")
	}
	if sessions != nil {
		t.Errorf("sessions = %+v, want nil", sessions)
	}
}

func TestCreateSessionMissingTable(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { utils.LogError(conn.Close()) })

	s := &db.Session{FocusSessionName: "focus"}

	if err := db.New(conn).CreateSession(s); err == nil {
		t.Error("expected error when sessions table is missing")
	}
	if s.ID != 0 {
		t.Errorf("ID = %d, want 0 after failed insert", s.ID)
	}
}

func TestMarkAbandonedSessions(t *testing.T) {
	database, _ := testsupport.OpenDB(t)

	crashed := &db.Session{FocusSessionName: "crashed", FocusSessionDuration: 25}
	if err := database.CreateSession(crashed); err != nil {
		t.Fatalf("create crashed session: %v", err)
	}

	closed := &db.Session{FocusSessionName: "closed", FocusSessionDuration: 25}
	if err := database.CreateSession(closed); err != nil {
		t.Fatalf("create closed session: %v", err)
	}
	if err := db.CloseSession(database, closed.ID, false); err != nil {
		t.Fatalf("close session: %v", err)
	}

	interrupted := &db.Session{
		FocusSessionName:     "interrupted",
		FocusSessionDuration: 25,
		Interrupted:          true,
	}
	if err := database.CreateSession(interrupted); err != nil {
		t.Fatalf("create interrupted session: %v", err)
	}

	if err := database.MarkAbandonedSessions(); err != nil {
		t.Fatalf("mark abandoned: %v", err)
	}

	gotCrashed, err := database.GetSession(crashed.ID)
	if err != nil {
		t.Fatalf("get crashed session: %v", err)
	}
	if !gotCrashed.Interrupted {
		t.Error("crashed session not marked interrupted")
	}
	if !gotCrashed.UpdatedAt.After(gotCrashed.CreatedAt) {
		t.Error("crashed session updated_at not moved")
	}

	gotClosed, err := database.GetSession(closed.ID)
	if err != nil {
		t.Fatalf("get closed session: %v", err)
	}
	if gotClosed.Interrupted {
		t.Error("closed session marked interrupted")
	}
	if !gotClosed.UpdatedAt.After(gotClosed.CreatedAt) {
		t.Error("closed session updated_at not after created_at")
	}

	gotInterrupted, err := database.GetSession(interrupted.ID)
	if err != nil {
		t.Fatalf("get interrupted session: %v", err)
	}
	if !gotInterrupted.Interrupted {
		t.Error("interrupted session flag lost")
	}
	if !gotInterrupted.UpdatedAt.Equal(gotInterrupted.CreatedAt) {
		t.Error("interrupted session updated_at changed")
	}

	if err := database.MarkAbandonedSessions(); err != nil {
		t.Fatalf("second mark abandoned: %v", err)
	}

	again, err := database.GetSession(crashed.ID)
	if err != nil {
		t.Fatalf("get crashed session after second sweep: %v", err)
	}
	if !again.UpdatedAt.Equal(gotCrashed.UpdatedAt) {
		t.Errorf("second sweep moved updated_at: %v, want %v", again.UpdatedAt, gotCrashed.UpdatedAt)
	}
}

func TestDeleteAllSessions(t *testing.T) {
	database, conn := testsupport.OpenDB(t)

	sessions := []*db.Session{
		{FocusSessionName: "one", BreakSessionName: "break", FocusSessionDuration: 25, BreakSessionDuration: 5},
		{FocusSessionName: "two", BreakSessionName: "break", FocusSessionDuration: 25, BreakSessionDuration: 5},
	}

	for _, s := range sessions {
		if err := database.CreateSession(s); err != nil {
			t.Fatalf("create session: %v", err)
		}
	}

	tag, err := database.GetOrCreateTag("focus")
	if err != nil {
		t.Fatalf("create tag: %v", err)
	}

	if err := database.AddTagToSession(sessions[0].ID, tag.ID); err != nil {
		t.Fatalf("add tag: %v", err)
	}

	activity := &db.SessionActivity{
		SessionID: sessions[0].ID,
		App:       "com.apple.Safari",
		AppName:   "Safari",
		Phase:     db.PhaseFocus,
		StartedAt: sessions[0].CreatedAt,
		EndedAt:   sessions[0].CreatedAt.Add(time.Minute),
	}
	if err := database.CreateSessionActivity(activity); err != nil {
		t.Fatalf("create activity: %v", err)
	}

	res, err := database.DeleteAllSessions()
	if err != nil {
		t.Fatalf("delete all sessions: %v", err)
	}

	if res.Sessions != 2 || res.Activities != 1 || res.TagLinks != 1 {
		t.Fatalf("result = %+v, want 2 sessions, 1 activity, 1 tag link", res)
	}

	var gotSessions, gotActivities, gotLinks, gotTags int

	if err := conn.QueryRow(`
		SELECT
			(SELECT COUNT(*) FROM sessions),
			(SELECT COUNT(*) FROM session_activities),
			(SELECT COUNT(*) FROM session_tags),
			(SELECT COUNT(*) FROM tags)
	`).Scan(&gotSessions, &gotActivities, &gotLinks, &gotTags); err != nil {
		t.Fatalf("count rows: %v", err)
	}

	if gotSessions != 0 || gotActivities != 0 || gotLinks != 0 {
		t.Errorf("rows = sessions %d, activities %d, links %d; want all 0", gotSessions, gotActivities, gotLinks)
	}

	if gotTags != 2 {
		t.Errorf("tags = %d, want 2 kept (work from migration, focus)", gotTags)
	}
}

func TestDeleteSessionCascadesChildren(t *testing.T) {
	database, conn := testsupport.OpenDB(t)

	s := &db.Session{
		FocusSessionName:     "focus",
		BreakSessionName:     "break",
		FocusSessionDuration: 25,
		BreakSessionDuration: 5,
	}
	if err := database.CreateSession(s); err != nil {
		t.Fatalf("create session: %v", err)
	}

	tag, err := database.GetOrCreateTag("focus")
	if err != nil {
		t.Fatalf("create tag: %v", err)
	}

	if err := database.AddTagToSession(s.ID, tag.ID); err != nil {
		t.Fatalf("add tag: %v", err)
	}

	activity := &db.SessionActivity{
		SessionID: s.ID,
		App:       "com.apple.Safari",
		AppName:   "Safari",
		Phase:     db.PhaseFocus,
		StartedAt: s.CreatedAt,
		EndedAt:   s.CreatedAt.Add(time.Minute),
	}
	if err := database.CreateSessionActivity(activity); err != nil {
		t.Fatalf("create activity: %v", err)
	}

	if err := database.DeleteSession(s.ID); err != nil {
		t.Fatalf("delete session: %v", err)
	}

	var gotActivities, gotLinks int

	if err := conn.QueryRow(`
		SELECT
			(SELECT COUNT(*) FROM session_activities WHERE session_id = ?),
			(SELECT COUNT(*) FROM session_tags WHERE session_id = ?)
	`, s.ID, s.ID).Scan(&gotActivities, &gotLinks); err != nil {
		t.Fatalf("count children: %v", err)
	}

	if gotActivities != 0 || gotLinks != 0 {
		t.Errorf("children = activities %d, links %d; want both 0", gotActivities, gotLinks)
	}
}

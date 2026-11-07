package db_test

import (
	"database/sql"
	"errors"
	"github.com/hpdipto/flowterm/db"
	"github.com/hpdipto/flowterm/internal/testsupport"
	"github.com/hpdipto/flowterm/utils"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGetTagByName(t *testing.T) {
	database, _ := testsupport.OpenDB(t)

	// The seed migration inserts the "work" tag.
	tag, err := database.GetTagByName("work")
	if err != nil {
		t.Fatalf("get tag: %v", err)
	}

	if tag.Name != "work" {
		t.Errorf("Name = %q, want %q", tag.Name, "work")
	}
	if tag.ID == 0 {
		t.Error("ID = 0, want assigned id")
	}
	if tag.CreatedAt.IsZero() || tag.UpdatedAt.IsZero() {
		t.Errorf("timestamps = (%v, %v), want non-zero", tag.CreatedAt, tag.UpdatedAt)
	}
}

func TestGetTagByNameMissing(t *testing.T) {
	database, _ := testsupport.OpenDB(t)

	tag, err := database.GetTagByName("missing")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("err = %v, want %v", err, sql.ErrNoRows)
	}
	if tag != nil {
		t.Errorf("tag = %+v, want nil", tag)
	}
}

func TestCreateTag(t *testing.T) {
	database, _ := testsupport.OpenDB(t)

	before := time.Now()
	tag := &db.Tag{Name: "go"}
	if err := database.CreateTag(tag); err != nil {
		t.Fatalf("create tag: %v", err)
	}
	after := time.Now()

	if tag.ID == 0 {
		t.Error("ID = 0, want assigned id")
	}
	if tag.CreatedAt.Before(before) || tag.CreatedAt.After(after) {
		t.Errorf("CreatedAt = %v, want between %v and %v", tag.CreatedAt, before, after)
	}
	if !tag.CreatedAt.Equal(tag.UpdatedAt) {
		t.Errorf("UpdatedAt = %v, want CreatedAt = %v", tag.UpdatedAt, tag.CreatedAt)
	}

	got, err := database.GetTagByName(tag.Name)
	if err != nil {
		t.Fatalf("get tag: %v", err)
	}
	if got.ID != tag.ID {
		t.Errorf("ID = %d, want %d", got.ID, tag.ID)
	}
	if got.Name != tag.Name {
		t.Errorf("Name = %q, want %q", got.Name, tag.Name)
	}
}

func TestGetOrCreateTagCreatesOnce(t *testing.T) {
	database, conn := testsupport.OpenDB(t)

	first, err := database.GetOrCreateTag("tui")
	if err != nil {
		t.Fatalf("get or create tag: %v", err)
	}
	if first.ID == 0 {
		t.Error("ID = 0, want assigned id")
	}

	second, err := database.GetOrCreateTag("tui")
	if err != nil {
		t.Fatalf("get or create tag: %v", err)
	}
	if second.ID != first.ID {
		t.Errorf("second ID = %d, want existing %d", second.ID, first.ID)
	}

	var count int
	if err := conn.QueryRow(
		"SELECT COUNT(*) FROM tags WHERE name = ?",
		"tui",
	).Scan(&count); err != nil {
		t.Fatalf("count tags: %v", err)
	}
	if count != 1 {
		t.Errorf("tag count = %d, want 1", count)
	}
}

func TestGetOrCreateTagRejectsInvalidName(t *testing.T) {
	database, conn := testsupport.OpenDB(t)

	names := []string{
		"",
		"Go",
		"bad name",
		"bad!",
		"-go",
		"go-",
		"go_tui",
		strings.Repeat("a", 31),
	}

	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			tag, err := database.GetOrCreateTag(name)
			if err == nil {
				t.Fatalf("expected error for %q", name)
			}
			if tag != nil {
				t.Errorf("tag = %+v, want nil", tag)
			}
		})
	}

	// Rejected names must not reach the tags table; only the seed row remains.
	var count int
	if err := conn.QueryRow("SELECT COUNT(*) FROM tags").Scan(&count); err != nil {
		t.Fatalf("count tags: %v", err)
	}
	if count != 1 {
		t.Errorf("tag count = %d, want 1 (seed only)", count)
	}
}

func TestGetOrCreateTagAcceptsBoundaryNames(t *testing.T) {
	database, _ := testsupport.OpenDB(t)

	names := []string{
		"a",
		"0",
		"a-b",
		strings.Repeat("a", 30),
		strings.Repeat("a", 28) + "-1",
	}

	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			tag, err := database.GetOrCreateTag(name)
			if err != nil {
				t.Fatalf("get or create tag %q: %v", name, err)
			}
			if tag.Name != name {
				t.Errorf("Name = %q, want %q", tag.Name, name)
			}
			if tag.ID == 0 {
				t.Error("ID = 0, want assigned id")
			}
		})
	}
}

func TestCreateTagMissingTable(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { utils.LogError(conn.Close()) })

	tag := &db.Tag{Name: "go"}

	if err := db.New(conn).CreateTag(tag); err == nil {
		t.Fatal("expected error when tags table is missing")
	}
	if tag.ID != 0 {
		t.Errorf("ID = %d, want 0 after failed insert", tag.ID)
	}
}

func TestGetOrCreateTagMissingTable(t *testing.T) {
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { utils.LogError(conn.Close()) })

	tag, err := db.New(conn).GetOrCreateTag("go")
	if err == nil {
		t.Fatal("expected error when tags table is missing")
	}
	if tag != nil {
		t.Errorf("tag = %+v, want nil", tag)
	}
}

func TestGetOrCreateTagCreateFailure(t *testing.T) {
	database, conn := testsupport.OpenDB(t)

	// Force every tag insert to fail, so the lookup succeeds but the create
	// inside GetOrCreateTag does not.
	if _, err := conn.Exec(`
		CREATE TRIGGER reject_tag_insert
		BEFORE INSERT ON tags
		BEGIN
			SELECT RAISE(FAIL, 'insert rejected');
		END;
	`); err != nil {
		t.Fatalf("create trigger: %v", err)
	}

	tag, err := database.GetOrCreateTag("go")
	if err == nil {
		t.Fatal("expected error when tag insert fails")
	}
	if tag != nil {
		t.Errorf("tag = %+v, want nil", tag)
	}

	var count int
	if err := conn.QueryRow(
		"SELECT COUNT(*) FROM tags WHERE name = ?",
		"go",
	).Scan(&count); err != nil {
		t.Fatalf("count tags: %v", err)
	}
	if count != 0 {
		t.Errorf("tag count = %d, want 0", count)
	}
}

func TestAddTagToSessionIgnoresDuplicates(t *testing.T) {
	database, conn := testsupport.OpenDB(t)

	session := &db.Session{FocusSessionName: "focus"}
	if err := database.CreateSession(session); err != nil {
		t.Fatalf("create session: %v", err)
	}

	goTag, err := database.GetOrCreateTag("go")
	if err != nil {
		t.Fatalf("get or create go tag: %v", err)
	}
	otherTag, err := database.GetOrCreateTag("other")
	if err != nil {
		t.Fatalf("get or create other tag: %v", err)
	}

	if err := database.AddTagToSession(session.ID, goTag.ID); err != nil {
		t.Fatalf("add tag: %v", err)
	}
	if err := database.AddTagToSession(session.ID, goTag.ID); err != nil {
		t.Fatalf("add duplicate tag: %v", err)
	}

	var count int
	if err := conn.QueryRow(
		"SELECT COUNT(*) FROM session_tags WHERE session_id = ? AND tag_id = ?",
		session.ID,
		goTag.ID,
	).Scan(&count); err != nil {
		t.Fatalf("count session tags: %v", err)
	}
	if count != 1 {
		t.Errorf("duplicate tag count = %d, want 1", count)
	}

	if err := database.AddTagToSession(session.ID, otherTag.ID); err != nil {
		t.Fatalf("add second tag: %v", err)
	}
	if err := conn.QueryRow(
		"SELECT COUNT(*) FROM session_tags WHERE session_id = ?",
		session.ID,
	).Scan(&count); err != nil {
		t.Fatalf("count session tags: %v", err)
	}
	if count != 2 {
		t.Errorf("session tag count = %d, want 2", count)
	}
}

package main

import (
	"errors"
	"flag"
	"hpdipto/flowterm/db"
	"hpdipto/flowterm/utils"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestParseGlobal(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    []string
		wantErr bool
	}{
		{name: "no args", args: nil},
		{name: "command", args: []string{"stats"}, want: []string{"stats"}},
		{name: "unknown flag", args: []string{"-x"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseGlobal(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !slices.Equal(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRunUnknownCommand(t *testing.T) {
	if err := run([]string{"bogus"}); err == nil {
		t.Fatal("expected error for unknown command")
	}
}

func TestRunHelp(t *testing.T) {
	if err := run([]string{"-h"}); !errors.Is(err, flag.ErrHelp) {
		t.Fatalf("got %v, want flag.ErrHelp", err)
	}
}

func TestRunStatsInvalidArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "since zero", args: []string{
			"stats", "--since",
			"0",
		}},
		{name: "extra argument", args: []string{
			"stats",
			"extra",
		}},
		{name: "unknown flag", args: []string{
			"stats",
			"--bogus",
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := run(tt.args); err == nil {
				t.Fatalf("run(%v) = nil, want error", tt.args)
			}
		})
	}
}

func TestFormatTags(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"work", "#work"},
		{"work,alpha", "#work #alpha"},
	}

	for _, tt := range tests {
		if got := formatTags(tt.in); got != tt.want {
			t.Errorf("formatTags(%q) = %q, want %q", tt.in, got,
				tt.want)
		}
	}
}

func seedCLIDatabase(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "flowterm.db")

	conn, err := db.Open(path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { utils.LogError(conn.Close()) })

	if err := db.Migrate(conn); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	database := db.New(conn)

	sessions := []*db.Session{
		{
			FocusSessionName:     "Deep work",
			BreakSessionName:     "Break",
			FocusSessionDuration: 25,
			BreakSessionDuration: 5,
		},
		{
			FocusSessionName:     "Review",
			FocusSessionDuration: 10,
		},
		{
			FocusSessionName:     "Inbox",
			FocusSessionDuration: 10,
			Interrupted:          true,
		},
	}

	for _, s := range sessions {
		if err := database.CreateSession(s); err != nil {
			t.Fatalf("create session %q: %v", s.FocusSessionName, err)
		}
	}

	tags := map[int][]string{
		0: {"work", "deep"},
		1: {"work"},
		2: {"work"},
	}

	for index, names := range tags {
		for _, name := range names {
			tag, err := database.GetOrCreateTag(name)
			if err != nil {
				t.Fatalf("get or create tag %q: %v", name, err)
			}

			if err := database.AddTagToSession(sessions[index].ID, tag.ID); err != nil {
				t.Fatalf("add tag %q: %v", name, err)
			}
		}
	}

	return path
}

func inCLIWorkspace(t *testing.T, dbPath string) {
	t.Helper()

	t.Setenv("DB_NAME", dbPath)

	dir := t.TempDir()
	env := "DB_NAME=" + dbPath + "\n"

	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(env), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	t.Chdir(dir)
}

func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}

	original := os.Stdout
	os.Stdout = w

	runErr := fn()

	os.Stdout = original

	if err := w.Close(); err != nil {
		t.Fatalf("close stdout pipe: %v", err)
	}

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}

	return string(out), runErr
}

func TestRunStatsFiltersInterruptedSessions(t *testing.T) {
	path := seedCLIDatabase(t)
	inCLIWorkspace(t, path)

	completed, err := captureStdout(t, func() error { return runStats(nil) })
	if err != nil {
		t.Fatalf("runStats: %v", err)
	}

	for _, want := range []string{"Deep work", "Review", "#work", "#deep"} {
		if !strings.Contains(completed, want) {
			t.Errorf("completed stats missing %q:\n%s", want, completed)
		}
	}

	for _, unwanted := range []string{"Inbox", "interrupted"} {
		if strings.Contains(completed, unwanted) {
			t.Errorf("completed stats unexpectedly contain %q:\n%s", unwanted, completed)
		}
	}

	all, err := captureStdout(t, func() error { return runStats([]string{"--all"}) })
	if err != nil {
		t.Fatalf("runStats --all: %v", err)
	}

	for _, want := range []string{"Inbox", "interrupted", "yes"} {
		if !strings.Contains(all, want) {
			t.Errorf("--all stats missing %q:\n%s", want, all)
		}
	}
}

func TestRunStatsSummaryCountsUniqueTags(t *testing.T) {
	path := seedCLIDatabase(t)
	inCLIWorkspace(t, path)

	out, err := captureStdout(t, func() error {
		return runStats([]string{"--summary"})
	})
	if err != nil {
		t.Fatalf("runStats --summary: %v", err)
	}

	for _, want := range []string{"SESSIONS: 2", "FOCUS TIME: 35 (min)", "TAGS: 2"} {
		if !strings.Contains(out, want) {
			t.Errorf("summary missing %q:\n%s", want, out)
		}
	}
}

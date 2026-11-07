package main

import (
	"errors"
	"flag"
	"github.com/hpdipto/flowterm/db"
	"github.com/hpdipto/flowterm/utils"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestParseGlobal(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		want        []string
		wantVersion bool
		wantErr     bool
	}{
		{name: "no args", args: nil},
		{name: "command", args: []string{"stats"}, want: []string{"stats"}},
		{name: "version flag", args: []string{"--version"}, wantVersion: true},
		{name: "unknown flag", args: []string{"-x"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, versionFlag, err := parseGlobal(tt.args)
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
			if versionFlag != tt.wantVersion {
				t.Fatalf("version flag = %v, want %v", versionFlag, tt.wantVersion)
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

	for _, index := range []int{0, 1} {
		if err := db.CloseSession(database, sessions[index].ID, false); err != nil {
			t.Fatalf("close session %q: %v", sessions[index].FocusSessionName, err)
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

	now := time.Now()
	activities := []*db.SessionActivity{
		{
			SessionID: sessions[0].ID,
			App:       "com.apple.Safari",
			AppName:   "Safari",
			Phase:     db.PhaseFocus,
			StartedAt: now,
			EndedAt:   now.Add(3 * time.Second),
		},
		{
			SessionID: sessions[0].ID,
			App:       "md.obsidian",
			AppName:   "Obsidian",
			Phase:     db.PhaseBreak,
			StartedAt: now,
			EndedAt:   now.Add(5 * time.Second),
		},
	}

	for _, a := range activities {
		if err := database.CreateSessionActivity(a); err != nil {
			t.Fatalf("create activity %q: %v", a.App, err)
		}
	}

	return path
}

func inCLIWorkspace(t *testing.T, dbPath string) {
	t.Helper()

	t.Setenv("DB_NAME", dbPath)
	t.Chdir(t.TempDir())
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

	for _, want := range []string{"Deep work", "Review", "#work", "#deep", "Safari: 3s", "Obsidian: 5s"} {
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

func TestRunStatsTagAndApps(t *testing.T) {
	path := seedCLIDatabase(t)
	inCLIWorkspace(t, path)

	deep, err := captureStdout(t, func() error {
		return runStats([]string{"--tag", "deep"})
	})
	if err != nil {
		t.Fatalf("runStats --tag deep: %v", err)
	}
	if !strings.Contains(deep, "Deep work") {
		t.Errorf("tag deep output missing Deep work:\n%s", deep)
	}
	if strings.Contains(deep, "Review") {
		t.Errorf("tag deep output contains Review:\n%s", deep)
	}

	work, err := captureStdout(t, func() error {
		return runStats([]string{"--tag", "work"})
	})
	if err != nil {
		t.Fatalf("runStats --tag work: %v", err)
	}
	for _, want := range []string{"Deep work", "Review"} {
		if !strings.Contains(work, want) {
			t.Errorf("tag work output missing %q:\n%s", want, work)
		}
	}
	if strings.Contains(work, "Inbox") {
		t.Errorf("tag work output contains interrupted session:\n%s", work)
	}

	apps, err := captureStdout(t, func() error {
		return runStats([]string{"--apps"})
	})
	if err != nil {
		t.Fatalf("runStats --apps: %v", err)
	}
	for _, want := range []string{"App", "Focus", "Break", "Safari", "Obsidian", "3s", "5s"} {
		if !strings.Contains(apps, want) {
			t.Errorf("apps output missing %q:\n%s", want, apps)
		}
	}
}

func TestRunDeleteInvalidArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "missing all flag", args: []string{"delete"}},
		{name: "extra argument", args: []string{"delete", "--all", "extra"}},
		{name: "unknown flag", args: []string{"delete", "--bogus"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := run(tt.args); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestDBPath(t *testing.T) {
	t.Run("defaults under config dir", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)
		t.Setenv("DB_NAME", "")

		path, err := dbPath()
		if err != nil {
			t.Fatalf("dbPath: %v", err)
		}

		want := filepath.Join(home, ".config", "flowterm", "flowterm.db")
		if path != want {
			t.Fatalf("dbPath = %q, want %q", path, want)
		}

		info, err := os.Stat(filepath.Dir(path))
		if err != nil {
			t.Fatalf("stat dir: %v", err)
		}
		if !info.IsDir() {
			t.Fatalf("%q is not a directory", filepath.Dir(path))
		}
	})

	t.Run("DB_NAME overrides", func(t *testing.T) {
		want := filepath.Join(t.TempDir(), "custom.db")
		t.Setenv("DB_NAME", want)

		path, err := dbPath()
		if err != nil {
			t.Fatalf("dbPath: %v", err)
		}
		if path != want {
			t.Fatalf("dbPath = %q, want %q", path, want)
		}
	})
}

package main

import (
	"errors"
	"flag"
	"slices"
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

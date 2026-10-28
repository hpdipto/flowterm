package main

import (
	"database/sql"
	"flag"
	"fmt"
	"hpdipto/flowterm/db"
	"hpdipto/flowterm/tui"
	"hpdipto/flowterm/utils"
	"os"
	"strconv"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/charmbracelet/x/term"
	"github.com/joho/godotenv"
)

func parseGlobal(args []string) ([]string, error) {
	fs := flag.NewFlagSet("flowterm", flag.ContinueOnError)

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	return fs.Args(), nil
}

func run(args []string) error {
	rest, err := parseGlobal(args)
	if err != nil {
		return err
	}

	if len(rest) == 0 {
		return runTUI()
	}

	switch rest[0] {
	case "stats":
		return runStats(rest[1:])
	default:
		return fmt.Errorf("unknown command %q", rest[0])
	}
}

func openConn() (*sql.DB, error) {
	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	conn, err := db.Open(os.Getenv("DB_NAME"))
	if err != nil {
		return nil, err
	}

	if err := db.Migrate(conn); err != nil {
		utils.LogDeferredError(conn.Close)
		return nil, err
	}

	return conn, nil
}

func runTUI() error {
	if err := godotenv.Load(); err != nil {
		return err
	}

	conn, err := db.Open(os.Getenv("DB_NAME"))
	if err != nil {
		return err
	}
	defer utils.LogDeferredError(conn.Close)

	database := db.New(conn)

	if err := db.Migrate(conn); err != nil {
		return err
	}

	tui.InitTUi(database)
	return nil
}

func formatTags(tags string) string {
	if tags == "" {
		return ""
	}
	return "#" + strings.ReplaceAll(tags, ",", " #")
}

func runStats(args []string) error {
	fs := flag.NewFlagSet("flowterm stats", flag.ContinueOnError)
	since := fs.Int("since", 1, "days to include, counting back from today")
	all := fs.Bool("all", false, "include interrupted sessions")
	summary := fs.Bool("summary", false, "print session totals above the table")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}

	if *since < 1 {
		return fmt.Errorf("--since must be at least 1")
	}

	conn, err := openConn()
	if err != nil {
		return err
	}
	defer utils.LogDeferredError(conn.Close)

	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0,
		0, 0, now.Location()).
		AddDate(0, 0, -(*since - 1))

	stats, err := db.New(conn).ListSessionStats(start, *all)
	if err != nil {
		return err
	}

	headers := []string{"name", "duration", "break_name", "break_duration", "tags"}
	if *all {
		headers = append(headers, "interrupted")
	}

	rows := make([][]string, 0, len(stats))
	for _, s := range stats {
		tags := formatTags(s.Tags)

		row := []string{
			s.FocusSessionName,
			strconv.Itoa(s.FocusSessionDuration),
			s.BreakSessionName,
			strconv.Itoa(s.BreakSessionDuration),
			tags,
		}

		if *all {
			interrupted := "no"
			if s.Interrupted {
				interrupted = "yes"
			}
			row = append(row, interrupted)
		}

		rows = append(rows, row)
	}

	if *summary {
		totalDuration := 0
		for _, s := range stats {
			totalDuration += s.FocusSessionDuration
		}

		card := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 3)

		sessions := card.Render(fmt.Sprintf("SESSIONS: %d", len(stats)))
		focus := card.Render(fmt.Sprintf("FOCUS TIME: %d (min)", totalDuration))

		cards := lipgloss.JoinHorizontal(lipgloss.Center, sessions, " ", focus)

		if width, _, err := term.GetSize(os.Stdout.Fd()); err == nil {
			cards = lipgloss.PlaceHorizontal(width, lipgloss.Center, cards)
		}

		fmt.Println(cards)
	}

	t := table.New().
		Headers(headers...).
		Rows(rows...).
		BorderRow(true)

	if width, _, err := term.GetSize(os.Stdout.Fd()); err == nil {
		t.Width(width)
	}

	fmt.Println(t.Render())
	return nil
}

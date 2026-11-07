package main

import (
	"database/sql"
	"flag"
	"fmt"
	"github.com/hpdipto/flowterm/db"
	"github.com/hpdipto/flowterm/tui"
	"github.com/hpdipto/flowterm/utils"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/charmbracelet/x/term"
)

func dbPath() (string, error) {
	if path := os.Getenv("DB_NAME"); path != "" {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(home, ".config", "flowterm")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	return filepath.Join(dir, "flowterm.db"), nil
}

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
	case "delete":
		return runDelete(rest[1:])
	default:
		return fmt.Errorf("unknown command %q", rest[0])
	}
}

func openConn() (*sql.DB, error) {
	path, err := dbPath()
	if err != nil {
		return nil, err
	}

	conn, err := db.Open(path)
	if err != nil {
		return nil, err
	}

	if err := db.Migrate(conn); err != nil {
		utils.LogDeferredError(conn.Close)
		return nil, err
	}

	if err := db.New(conn).MarkAbandonedSessions(); err != nil {
		utils.LogDeferredError(conn.Close)
		return nil, err
	}

	return conn, nil
}

func runTUI() error {
	path, err := dbPath()
	if err != nil {
		return err
	}

	lock, err := acquireLock(path)
	if err != nil {
		return err
	}
	defer utils.LogDeferredError(lock.Close)

	conn, err := db.Open(path)
	if err != nil {
		return err
	}
	defer utils.LogDeferredError(conn.Close)

	database := db.New(conn)

	if err := db.Migrate(conn); err != nil {
		return err
	}

	if err := database.MarkAbandonedSessions(); err != nil {
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

func activityCell(stats []db.ActivityStat) string {
	cells := make([]string, len(stats))
	for i, a := range stats {
		cells[i] = fmt.Sprintf("%s: %s", a.Name, a.Duration.Round(time.Second))
	}

	return strings.Join(cells, "\n")
}

func runStats(args []string) error {
	fs := flag.NewFlagSet("flowterm stats", flag.ContinueOnError)
	since := fs.Int("since", 1, "days to include, counting back from today")
	all := fs.Bool("all", false, "include interrupted sessions")
	summary := fs.Bool("summary", false, "print session totals above the table")
	tag := fs.String("tag", "", "only sessions with this tag")
	apps := fs.Bool("apps", false, "print app totals")

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

	activities, err := db.New(conn).ListSessionActivityStats(start, *all)
	if err != nil {
		return err
	}

	if *tag != "" {
		kept := stats[:0]
		for _, s := range stats {
			for _, name := range strings.Split(s.Tags, ",") {
				if name == *tag {
					kept = append(kept, s)
					break
				}
			}
		}
		stats = kept
	}

	headers := []string{"Focus", "Duration", "Break", "Duration", "Tags", "Focus Activities", "Break Activities"}
	if *all {
		headers = append(headers, "interrupted")
	}
	headers = append(headers, "Date")

	rows := make([][]string, 0, len(stats))
	for _, s := range stats {
		tags := formatTags(s.Tags)

		acts := activities[s.ID]

		row := []string{
			s.FocusSessionName,
			strconv.Itoa(s.FocusSessionDuration),
			s.BreakSessionName,
			strconv.Itoa(s.BreakSessionDuration),
			tags,
			activityCell(acts.Focus),
			activityCell(acts.Break),
		}

		if *all {
			interrupted := "no"
			if s.Interrupted {
				interrupted = "yes"
			}
			row = append(row, interrupted)
		}

		row = append(row, s.CreatedAt.Format("02 Jan 2006 15:04"))

		rows = append(rows, row)
	}

	if *summary {
		totalDuration := 0
		tagSet := map[string]struct{}{}
		for _, s := range stats {
			totalDuration += s.FocusSessionDuration

			for _, t := range strings.Split(s.Tags, ",") {
				if t != "" {
					tagSet[t] = struct{}{}
				}
			}
		}

		card := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 3)

		sessions := card.Render(fmt.Sprintf("SESSIONS: %d", len(stats)))
		focus := card.Render(fmt.Sprintf("FOCUS TIME: %d (min)", totalDuration))
		tags := card.Render(fmt.Sprintf("TAGS: %d", len(tagSet)))

		cards := lipgloss.JoinHorizontal(lipgloss.Center, sessions, " ", focus, " ", tags)

		if width, _, err := term.GetSize(os.Stdout.Fd()); err == nil {
			cards = lipgloss.PlaceHorizontal(width, lipgloss.Center, cards)
		}

		fmt.Println(cards)
	}

	t := table.New().
		Headers(headers...).
		Rows(rows...).
		Border(lipgloss.RoundedBorder()).
		BorderRow(true)

	if width, _, err := term.GetSize(os.Stdout.Fd()); err == nil {
		t.Width(width)
	}

	fmt.Println(t.Render())

	if *apps {
		fmt.Println(appTable(stats, activities))
	}

	return nil
}

type appTotal struct {
	name  string
	focus time.Duration
	rest  time.Duration
}

func appTable(sessions []db.SessionStat, activities map[int64]db.ActivityStats) string {
	totals := map[string]*appTotal{}
	add := func(stats []db.ActivityStat, focus bool) {
		for _, a := range stats {
			total := totals[a.App]
			if total == nil {
				total = &appTotal{name: a.Name}
				totals[a.App] = total
			}
			if total.name == "" && a.Name != "" {
				total.name = a.Name
			}
			if focus {
				total.focus += a.Duration
			} else {
				total.rest += a.Duration
			}
		}
	}

	for _, s := range sessions {
		acts := activities[s.ID]
		add(acts.Focus, true)
		add(acts.Break, false)
	}

	apps := make([]string, 0, len(totals))
	for app := range totals {
		apps = append(apps, app)
	}
	sort.Slice(apps, func(i, j int) bool {
		a := totals[apps[i]]
		b := totals[apps[j]]
		return a.focus+a.rest > b.focus+b.rest
	})

	rows := make([][]string, 0, len(apps))
	for _, app := range apps {
		total := totals[app]
		name := total.name
		if name == "" {
			name = app
		}
		rows = append(rows, []string{
			name,
			total.focus.Round(time.Second).String(),
			total.rest.Round(time.Second).String(),
		})
	}

	t := table.New().
		Headers("App", "Focus", "Break").
		Rows(rows...).
		Border(lipgloss.RoundedBorder()).
		BorderRow(true)

	if width, _, err := term.GetSize(os.Stdout.Fd()); err == nil {
		t.Width(width)
	}

	return t.Render()
}

func runDelete(args []string) error {
	fs := flag.NewFlagSet("flowterm delete", flag.ContinueOnError)
	all := fs.Bool("all", false, "delete every session with its activities and tag links")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected argument %q", fs.Arg(0))
	}

	if !*all {
		return fmt.Errorf("--all is required")
	}

	conn, err := openConn()
	if err != nil {
		return err
	}
	defer utils.LogDeferredError(conn.Close)

	res, err := db.New(conn).DeleteAllSessions()
	if err != nil {
		return err
	}

	fmt.Printf(
		"deleted %d sessions, %d activities, %d tag links\n",
		res.Sessions, res.Activities, res.TagLinks,
	)

	return nil
}

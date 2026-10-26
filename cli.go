package main

import (
	"database/sql"
	"flag"
	"fmt"
	"hpdipto/flowterm/db"
	"hpdipto/flowterm/tui"
	"hpdipto/flowterm/utils"
	"os"
	"text/tabwriter"
	"time"

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

func runStats(args []string) error {
	fs := flag.NewFlagSet("flowterm stats", flag.ContinueOnError)
	since := fs.Int("since", 1, "days to include, counting back from today")
	interrupted := fs.Bool("interrupted", false, "only interrupted sessions")

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

	stats, err := db.New(conn).ListSessionStats(start,
		*interrupted)
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "name\tduration\tbreak_name\tbreak_duration\ttags")

	for _, s := range stats {
		fmt.Fprintf(w, "%s\t%d\t%s\t%d\t%s\n",
			s.FocusSessionName,
			s.FocusSessionDuration,
			s.BreakSessionName,
			s.BreakSessionDuration,
			s.Tags,
		)
	}

	return w.Flush()
}

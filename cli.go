package main

import (
	"flag"
	"fmt"
	"hpdipto/flowterm/db"
	"hpdipto/flowterm/tui"
	"hpdipto/flowterm/utils"
	"os"

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

	return fmt.Errorf("unknown command %q", rest[0])
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

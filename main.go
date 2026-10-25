package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

func main() {
	err := run(os.Args[1:])

	if errors.Is(err, flag.ErrHelp) {
		return
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

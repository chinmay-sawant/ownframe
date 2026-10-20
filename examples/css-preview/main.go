package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	perf := flag.Bool("perf", false, "enable perf sampling and open the devtools dock")
	flag.Parse()
	if flag.NArg() > 0 {
		fmt.Fprintln(os.Stderr, "usage: go run ./examples/css-preview [-perf]")
		os.Exit(2)
	}
	if err := run(*perf); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

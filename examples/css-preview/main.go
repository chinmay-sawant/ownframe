package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	perf := flag.Bool("perf", false, "enable perf sampling and open the devtools dock")
	width := flag.Int("width", 960, "window width in CSS pixels")
	height := flag.Int("height", 860, "window height in CSS pixels")
	flag.Parse()
	if flag.NArg() > 0 || *width <= 0 || *height <= 0 {
		fmt.Fprintln(os.Stderr, "usage: go run ./examples/css-preview [-perf] [-width N] [-height N]")
		os.Exit(2)
	}
	if err := run(*perf, *width, *height); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

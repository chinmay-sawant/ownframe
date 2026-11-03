package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	perf := flag.Bool("perf", false, "enable perf sampling and open the devtools dock (sampling feeds -debug-addr)")
	debugAddr := flag.String("debug-addr", "", "serve window profiling on this loopback address (off by default; needs -tags ownframe_debug)")
	width := flag.Int("width", defaultWidth, "window width in CSS pixels")
	height := flag.Int("height", defaultHeight, "window height in CSS pixels")
	flag.Parse()
	if flag.NArg() > 0 || *width <= 0 || *height <= 0 {
		fmt.Fprintln(os.Stderr, "usage: go run ./examples/css-preview [-perf] [-debug-addr ADDR] [-width N] [-height N]")
		os.Exit(exitUsage)
	}
	if err := run(*perf, *debugAddr, *width, *height); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Default window size and the usage-error exit code.
const (
	defaultWidth  = 960
	defaultHeight = 860
	exitUsage     = 2
)

//go:build !js

// Command departure-board opens the split-flap departure board example in a
// native window. Pass -web to serve the picture in a browser instead, and
// -livery to start on another livery. Pass -dump to measure the board
// headless, writing JSON under -out.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/departure-board/board"
	"github.com/chinmay-sawant/ownframe/examples/departure-board/dump"
)

func main() {
	webMode := flag.Bool("web", false, "serve the picture in a browser on -addr")
	addr := flag.String("addr", "127.0.0.1:8137", "listen address for -web")
	livery := flag.String("livery", "amber", "livery to start on: amber, flap, lcd, or paper")
	dumpMode := flag.Bool("dump", false, "measure headless instead of opening a window")
	scenario := flag.String("scenario", "", "comma-separated dump scenarios (empty means all)")
	out := flag.String("out", "temp/board-dump", "dump output directory")
	cpu := flag.String("cpu", "", "CPU profile path for -dump")
	pace := flag.Duration("pace", 0, "sleep between -dump ticks to mimic the frame cadence, for example 16ms")
	perf := flag.Bool("perf", false, "enable perf sampling for the devtools Frame tab")
	flag.Parse()

	if *dumpMode {
		dump.Pace = *pace

		if err := dump.RunFlag(*scenario, *out, *cpu); err != nil {
			log.Fatal(err)
		}
		return
	}

	app, err := board.New()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	if *perf {
		app.Page().Window().SetPerf(true)
	}

	if *livery != board.DefaultLivery {
		if err := app.SetLivery(*livery); err != nil {
			log.Fatal(err)
		}
	}

	if err := app.Redraw(ctx); err != nil {
		log.Fatal(err)
	}

	if *webMode {
		if err := ownframe.ServeWithOptions(ctx, app.Page(), *addr, ownframe.ServeOptions{InsecureAllowPprof: *perf}); err != nil {
			log.Fatal(err)
		}

		return
	}

	if err := ownframe.Run(ctx, app.Page()); err != nil {
		log.Fatal(err)
	}
}

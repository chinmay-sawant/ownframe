//go:build !js

// Command tetris opens the tetris example in a native window.
package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/tetris/scene"
	"github.com/chinmay-sawant/ownframe/examples/tetris/store"
)

func main() {
	dir := flag.String("data", "", "score database directory override")
	seed := flag.Uint64("seed", 0, "piece seed; 0 uses the clock")
	perf := flag.Bool("perf", false, "record pipeline timing and window sampling for the Frame tab and -debug-addr")
	debugAddr := flag.String("debug-addr", "", "serve window profiling on this loopback address (off by default; needs -tags ownframe_debug)")
	flag.Parse()

	if *seed == 0 {
		*seed = uint64(time.Now().UnixNano())
	}

	opts := scene.Options{Focused: ebiten.IsFocused, Perf: *perf}

	st, err := store.Open(*dir)
	if err != nil {
		log.Printf("tetris: %v; falling back to a temporary in-memory store", err)

		st, err = store.OpenMemory()
		if err != nil {
			log.Printf("tetris: memory store: %v; scores will not be saved", err)

			st = nil
		} else {
			opts.Status = "STORAGE TEMPORARY - SCORES NOT KEPT"
		}
	}

	var worker scene.Store
	if st != nil {
		worker = scene.NewWorker(st)
	}

	app, err := scene.New(scene.NewCore(*seed), worker, opts)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	run := ownframe.Run
	if *perf || *debugAddr != "" {
		run = func(ctx context.Context, page *ownframe.Page) error {
			return ownframe.RunWithOptions(ctx, page, ownframe.WindowOptions{Perf: *perf, DebugAddr: *debugAddr})
		}
	}

	if err := run(ctx, app.Page()); err != nil {
		log.Fatal(err)
	}

	if err := app.Close(); err != nil {
		log.Printf("tetris: shutdown: %v", err)
	}
}

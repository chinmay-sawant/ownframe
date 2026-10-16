//go:build !js

// Command pocket-synth opens the Pocket Synthesizer example in a native window.
// Pass -web to serve the synthesizer in a browser instead.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/pocket-synth/synth"
)

func main() {
	webMode := flag.Bool("web", false, "serve the picture in a browser on -addr")
	addr := flag.String("addr", "127.0.0.1:8141", "listen address for -web")
	perf := flag.Bool("perf", false, "enable perf sampling for the devtools Frame tab")
	flag.Parse()

	app, err := synth.New()
	if err != nil {
		log.Fatal(err)
	}
	defer app.Close()

	if *perf {
		app.Page().SetPerf(true)
	}

	ctx := context.Background()

	if *webMode {
		if err := ownframe.ServeWithOptions(ctx, app.Page(), *addr, ownframe.ServeOptions{Perf: *perf}); err != nil {
			log.Fatal(err)
		}
		return
	}

	if err := ownframe.RunWithOptions(ctx, app.Page(), ownframe.WindowOptions{Perf: *perf}); err != nil {
		log.Fatal(err)
	}
}

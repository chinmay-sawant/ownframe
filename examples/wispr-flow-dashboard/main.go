//go:build !js

// Command wispr-flow-dashboard opens the Wispr Flow app example.
// Pass -web to serve the picture in a browser instead.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/wispr-flow-dashboard/app"
)

func main() {
	webMode := flag.Bool("web", false, "serve the picture in a browser on -addr")
	addr := flag.String("addr", "127.0.0.1:8117", "listen address for -web")
	perf := flag.Bool("perf", false, "enable perf sampling for the devtools Frame tab")
	flag.Parse()

	app, err := app.New()
	if err != nil {
		log.Fatal(err)
	}

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

	if err := ownframe.Run(ctx, app.Page()); err != nil {
		log.Fatal(err)
	}
}

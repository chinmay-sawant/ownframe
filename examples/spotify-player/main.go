//go:build !js

// Command spotify-player opens the Spotify-like player example.
// Pass -web to serve the picture in a browser instead.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/spotify-player/player"
)

func main() {
	webMode := flag.Bool("web", false, "serve the picture in a browser on -addr")
	addr := flag.String("addr", "127.0.0.1:8119", "listen address for -web")
	perf := flag.Bool("perf", false, "enable perf sampling for the devtools Frame tab")
	flag.Parse()

	app, err := player.New()
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

	if err := ownframe.Run(ctx, app.Page()); err != nil {
		log.Fatal(err)
	}
}

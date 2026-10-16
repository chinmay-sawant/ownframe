//go:build !js

// Command cyber-pet runs the interactive cyber pet companion in a window.
// Pass -web to serve the picture in a browser instead.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/cyber-pet/pet"
)

func main() {
	web := flag.Bool("web", false, "serve the picture in a browser on -addr")
	addr := flag.String("addr", "127.0.0.1:8142", "listen address for -web")
	perf := flag.Bool("perf", false, "enable perf sampling for the devtools Frame tab")
	flag.Parse()

	app, err := pet.New()
	if err != nil {
		log.Fatal(err)
	}

	if *perf {
		app.Page().SetPerf(true)
	}

	ctx := context.Background()
	if *web {
		if err := ownframe.ServeWithOptions(ctx, app.Page(), *addr, ownframe.ServeOptions{Perf: *perf}); err != nil {
			log.Fatal(err)
		}
		return
	}

	if err := ownframe.RunWithOptions(ctx, app.Page(), ownframe.WindowOptions{Perf: *perf}); err != nil {
		log.Fatal(err)
	}
}

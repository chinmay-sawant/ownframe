//go:build !js

// Command trade opens the Meridian trading terminal example.
// Pass -web to serve the static picture in a browser instead.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/chinmay-sawant/ownframe"
)

func main() {
	webMode := flag.Bool("web", false, "serve the picture in a browser on -addr")
	addr := flag.String("addr", "127.0.0.1:8144", "listen address for -web")
	perf := flag.Bool("perf", true, "enable pipeline sampling for the devtools Frame tab")
	dev := flag.Bool("dev", true, "open the developer tools dock at startup")
	flag.Parse()

	app, err := New()
	if err != nil {
		log.Fatal(err)
	}

	app.Page().Window().SetDevTools(*dev)

	if *perf {
		app.Page().Window().SetPerf(true)
	}

	ctx := context.Background()

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

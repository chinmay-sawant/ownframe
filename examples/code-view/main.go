//go:build !js

// Command code-view opens the Code Review example.
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
	addr := flag.String("addr", "127.0.0.1:8145", "listen address for -web")
	root := flag.String("root", ".", "directory to browse; defaults to the current directory")
	dev := flag.Bool("dev", true, "open the developer tools dock at startup")
	perf := flag.Bool("perf", true, "enable pipeline sampling for the devtools Frame tab")
	flag.Parse()

	app, err := New(*root)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	app.reloadChanges(ctx, false)
	app.SetDevTools(*dev)

	if *perf {
		app.Page().Window().SetPerf(true)
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

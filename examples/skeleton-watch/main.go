//go:build !js

// Command skeleton-watch opens a mechanical skeleton watch and tourbillon in a
// native desktop window. Pass -web to serve the picture in a browser on port
// 8143 instead.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/skeleton-watch/watch"
)

func main() {
	webMode := flag.Bool("web", false, "serve watch in browser on -addr")
	addr := flag.String("addr", "127.0.0.1:8143", "listen address for -web")
	perf := flag.Bool("perf", false, "enable performance sampling")
	flag.Parse()

	app, err := watch.New()
	if err != nil {
		log.Fatalf("skeleton watch new: %v", err)
	}

	ctx := context.Background()
	if *perf {
		app.Page().SetPerf(true)
	}

	if err := app.Redraw(ctx); err != nil {
		log.Fatalf("skeleton watch redraw: %v", err)
	}

	if *webMode {
		opts := ownframe.ServeOptions{Perf: *perf}
		if err := ownframe.ServeWithOptions(ctx, app.Page(), *addr, opts); err != nil {
			log.Fatalf("serve watch: %v", err)
		}
		return
	}

	if err := ownframe.RunWithOptions(ctx, app.Page(), ownframe.WindowOptions{Perf: *perf}); err != nil {
		log.Fatalf("run watch: %v", err)
	}
}

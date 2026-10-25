//go:build !js

// Command motion-conformance shows the CSS animation and transition property
// classes the automatic driver ships, one card per class. The page registers no
// tick callback, builds no player, and imports no internal package.
// Pass -web to serve the picture in a browser on -addr instead of a native
// window. Serve does not tick, so the browser shows the committed frame.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/chinmay-sawant/ownframe"
)

func main() {
	webMode := flag.Bool("web", false, "serve the picture in a browser on -addr")
	addr := flag.String("addr", "127.0.0.1:8143", "listen address for -web")
	flag.Parse()

	p, err := newPage()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	if *webMode {
		if err := ownframe.Serve(ctx, p, *addr); err != nil {
			log.Fatal(err)
		}

		return
	}

	if err := ownframe.Run(ctx, p); err != nil {
		log.Fatal(err)
	}
}

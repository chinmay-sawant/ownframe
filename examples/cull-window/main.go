//go:build !js

// Command cull-window demonstrates opt-in viewport culling (Config.CullWindow)
// on a 1,000-row fixed-height HTML/CSS list. Pass -web to serve in a browser.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/chinmay-sawant/ownframe"
)

func main() {
	webMode := flag.Bool("web", false, "serve the picture in a browser on -addr")
	addr := flag.String("addr", "127.0.0.1:8124", "listen address for -web")
	flag.Parse()

	p, err := newCullWindowPage(true)
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

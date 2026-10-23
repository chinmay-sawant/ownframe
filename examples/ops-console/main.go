//go:build !js

// Command ops-console opens the five-screen yard ops sample.
// Pass -web to serve the picture in a browser instead.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/ops-console/console"
)

func main() {
	webMode := flag.Bool("web", false, "serve the picture in a browser on -addr")
	addr := flag.String("addr", "127.0.0.1:8141", "listen address for -web")
	flag.Parse()

	app, err := console.New()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	if *webMode {
		if err := ownframe.Serve(ctx, app.Page(), *addr); err != nil {
			log.Fatal(err)
		}

		return
	}

	if err := ownframe.Run(ctx, app.Page()); err != nil {
		log.Fatal(err)
	}
}

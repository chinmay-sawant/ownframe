//go:build !js

// Command departure-board opens the split-flap departure board example in a
// native window. Pass -web to serve the picture in a browser instead, and
// -livery to start on another livery.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/departure-board/board"
)

func main() {
	webMode := flag.Bool("web", false, "serve the picture in a browser on -addr")
	addr := flag.String("addr", "127.0.0.1:8137", "listen address for -web")
	livery := flag.String("livery", "amber", "livery to start on: amber, flap, lcd, or paper")
	flag.Parse()

	app, err := board.New()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	if *livery != board.DefaultLivery {
		if err := app.SetLivery(*livery); err != nil {
			log.Fatal(err)
		}
	}

	if err := app.Redraw(ctx); err != nil {
		log.Fatal(err)
	}

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

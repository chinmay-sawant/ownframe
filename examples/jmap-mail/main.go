//go:build !js

// Command jmap-mail showcases the public OwnFrame UI builder with a local,
// JMAP-inspired mail client. No account or JMAP endpoint is required.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/chinmay-sawant/ownframe"
)

func main() {
	webMode := flag.Bool("web", false, "serve a browser preview instead of opening a native window")
	addr := flag.String("addr", "127.0.0.1:8132", "address for -web")
	flag.Parse()

	client, err := newApp()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	if err := client.Page().Redraw(ctx); err != nil {
		log.Fatal(err)
	}

	if *webMode {
		if err := ownframe.Serve(ctx, client.Page(), *addr); err != nil {
			log.Fatal(err)
		}
		return
	}
	// Use OwnFrame's native drag region and custom chrome, not JavaScript.
	// The toolbar is the first 78 CSS pixels; clicks on its controls still work.
	options := ownframe.WindowOptions{
		CustomChrome: true,
		PlacementID:  "jmap-mail",
		Draggable: func(x, y int) bool {
			return x >= 0 && y >= 0 && y < 78
		},
	}
	if err := ownframe.RunWithOptions(ctx, client.Page(), options); err != nil {
		log.Fatal(err)
	}
}

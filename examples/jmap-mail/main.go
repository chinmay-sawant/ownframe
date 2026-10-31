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
	if err := ownframe.Run(ctx, client.Page()); err != nil {
		log.Fatal(err)
	}
}

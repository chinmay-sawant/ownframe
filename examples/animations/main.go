//go:build !js

// Command animations launches the CSS animation showcase gallery.
// Pass -web to serve the picture in a browser on -addr instead of a native window.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/chinmay-sawant/ownframe"
)

func main() {
	webMode := flag.Bool("web", false, "serve the picture in a browser on -addr")
	addr := flag.String("addr", "127.0.0.1:8142", "listen address for -web")
	devtools := flag.Bool("dev", true, "open the developer tools dock (default true; use -dev=false to close)")
	flag.Parse()

	app, err := newApp(*devtools)
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

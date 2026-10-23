//go:build !js

// Command web opens the web mode example in a native window.
// Pass -web to serve the picture in a browser instead.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/web/web"
)

func main() {
	webMode := flag.Bool("web", false, "serve the picture in a browser on -addr")
	addr := flag.String("addr", "127.0.0.1:8110", "listen address for -web")
	flag.Parse()

	app, err := web.New()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	if *webMode {
		// Serve owns the HTTP routes: GET /, GET /frame.png, POST /click,
		// POST /type, and POST /backspace act on this same page.
		if err := ownframe.Serve(ctx, app.Page(), *addr); err != nil {
			log.Fatal(err)
		}

		return
	}

	if err := ownframe.Run(ctx, app.Page()); err != nil {
		log.Fatal(err)
	}
}

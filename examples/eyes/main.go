//go:build !js

// Command eyes opens the cute robot eyes demo.
// Pass -web to serve the picture in a browser instead.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/eyes/eyes"
)

func main() {
	web := flag.Bool("web", false, "serve the picture in a browser on -addr")
	addr := flag.String("addr", "127.0.0.1:8140", "listen address for -web")
	flag.Parse()
	app, err := eyes.New()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	if *web {
		if err := ownframe.Serve(ctx, app.Page(), *addr); err != nil {
			log.Fatal(err)
		}
		return
	}
	if err := ownframe.Run(ctx, app.Page()); err != nil {
		log.Fatal(err)
	}
}

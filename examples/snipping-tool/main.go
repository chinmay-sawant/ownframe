//go:build !js

// Command snipping-tool opens the desktop snipping tool in a window.
// Pass -web to serve a still preview in a browser instead.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/snipping-tool/snip"
)

func main() {
	webMode := flag.Bool("web", false, "serve a still preview in a browser on -addr")
	addr := flag.String("addr", "127.0.0.1:8136", "listen address for -web")
	data := flag.String("data", "", "save directory override (default os.UserConfigDir()/ownframe/snipping-tool)")
	flag.Parse()

	app, err := snip.New()
	if err != nil {
		log.Fatal(err)
	}

	if *data != "" {
		app.SetSaveDir(*data)
	}

	ctx := context.Background()

	if *webMode {
		app.SetWeb(true)

		if err := ownframe.Serve(ctx, app.Page(), *addr); err != nil {
			log.Fatal(err)
		}

		return
	}

	if err := ownframe.Run(ctx, app.Page()); err != nil {
		log.Fatal(err)
	}
}

//go:build !js

// Command login opens the sign-in example in a native window.
// Pass -web to serve the picture in a browser instead.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/login/login"
)

func main() {
	webMode := flag.Bool("web", false, "serve the picture in a browser on -addr (loopback only, plaintext HTTP; demo creds travel in clear)")
	addr := flag.String("addr", "127.0.0.1:8091", "listen address for -web (keep loopback; HTTP is plaintext)")
	flag.Parse()

	app, err := login.New()
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

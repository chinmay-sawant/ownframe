//go:build js

// Command spotify-player draws the Spotify-like player example into the browser canvas.
package main

import (
	"context"
	"log"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/spotify-player/player"
)

func main() {
	app, err := player.New()
	if err != nil {
		log.Fatal(err)
	}

	defer app.Close()

	if err := ownframe.Run(context.Background(), app.Page()); err != nil {
		log.Fatal(err)
	}
}

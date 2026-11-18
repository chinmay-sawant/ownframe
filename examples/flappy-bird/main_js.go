//go:build js

// Command flappy-bird draws the Flappy Bird example into the browser canvas.
package main

import (
	"context"
	"log"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/flappy-bird/flappy"
)

func main() {
	app, err := flappy.New()
	if err != nil {
		log.Fatal(err)
	}

	// On wasm, ownframe.Run draws into the browser canvas and blocks.
	if err := ownframe.Run(context.Background(), app.Page()); err != nil {
		log.Fatal(err)
	}
}

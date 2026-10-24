//go:build !js

// Command grid opens a CSS grid example in a native window.
package main

import (
	"context"
	"log"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/grid/grid"
)

func main() {
	app, err := grid.New()
	if err != nil {
		log.Fatal(err)
	}

	if err := ownframe.Run(context.Background(), app.Page()); err != nil {
		log.Fatal(err)
	}
}

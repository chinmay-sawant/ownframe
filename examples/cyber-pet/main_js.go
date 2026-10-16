//go:build js

// Command cyber-pet draws the cyber pet companion on the browser canvas.
package main

import (
	"context"
	"log"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/cyber-pet/pet"
)

func main() {
	app, err := pet.New()
	if err != nil {
		log.Fatal(err)
	}

	if err := ownframe.Run(context.Background(), app.Page()); err != nil {
		log.Fatal(err)
	}
}

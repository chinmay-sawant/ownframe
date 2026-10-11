//go:build !js

// Command accessibility opens the accessibility and desktop input demo.
package main

import (
	"context"
	"log"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/accessibility/demo"
)

func main() {
	app, err := demo.New()
	if err != nil {
		log.Fatal(err)
	}
	if err := ownframe.Run(context.Background(), app.Page()); err != nil {
		log.Fatal(err)
	}
}

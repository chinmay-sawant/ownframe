//go:build js

package main

import (
	"context"
	"log"

	"github.com/chinmay-sawant/ownframe"
)

func main() {
	app, err := newApp(false, false)
	if err != nil {
		log.Fatal(err)
	}

	if err := ownframe.Run(context.Background(), app.Page()); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"context"
	"log"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/digital-wellbeing/dashboard"
)

func main() {
	app, err := dashboard.New()
	if err != nil {
		log.Fatal(err)
	}
	if err := ownframe.Run(context.Background(), app.Page()); err != nil {
		log.Fatal(err)
	}
}

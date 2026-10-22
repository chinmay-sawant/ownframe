//go:build !js

package main

import (
	"context"
	"errors"
	"flag"
	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/pet-care/app"
	"github.com/chinmay-sawant/ownframe/examples/pet-care/store"
	"log"
)

func main() {
	path := flag.String("db", "", "SQLite file (defaults to the user config directory)")
	flag.Parse()
	dbPath := *path
	if dbPath == "" {
		var err error
		dbPath, err = store.DefaultPath()
		if err != nil {
			log.Fatal(err)
		}
	}
	a, err := app.New(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	if err := ownframe.Run(context.Background(), a.Page()); err != nil {
		log.Fatal(errors.Join(err, a.Close()))
	}
	if err := a.Close(); err != nil {
		log.Fatal(err)
	}
}

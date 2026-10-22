//go:build !js

// Command download-manager opens the download queue example in a native
// window. Dummy data is the default; -real uses the HTTP transport.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/download-manager/ui"
	"github.com/chinmay-sawant/ownframe/examples/download-manager/ui/wire"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	dataDir := flag.String("data", "", "storage directory; empty uses the per-user default")
	real := flag.Bool("real", false, "use the real HTTP transport instead of dummy data")
	fixture := flag.Bool("fixture", false, "start the local HTTP fixture service and log its URLs")
	flag.Parse()

	ctx := context.Background()
	backend, err := wire.New(ctx, wire.Config{
		DataDir: *dataDir,
		Dummy:   !*real,
		Fixture: *fixture,
	})
	if err != nil {
		return err
	}

	screen, err := ui.New(ctx, backend)
	if err != nil {
		_ = backend.Close()
		return err
	}

	defer func() {
		if err := screen.Close(); err != nil {
			log.Printf("shutdown: %v", err)
		}
	}()

	if err := ownframe.Run(ctx, screen.Page()); err != nil {
		return err
	}

	return nil
}

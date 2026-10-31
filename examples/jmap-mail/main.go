//go:build !js

// Command jmap-mail showcases the public OwnFrame UI builder with a local,
// JMAP-inspired mail client. No account or JMAP endpoint is required.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/chinmay-sawant/ownframe"
)

func main() {
	webMode := flag.Bool("web", false, "serve a browser preview instead of opening a native window")
	addr := flag.String("addr", "127.0.0.1:8132", "address for -web")
	flag.Parse()

	if err := run(*webMode, *addr); err != nil {
		log.Fatal(err)
	}
}

func run(webMode bool, addr string) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	client, err := newApp()
	if err != nil {
		return err
	}
	if err := client.Page().Redraw(ctx); err != nil {
		return fmt.Errorf("redraw mail view: %w", err)
	}

	if webMode {
		if err := ownframe.Serve(ctx, client.Page(), addr); err != nil {
			return fmt.Errorf("serve mail preview: %w", err)
		}
		return nil
	}
	// Use OwnFrame's native drag region and custom chrome, not JavaScript.
	// The toolbar is the first 78 CSS pixels; clicks on its controls still work.
	options := ownframe.WindowOptions{
		CustomChrome: true,
		PlacementID:  mailPlacement,
		Draggable: func(x, y int) bool {
			return x >= 0 && y >= 0 && y < mailDragLimitY
		},
	}
	if err := ownframe.RunWithOptions(ctx, client.Page(), options); err != nil {
		return fmt.Errorf("run mail window: %w", err)
	}
	return nil
}

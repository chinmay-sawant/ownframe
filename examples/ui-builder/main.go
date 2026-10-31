// Command ui-builder showcases interactive OwnFrame components and an
// illustrative system-monitor dashboard. Values are local fixtures, not
// measurements collected from the user's device.
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
	web := flag.Bool("web", false, "serve the preview in a browser")
	addr := flag.String("addr", "127.0.0.1:8131", "address for -web")
	flag.Parse()

	if err := run(*web, *addr); err != nil {
		log.Fatal(err)
	}
}

func run(webMode bool, addr string) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app, err := newMonitorApp()
	if err != nil {
		return err
	}
	if err := app.page.Redraw(ctx); err != nil {
		return fmt.Errorf("redraw builder dashboard: %w", err)
	}
	if webMode {
		if err := ownframe.Serve(ctx, app.page, addr); err != nil {
			return fmt.Errorf("serve builder preview: %w", err)
		}
		return nil
	}
	// Native drag is handled by OwnFrame, not by injected JS or Chromium.
	options := ownframe.WindowOptions{
		CustomChrome: true,
		PlacementID:  builderPlacement,
		Draggable: func(x, y int) bool {
			return x >= 0 && y >= 0 && y < builderDragLimitY
		},
	}
	if err := ownframe.RunWithOptions(ctx, app.page, options); err != nil {
		return fmt.Errorf("run builder window: %w", err)
	}
	return nil
}

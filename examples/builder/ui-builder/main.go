// Command ui-builder is the interactive OwnFrame component-library gallery.
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
	dev := flag.Bool("dev", false, "open the developer tools dock at startup (dock sampling also feeds -debug-addr)")
	perf := flag.Bool("perf", false, "enable perf sampling for the devtools Frame tab and the -debug-addr endpoints")
	debugAddr := flag.String("debug-addr", "", "serve window profiling on this loopback address (off by default; needs -tags ownframe_debug)")
	logEnabled := flag.Bool("log-enabled", false, "enable diagnostic logging to stderr and file")
	logPath := flag.String(
		"log",
		"",
		"diagnostic log file for -log-enabled (default os.UserCacheDir()/ownframe/ui-builder/ui-builder.log)",
	)
	flag.Parse()

	logging := builderLogOptions{enabled: *logEnabled, path: *logPath}
	if err := run(*web, *addr, *dev, *perf, *debugAddr, logging); err != nil {
		log.Fatal(err)
	}
}

func run(webMode bool, addr string, dev bool, perf bool, debugAddr string, logging builderLogOptions) error {
	logger, file, actualPath := openBuilderLog(logging)
	if file != nil {
		defer func() { _ = file.Close() }()
	}
	mode := "native"
	if webMode {
		mode = "web"
	}
	logStartup(logger, actualPath, mode, addr)
	builderLogf(logger, "startup gallery dev=%v placement=%q", dev, builderPlacement)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app, err := newBuilderApp()
	if err != nil {
		builderLogf(logger, "startup new app err=%v", err)
		return err
	}
	app.setLogger(logger)
	app.page.Window().SetDevTools(dev)
	app.page.Window().SetPerf(perf)
	if err := app.page.Redraw(ctx); err != nil {
		builderLogf(logger, "startup redraw err=%v", err)
		return fmt.Errorf("redraw builder gallery: %w", err)
	}
	if err := app.refresh(ctx); err != nil {
		builderLogf(logger, "startup refresh err=%v", err)
		return fmt.Errorf("reload builder gallery: %w", err)
	}
	if webMode {
		builderLogf(logger, "serve addr=%s", addr)
		if err := ownframe.Serve(ctx, app.page, addr); err != nil {
			builderLogf(logger, "serve exit err=%v", err)
			return fmt.Errorf("serve builder preview: %w", err)
		}
		builderLogf(logger, "serve exit ok")
		return nil
	}
	// Native drag is handled by OwnFrame, not by injected JS or Chromium.
	options := ownframe.WindowOptions{
		CustomChrome: true,
		PlacementID:  builderPlacement,
		DebugAddr:    debugAddr,
		Draggable: func(x, y int) bool {
			return x >= 0 && y >= 0 && y < builderDragLimitY
		},
	}
	if err := ownframe.RunWithOptions(ctx, app.page, options); err != nil {
		builderLogf(logger, "run exit err=%v", err)
		return fmt.Errorf("run builder window: %w", err)
	}
	builderLogf(logger, "run exit ok")
	return nil
}

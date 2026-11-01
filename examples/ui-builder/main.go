// Command ui-builder showcases the Go ui builder across eight small tabs:
// widgets, settings, layout, todo, list, theme, builder-vs-HTML, and a
// backdrop dialog.
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
	dev := flag.Bool("dev", true, "open the developer tools dock at startup")
	logPath := flag.String("log", "", "diagnostic log file (default os.UserCacheDir()/ownframe/ui-builder/ui-builder.log)")
	flag.Parse()

	if err := run(*web, *addr, *dev, *logPath); err != nil {
		log.Fatal(err)
	}
}

func run(webMode bool, addr string, dev bool, logPath string) error {
	logger, file, actualPath := openBuilderLog(logPath)
	if file != nil {
		defer func() { _ = file.Close() }()
	}
	mode := "native"
	if webMode {
		mode = "web"
	}
	logStartup(logger, actualPath, mode, addr)
	logger.Printf("startup gallery dev=%v placement=%q", dev, builderPlacement)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app, err := newBuilderApp()
	if err != nil {
		logger.Printf("startup new app err=%v", err)
		return err
	}
	app.setLogger(logger)
	if err := app.page.Redraw(ctx); err != nil {
		logger.Printf("startup redraw err=%v", err)
		return fmt.Errorf("redraw builder gallery: %w", err)
	}
	if err := app.refresh(ctx); err != nil {
		logger.Printf("startup refresh err=%v", err)
		return fmt.Errorf("reload builder gallery: %w", err)
	}
	app.page.Window().SetDevTools(dev)
	if webMode {
		logger.Printf("serve addr=%s", addr)
		if err := ownframe.Serve(ctx, app.page, addr); err != nil {
			logger.Printf("serve exit err=%v", err)
			return fmt.Errorf("serve builder preview: %w", err)
		}
		logger.Printf("serve exit ok")
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
		logger.Printf("run exit err=%v", err)
		return fmt.Errorf("run builder window: %w", err)
	}
	logger.Printf("run exit ok")
	return nil
}

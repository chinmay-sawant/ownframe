//go:build !js

// Command code-view opens the Code Review example.
// Pass -web to serve the interactive picture in a browser instead.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/chinmay-sawant/ownframe"
	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	webMode := flag.Bool("web", false, "serve the picture in a browser on -addr")
	addr := flag.String("addr", "127.0.0.1:8145", "listen address for -web")
	root := flag.String("root", ".", "directory to browse; defaults to the current directory")
	files := flag.Int("files", maxWorkspaceFiles, "maximum files to scan; raise it for a large repository")
	dev := flag.Bool("dev", false, "open the developer tools dock at startup")
	perf := flag.Bool("perf", true, "enable pipeline sampling for the devtools Frame tab")
	chrome := flag.Bool("chrome", true, "draw custom minimize, maximize, and close controls in the titlebar")
	addWSLFlags(flag.CommandLine)
	flag.Parse()
	if maybeReexecInWSL() {
		return
	}

	app, err := NewWithLimit(*root, *files)
	if err != nil {
		log.Fatal(err)
	}

	app.EnableSession()

	devSet := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "dev" {
			devSet = true
		}
	})

	ctx := context.Background()
	app.reloadChanges(ctx, false)
	// The dock stays hidden by default. A restored session that remembers an
	// open dock must not reopen it unless the launch passed -dev=true;
	// F12/Ctrl+Shift+I and SetDevTools still toggle it at runtime.
	if devSet {
		app.SetDevTools(*dev)
	} else {
		app.SetDevTools(false)
	}

	if *perf {
		app.Page().Window().SetPerf(true)
	}

	if *webMode {
		if err := ownframe.ServeWithOptions(ctx, app.Page(), *addr, ownframe.ServeOptions{InsecureAllowPprof: *perf}); err != nil {
			log.Fatal(err)
		}

		return
	}

	// Custom chrome is a desktop window only. The web preview keeps the
	// native frame and hides the in-page controls.
	app.chrome = *chrome

	if err := app.StartWatcher(ctx); err != nil {
		log.Printf("code-view: watcher unavailable: %v", err)
	}
	options := ownframe.WindowOptions{CustomChrome: app.chrome, PlacementID: "code-view"}
	if app.chrome {
		options.Draggable = app.chromeDrag
	}
	runErr := ownframe.RunWithOptions(ctx, app.Page(), options)
	app.StopWatcher()

	if runErr != nil {
		log.Fatal(runErr)
	}
}

// chromeDrag moves a borderless code-view window by its titlebar. Ebiten
// hands the cursor in device pixels, so the strip scales with the display
// and the layout zoom. A press on a titlebar button still reaches the page
// as a click when it is released without moving.
func (a *App) chromeDrag(x, y int) bool {
	return float64(y) < float64(a.zx(chromeTop))*chromeDragScale()
}

// chromeDragScale reads the monitor scale for the drag strip. Without a
// monitor the strip falls back to one CSS pixel per device pixel.
func chromeDragScale() float64 {
	if monitor := ebiten.Monitor(); monitor != nil {
		return monitor.DeviceScaleFactor()
	}

	return 1
}

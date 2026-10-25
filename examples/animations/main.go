//go:build !js

// Command animations launches the CSS animation showcase gallery.
// Pass -web to serve the picture in a browser on -addr instead of a native window.
package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	webMode := flag.Bool("web", false, "serve the picture in a browser on -addr")
	addr := flag.String("addr", "127.0.0.1:8142", "listen address for -web")
	devtools := flag.Bool("dev", true, "open the developer tools dock (default true; use -dev=false to close)")
	performance := flag.Bool("performance", true, "record pipeline and frame timings for the DevTools Frame tab (default true; use -performance=false to measure without them)")
	profileSeconds := flag.Int("profile-seconds", 0, "log a performance summary after this many seconds of measurement, plus a 2s warmup, and close (0 runs until closed; desktop only)")
	width := flag.Int("width", 0, "window width (0 uses the example default)")
	height := flag.Int("height", 0, "window height (0 uses the example default)")
	monitor := flag.Int("monitor", -1, "monitor index to open on (-1 uses the current monitor)")
	vsync := flag.Bool("vsync", true, "wait for the display refresh; pass -vsync=false on a virtual display that holds a large window below its rate")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())

	app, err := newApp(*devtools, *performance)
	if err != nil {
		log.Fatal(err)
	}

	if *width > 0 && *height > 0 {
		app.Page().SetSize(*width, *height)
		if err := app.Page().Redraw(ctx); err != nil {
			log.Fatal(err)
		}
	}
	if *monitor >= 0 {
		if monitors := ebiten.AppendMonitors(nil); *monitor < len(monitors) {
			ebiten.SetMonitor(monitors[*monitor])
		}
	}

	if *webMode {
		if err := ownframe.Serve(ctx, app.Page(), *addr); err != nil {
			log.Fatal(err)
		}
		return
	}

	var profile <-chan profileResult
	if *profileSeconds > 0 {
		profile = startProfile(time.Duration(*profileSeconds)*time.Second, cancel)
	}

	if err := ownframe.RunWithOptions(ctx, app.Page(), ownframe.WindowOptions{NoVsync: !*vsync, Perf: *performance}); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatal(err)
	}
	cancel()

	if profile != nil {
		select {
		case result := <-profile:
			logProfileSummary(app.Page(), result.samples, result.elapsed, result.winW, result.winH)
		case <-time.After(profileDrain):
		}
	}
}

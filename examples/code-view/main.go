//go:build !js

// Command code-view opens the Code Review example.
// Pass -web to serve the interactive picture in a browser instead.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/chinmay-sawant/ownframe"
)

func main() {
	webMode := flag.Bool("web", false, "serve the picture in a browser on -addr")
	addr := flag.String("addr", "127.0.0.1:8145", "listen address for -web")
	root := flag.String("root", ".", "directory to browse; defaults to the current directory")
	files := flag.Int("files", maxWorkspaceFiles, "maximum files to scan; raise it for a large repository")
	dev := flag.Bool("dev", true, "open the developer tools dock at startup")
	perf := flag.Bool("perf", true, "enable pipeline sampling for the devtools Frame tab")
	flag.Parse()

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
	// The flag default is on. A saved session must not hide the dock unless
	// the launch passed -dev=false.
	if devSet || !app.sessionRestored || *dev {
		app.SetDevTools(*dev)
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

	if err := app.StartWatcher(ctx); err != nil {
		log.Printf("code-view: watcher unavailable: %v", err)
	}
	runErr := ownframe.Run(ctx, app.Page())
	app.StopWatcher()

	if runErr != nil {
		log.Fatal(runErr)
	}
}

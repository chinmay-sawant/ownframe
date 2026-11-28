//go:build !js

// Command codegen demonstrates compile-time HTML-to-Go code generation with
// cmd/ownframegen: a typed fluent view builder, zero-reflection Config.Render,
// and a compile-time-checked PatchBuilder for retained DOM updates.
package main

//go:generate go run ../../cmd/ownframegen -in view.html -out view_gen.go -pkg main -type DashboardView

import (
	"context"
	"flag"
	"log"

	"github.com/chinmay-sawant/ownframe"
)

func main() {
	webMode := flag.Bool("web", false, "serve the picture in a browser on -addr")
	addr := flag.String("addr", "127.0.0.1:8152", "listen address for -web")
	devtools := flag.Bool("dev", false, "open the developer tools dock at startup")
	perf := flag.Bool("perf", false, "record pipeline and frame timings")
	debugAddr := flag.String("debug-addr", "", "serve window profiling on this loopback address (needs -tags ownframe_debug)")
	flag.Parse()

	app, err := New(*devtools, *perf)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	if *webMode {
		if err := ownframe.ServeWithOptions(ctx, app.Page(), *addr, ownframe.ServeOptions{InsecureAllowPprof: *perf}); err != nil {
			log.Fatal(err)
		}

		return
	}

	if err := ownframe.RunWithOptions(ctx, app.Page(), ownframe.WindowOptions{
		Perf:      *perf,
		DebugAddr: *debugAddr,
	}); err != nil {
		log.Fatal(err)
	}
}

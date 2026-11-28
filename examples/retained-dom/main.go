//go:build !js

// Command retained-dom visualizes the OwnFrame v0.0.9 features in one screen:
// retained DOM mutations (PatchElement / PatchElements with class, attribute,
// text, AppendHTML, InnerHTML, and Remove), in-place CSS motion fast paths
// (border-radius, color alpha, individual translate/rotate/scale, planar
// translate3d/scale3d), and clipped retained layers beside blend siblings.
// DevTools and Perf sampling are enabled by default.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/chinmay-sawant/ownframe"
)

func main() {
	webMode := flag.Bool("web", false, "serve the picture in a browser on -addr")
	addr := flag.String("addr", "127.0.0.1:8151", "listen address for -web")
	devtools := flag.Bool("dev", true, "open the developer tools dock at startup (use -dev=false to hide)")
	perf := flag.Bool("perf", true, "record pipeline and frame timings for DevTools and -debug-addr")
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

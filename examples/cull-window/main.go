//go:build !js

// Command cull-window demonstrates opt-in viewport culling (Config.CullWindow)
// on a 1,000-row fixed-height HTML/CSS list. Pass -cull=false to compare
// against unculled layout, -dev/-perf to control DevTools and frame timings,
// or -web to serve in a browser.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/chinmay-sawant/ownframe"
)

func main() {
	webMode := flag.Bool("web", false, "serve the picture in a browser on -addr")
	addr := flag.String("addr", "127.0.0.1:8124", "listen address for -web")
	cull := flag.Bool("cull", true, "enable Config.CullWindow viewport culling (use -cull=false to compare full layout)")
	devtools := flag.Bool("dev", true, "open the developer tools dock (use -dev=false to hide)")
	perf := flag.Bool("perf", true, "record pipeline and frame timings for DevTools and -debug-addr")
	debugAddr := flag.String("debug-addr", "", "serve window profiling on this loopback address (needs -tags ownframe_debug)")
	flag.Parse()

	p, err := newCullWindowPage(*cull, *devtools, *perf)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	if *webMode {
		if err := ownframe.Serve(ctx, p, *addr); err != nil {
			log.Fatal(err)
		}

		return
	}

	if err := ownframe.RunWithOptions(ctx, p, ownframe.WindowOptions{
		Perf:      *perf,
		DebugAddr: *debugAddr,
	}); err != nil {
		log.Fatal(err)
	}
}

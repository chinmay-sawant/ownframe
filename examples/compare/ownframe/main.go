// Command ownframe opens the compare thousand-row table built with the UI builder.
// Pass -web to serve the picture in a browser instead.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/chinmay-sawant/ownframe"
)

func main() {
	web := flag.Bool("web", false, "serve the picture in a browser on -addr")
	addr := flag.String("addr", "127.0.0.1:8146", "listen address for -web")
	dev := flag.Bool("dev", false, "open the developer tools dock at startup")
	perf := flag.Bool("perf", false, "enable perf sampling for the devtools Frame tab")
	debugAddr := flag.String("debug-addr", "", "serve window profiling on this loopback address (off by default; needs -tags ownframe_debug)")
	flag.Parse()

	app, err := newListApp()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	app.Page().Window().SetDevTools(*dev)
	app.Page().Window().SetPerf(*perf)
	if err := app.Page().Redraw(ctx); err != nil {
		log.Fatal(err)
	}

	if *web {
		if err := ownframe.Serve(ctx, app.Page(), *addr); err != nil {
			log.Fatal(err)
		}
		return
	}

	if *debugAddr != "" {
		if err := ownframe.RunWithOptions(ctx, app.Page(), ownframe.WindowOptions{DebugAddr: *debugAddr}); err != nil {
			log.Fatal(err)
		}
		return
	}

	if err := ownframe.Run(ctx, app.Page()); err != nil {
		log.Fatal(err)
	}
}

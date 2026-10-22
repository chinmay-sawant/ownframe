// Command perf-complex opens the complex performance baseline dashboard.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/chinmay-sawant/ownframe"
)

func main() {
	dump := flag.Bool("dump", false, "capture headless profiles instead of opening a window")
	mode := flag.String("mode", "cached", "dump scenario: initial, cached, data, resize, windowed")
	out := flag.String("out", "temp/complex-dump", "dump directory")
	web := flag.Bool("web", false, "serve the picture on -addr")
	addr := flag.String("addr", "127.0.0.1:8135", "web address")
	windowed := flag.Bool("windowed", true, "lay out visible rows with overscan")
	flag.Parse()
	if *dump {
		if err := run(*mode, *out); err != nil {
			log.Fatal(err)
		}
		return
	}
	p, v, err := fixture(false)
	if err != nil {
		log.Fatal(err)
	}
	if *windowed {
		installWindow(p, v)
	}
	if err := installHover(p); err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	if *web {
		if err := ownframe.Serve(ctx, p, *addr); err != nil {
			log.Fatal(err)
		}
		return
	}
	if err := ownframe.Run(ctx, p); err != nil {
		log.Fatal(err)
	}
}

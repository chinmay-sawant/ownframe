//go:build !js

// Command large opens the Benchmark B thousand-row list in a native window.
// Pass -web to serve the picture in a browser instead.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/chinmay-sawant/ownframe/examples/perf-benchmarks/benchutil"
)

func main() {
	web, addr := benchutil.Flags("127.0.0.1:8132")
	flag.Parse()

	// rowCount is the Benchmark B list length.
	const rowCount = 1000

	app, err := benchutil.NewLarge(rowCount)
	if err != nil {
		log.Fatal(err)
	}

	if err := benchutil.ServeOrRun(context.Background(), app.Page(), *web, *addr); err != nil {
		log.Fatal(err)
	}
}

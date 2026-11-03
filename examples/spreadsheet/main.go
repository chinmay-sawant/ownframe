// Command spreadsheet opens the spreadsheet example in a native window.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/ui"
	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/ui/corebackend"
)

func main() {
	data := flag.String("data", "", "data directory override (default os.UserConfigDir()/ownframe/spreadsheet)")
	book := flag.String("workbook", "", "workbook name to open (default the first seeded workbook)")
	perf := flag.Bool("perf", false, "record pipeline timing for the devtools Frame tab and -debug-addr")
	debugAddr := flag.String("debug-addr", "", "serve window profiling on this loopback address (off by default; needs -tags ownframe_debug)")
	flag.Parse()

	ctx := context.Background()
	backend, err := corebackend.Open(*data)
	if err != nil {
		log.Fatal(err)
	}

	if err := backend.Seed(ctx, *book); err != nil {
		_ = backend.Close()
		log.Fatal(err)
	}

	app, err := ui.New(ui.Options{Backend: backend, DataDir: backend.Dir(), Perf: *perf})
	if err != nil {
		_ = backend.Close()
		log.Fatal(err)
	}

	err = ownframe.RunWithOptions(ctx, app.Page(), ownframe.WindowOptions{DebugAddr: *debugAddr})
	if cerr := app.Close(); cerr != nil && err == nil {
		err = cerr
	}

	if err != nil {
		log.Fatal(err)
	}
}

// Command excel is a spreadsheet built with the ownframe builder pattern:
// a virtualized grid, a formula engine, a ribbon and dialogs, with no
// browser engine behind it.
package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

	"github.com/chinmay-sawant/ownframe"
)

func main() {
	web := flag.Bool("web", false, "serve the spreadsheet in a browser preview")
	webOp := flag.Bool("web-op", false, "same as -web")
	headless := flag.Bool("headless", false, "serve without opening a window (same as -web)")
	addr := flag.String("addr", "127.0.0.1:8149", "address for -web or -headless")
	dev := flag.Bool("dev", false, "open the developer tools dock at startup")
	performance := flag.Bool("performance", false, "enable perf sampling and the -web pprof endpoints")
	perf := flag.Bool("perf", false, "same as -performance")
	debugAddr := flag.String("debug-addr", "", "serve window profiling on this loopback address (needs -tags ownframe_debug)")
	flag.Parse()
	err := run(options{
		web:       *web || *webOp || *headless,
		addr:      *addr,
		dev:       *dev,
		perf:      *performance || *perf,
		debugAddr: *debugAddr,
	})
	if err != nil {
		log.Fatal(err)
	}
}

type options struct {
	web       bool
	addr      string
	dev, perf bool
	debugAddr string
}

func run(opt options) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	a, err := newApp()
	if err != nil {
		return err
	}
	a.web = opt.web
	a.page.Window().SetDevTools(opt.dev)
	a.page.Window().SetPerf(opt.perf)
	if opt.web {
		// The web preview keeps the browser frame, so the in-page window
		// controls are hidden.
		a.redraw()
		return ownframe.ServeWithOptions(ctx, a.page, opt.addr, ownframe.ServeOptions{InsecureAllowPprof: opt.perf})
	}
	if err := a.page.Redraw(ctx); err != nil {
		return err
	}
	return ownframe.RunWithOptions(ctx, a.page, ownframe.WindowOptions{
		CustomHeader: true,
		PlacementID:  excelPlacement,
		DebugAddr:    opt.debugAddr,
		Draggable: func(x, y int) bool {
			// The strip yields while an overlay is open so presses on a
			// dialog's padding reach the dialog instead of moving the window.
			if a.overlaysOpen() {
				return false
			}
			return x >= 0 && y >= 0 && y < titleBarHeight
		},
	})
}

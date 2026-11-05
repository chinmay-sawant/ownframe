package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

	"github.com/chinmay-sawant/ownframe"
)

const (
	inventoryTopbarHeight = 60
	inventoryPlacement    = "inventory"
)

func main() {
	web := flag.Bool("web", false, "serve the dashboard in a browser preview")
	headless := flag.Bool("headless", false, "serve without opening a window (same as -web)")
	addr := flag.String("addr", "127.0.0.1:8135", "address for -web or -headless")
	dev := flag.Bool("dev", false, "open the developer tools dock at startup (dock sampling also feeds -debug-addr)")
	perf := flag.Bool("perf", false, "enable perf sampling for the devtools Frame tab and the -debug-addr endpoints")
	debugAddr := flag.String("debug-addr", "", "serve window profiling on this loopback address (off by default; needs -tags ownframe_debug)")
	flag.Parse()
	if err := run(*web || *headless, *addr, *dev, *perf, *debugAddr); err != nil {
		log.Fatal(err)
	}
}

func run(web bool, addr string, dev bool, perf bool, debugAddr string) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	a, err := newApp()
	if err != nil {
		return err
	}
	a.page.Window().SetDevTools(dev)
	a.page.Window().SetPerf(perf)
	if err := a.page.Redraw(ctx); err != nil {
		return err
	}
	if web {
		return ownframe.Serve(ctx, a.page, addr)
	}
	// Match the ui-builder chrome: no OS title bar, the page draws its own
	// controls. The header is normal flow in the 100vh flex shell (static,
	// not an overlay); the drag strip is unchanged at y < inventoryTopbarHeight.
	// Controls win over drags because OwnFrame skips a drag that starts on
	// a button.
	options := ownframe.WindowOptions{
		CustomChrome: true,
		PlacementID:  inventoryPlacement,
		DebugAddr:    debugAddr,
		Draggable: func(x, y int) bool {
			// While an overlay is open (search palette, notification or
			// account popover, PO dialog), the strip under it must not
			// start a window drag: presses on overlay padding carry no
			// action, so without this guard a click there would move the
			// window instead of reaching the overlay.
			if a.overlaysOpen() {
				return false
			}
			return x >= 0 && y >= 0 && y < inventoryTopbarHeight
		},
	}
	return ownframe.RunWithOptions(ctx, a.page, options)
}

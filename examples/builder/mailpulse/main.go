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
	// topbarDragLimitY is the draggable strip height for the custom chrome.
	// It matches the .mp-topbar box: 14px padding + 36px controls + 14px.
	topbarDragLimitY = 64
)

func main() {
	web := flag.Bool("web", false, "serve the dashboard in a browser preview")
	headless := flag.Bool("headless", false, "serve without opening a window (same as -web)")
	addr := flag.String("addr", "127.0.0.1:8148", "address for -web or -headless")
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
	// Match the other builder apps: no OS title bar, the page draws its own
	// chrome, and the top bar is the drag region. Controls win over drags
	// because OwnFrame skips a drag that starts on a button.
	options := ownframe.WindowOptions{
		CustomChrome: true,
		PlacementID:  "mailpulse",
		DebugAddr:    debugAddr,
		Draggable: func(x, y int) bool {
			return x >= 0 && y >= 0 && y < topbarDragLimitY
		},
	}
	return ownframe.RunWithOptions(ctx, a.page, options)
}

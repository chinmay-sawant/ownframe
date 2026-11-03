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
	web := flag.Bool("web", false, "serve the dashboard in a browser preview")
	headless := flag.Bool("headless", false, "serve without opening a window (same as -web)")
	addr := flag.String("addr", "127.0.0.1:8132", "address for -web or -headless")
	dev := flag.Bool("dev", false, "open the developer tools dock at startup")
	flag.Parse()
	if err := run(*web || *headless, *addr, *dev); err != nil {
		log.Fatal(err)
	}
}

func run(web bool, addr string, dev bool) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	a, err := newApp()
	if err != nil {
		return err
	}
	a.page.Window().SetDevTools(dev)
	if err := a.page.Redraw(ctx); err != nil {
		return err
	}
	if web {
		return ownframe.Serve(ctx, a.page, addr)
	}
	// Match the ui-builder chrome: no OS title bar, the page draws its own
	// controls, and the top bar is the drag region. Controls win over drags
	// because OwnFrame skips a drag that starts on a button.
	options := ownframe.WindowOptions{
		CustomChrome: true,
		PlacementID:  "loglens",
		Draggable: func(x, y int) bool {
			return x >= 0 && y >= 0 && y < topbarHeight
		},
	}
	return ownframe.RunWithOptions(ctx, a.page, options)
}

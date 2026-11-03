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
	addr := flag.String("addr", "127.0.0.1:8132", "address for -web")
	dev := flag.Bool("dev", false, "open the developer tools dock at startup")
	flag.Parse()
	if err := run(*web, *addr, *dev); err != nil {
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
	return ownframe.Run(ctx, a.page)
}

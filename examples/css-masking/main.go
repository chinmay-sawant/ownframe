//go:build !js

package main

import (
	"context"
	"flag"
	"log"

	"github.com/chinmay-sawant/ownframe"
)

func main() {
	web := flag.Bool("web", false, "serve the picture in a browser")
	addr := flag.String("addr", "127.0.0.1:8139", "listen address for -web")
	flag.Parse()
	p, err := newPage()
	if err != nil {
		log.Fatal(err)
	}
	if *web {
		if err := ownframe.Serve(context.Background(), p, *addr); err != nil {
			log.Fatal(err)
		}
		return
	}
	if err := ownframe.Run(context.Background(), p); err != nil {
		log.Fatal(err)
	}
}

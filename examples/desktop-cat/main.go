//go:build !js && !android && !ios

// Command desktop-cat opens a transparent, click-through desktop cat.
package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/chinmay-sawant/ownframe/examples/desktop-cat/cat"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	o := flags()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if o.snapshot {
		o.notifyAddr = "off"
	}
	inbox := cat.NewInbox()
	closeServer, err := startNotifications(ctx, o.notifyAddr, inbox)
	if err != nil {
		return err
	}
	defer closeServer()
	if o.stdin {
		go readNotifications(inbox)
	}

	if !o.web {
		handled, err := launchNative(ctx, inbox, o.notifyAddr)
		if handled {
			if err != nil && !errors.Is(err, context.Canceled) {
				return err
			}

			return nil
		}
	}

	if o.snapshot {
		printMedia(ctx)
		return nil
	}
	if o.media {
		startMedia(ctx, inbox, o.mediaEndpoint)
	}
	companion, err := cat.NewWithInbox(o.variant, inbox)
	if err != nil {
		return err
	}

	companion.RandomBehavior = o.random
	if o.web {
		err = companion.ServePreview(ctx, o.addr)
	} else {
		err = runOverlay(ctx, companion, o.margin)
	}

	if err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}

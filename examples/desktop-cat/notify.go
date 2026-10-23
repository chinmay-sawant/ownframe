//go:build !js && !android && !ios

package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/chinmay-sawant/ownframe/examples/desktop-cat/cat"
)

// startNotifications serves the loopback agent API. The default "auto"
// listens on a 0600 Unix socket; a tcp address is explicit opt-in and
// non-loopback binds log a warning: add auth before exposing them.
func startNotifications(ctx context.Context, addr string, inbox *cat.Inbox) (func(), error) {
	if addr == "off" {
		return func() {}, nil
	}
	listener, err := listenNotify(ctx, addr)
	if err != nil {
		return nil, err
	}
	server := &http.Server{Handler: inbox.Handler(), ReadHeaderTimeout: notifyReadHeaderTimeout, ReadTimeout: notifyReadTimeout, WriteTimeout: notifyWriteTimeout, IdleTimeout: notifyIdleTimeout, MaxHeaderBytes: maxNotifyHeaderBytes}
	goSafe("notify serve", func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Print(err)
		}
	})
	log.Printf("desktop cat notifications: %s://%s", listener.Addr().Network(), listener.Addr())
	stop := func() { _ = server.Close() }
	goSafe("notify stop", func() { <-ctx.Done(); stop() })
	return stop, nil
}

func readNotifications(inbox *cat.Inbox) {
	decoder := json.NewDecoder(os.Stdin)
	for {
		var n cat.Notification
		if err := decoder.Decode(&n); err != nil {
			return
		}
		if err := inbox.Submit(n); err != nil {
			log.Print(err)
		}
	}
}

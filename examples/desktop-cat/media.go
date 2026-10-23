//go:build !js && !android && !ios

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/desktop-cat/cat"
	"github.com/chinmay-sawant/ownframe/examples/desktop-cat/nowplaying"
)

func startMedia(ctx context.Context, inbox *cat.Inbox, endpoint string) {
	goSafe("media watch", func() {
		publish := func(track nowplaying.Track) error {
			text := []rune("Now playing: " + track.Title)
			if len(text) > mediaTextLimit {
				text = text[:mediaTextLimit]
			}
			n := cat.Notification{Expression: "happy", Message: string(text), Source: "Chrome"}
			if endpoint != "" {
				return postMedia(ctx, endpoint, n)
			}
			return inbox.Submit(n)
		}
		lastError := ""
		for {
			err := nowplaying.Watch(ctx, publish)
			if ctx.Err() != nil || errors.Is(err, nowplaying.ErrUnsupported) {
				return
			}
			if err != nil && err.Error() != lastError {
				lastError = err.Error()
				log.Printf("Chrome media monitor: %v; retrying", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(mediaRetryInterval):
			}
		}
	})
}

func printMedia(ctx context.Context) {
	track, err := nowplaying.Snapshot(ctx)
	if err != nil {
		log.Print(err)
		return
	}
	data, _ := json.Marshal(track)
	fmt.Println(string(data))
}

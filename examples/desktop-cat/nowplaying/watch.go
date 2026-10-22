// Package nowplaying reads Chrome's native system media session.
package nowplaying

import (
	"context"
	"errors"
	"runtime"
	"time"
)

var ErrUnsupported = errors.New("native media monitoring is available on Windows only")

type Track struct {
	Title   string `json:"title"`
	Artist  string `json:"artist,omitempty"`
	App     string `json:"app,omitempty"`
	Playing bool   `json:"playing"`
}

type reader interface {
	read(context.Context) (Track, error)
	close()
}

// Snapshot reads media metadata without opening a window or changing playback.
func Snapshot(ctx context.Context) (Track, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	r, err := open(ctx)
	if err != nil {
		return Track{}, err
	}
	defer r.close()
	return r.read(ctx)
}

// Watch publishes settled, changed titles. Pausing does not repeat a title.
func Watch(ctx context.Context, publish func(Track) error) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	r, err := open(ctx)
	if err != nil {
		return err
	}
	defer r.close()
	ticker := time.NewTicker(watchInterval)
	defer ticker.Stop()
	var state tracker
	for {
		track, err := r.read(ctx)
		if err != nil {
			return err
		}
		if key := state.next(track); key != "" {
			if err := publish(track); err == nil {
				state.last = key
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

package main

import (
	"context"
	"time"
)

// partialLimit is the most samples whose damage still fits the window's
// eight-rect partial replay. Past it the union covers the page, and a full
// replay is cheaper than a tile rebuild.
const partialLimit = 8

func (d *demo) tick(ctx context.Context) error {
	if len(d.samples) == 0 && len(d.pending) == 0 {
		d.p.UseFrameDirty()
		return nil
	}
	if err := d.capturePending(ctx); err != nil {
		return err
	}
	if len(d.pending) > 0 && time.Since(d.pendingAt) > 40*time.Millisecond {
		d.pendingAt = time.Now()
		pending := d.pending
		d.pending = nil
		boxes := d.p.Boxes()
		for _, key := range pending {
			started, err := d.begin(key, boxes)
			if err != nil {
				return err
			}
			if !started {
				d.pending = append(d.pending, key)
			}
		}
		if err := d.rebind(); err != nil {
			return err
		}
		d.gen = d.p.Generation()
	} else if d.gen != d.p.Generation() {
		if err := d.rebind(); err != nil {
			return err
		}
		d.gen = d.p.Generation()
	}
	active := 0
	for _, s := range d.samples {
		if s.active {
			active++
		}
	}
	partial := active > 0 && active <= partialLimit
	if partial || active == 0 {
		d.p.UseFrameDirty()
	}
	now := time.Now()
	live, waiting := d.stepSamples(now, partial)
	switch {
	case live:
		d.p.WantFrameIn(0)
	case waiting:
		d.p.WantFrameIn(50 * time.Millisecond)
	case len(d.pending) > 0:
		// A queued row waits for its first render. Keep the tick so the
		// row that scrolls in captures its starting paint.
		d.p.WantFrameIn(50 * time.Millisecond)
	default:
		d.p.SetTick(nil)
	}
	return nil
}

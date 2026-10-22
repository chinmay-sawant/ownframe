package main

import (
	"context"
	"time"
)

func (d *demo) tick(ctx context.Context) error {
	if len(d.samples) == 0 && len(d.pending) == 0 && !d.treeDirty && !d.remeasure {
		d.p.UseFrameDirty()
		return nil
	}
	if err := d.capturePending(ctx); err != nil {
		return err
	}
	if d.treeDirty {
		// The tree carries state the engine refused to splice. A layout
		// shows it, and the sample binds against the laid out page.
		d.treeDirty = false
		if err := d.p.Redraw(ctx); err != nil {
			return err
		}
	}
	if d.remeasure && time.Since(d.resizeAt) > resizeSettle {
		d.remeasure = false
		if err := d.remeasureCatalog(ctx); err != nil {
			return err
		}
	}
	if len(d.pending) > 0 && time.Since(d.pendingAt) > 40*time.Millisecond {
		d.pendingAt = time.Now()
		if err := d.drain(); err != nil {
			return err
		}
		d.gen = d.p.Generation()
	} else if d.gen != d.p.Generation() {
		if err := d.rebind(); err != nil {
			return err
		}
		d.gen = d.p.Generation()
	}
	active := d.activeCount()
	partial := active > 0 && active <= partialLimit
	if partial || active == 0 {
		d.p.UseFrameDirty()
	}
	now := time.Now()
	live, waiting := d.stepSamples(now, partial)
	d.endRun(ctx, live)
	switch {
	case live:
		d.p.WantFrameIn(0)
	case waiting:
		d.p.WantFrameIn(frameWait)
	case d.remeasure:
		d.p.WantFrameIn(remeasureWait)
	case len(d.pending) > 0:
		// A queued row waits for its first render. The scroll window
		// queues it again when the row comes into view.
		d.p.WantFrameIn(frameWait)
	default:
		d.p.SetTick(nil)
	}
	return nil
}

// activeCount counts the samples that still run.
func (d *demo) activeCount() int {
	active := 0
	for _, s := range d.samples {
		if s.active {
			active++
		}
	}

	return active
}

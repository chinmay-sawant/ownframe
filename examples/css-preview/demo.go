package main

import (
	"context"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// demo is the application state around the page. CSS animations run on the
// page's own driver, so this type keeps only the view data and the temporary
// manual paths that the runtime does not cover yet: the geometry CSS
// transitions and the three trigger cards whose started effect does not
// survive the publish that carries the click's new class.
type demo struct {
	p *ownframe.Page
	v *view
	// readout is the text the current draw carries. The walker reset clears
	// it with the view field it mirrors.
	readout string
	// window is the row window that slices the catalog.
	window *catalogWindow
	// sizeW and sizeH are the page size the geometry was measured at.
	sizeW, sizeH int
	// remeasure marks that the window size changed, so the catalog has to
	// be measured again.
	remeasure bool
	// resizeAt is when the window size last changed.
	resizeAt time.Time
	// manual holds the previews that still step by hand. Their file has the
	// reason and the removal condition.
	manual map[string]*manualSample
	// startedAt and startedSeq are the click markers the walker's HTTP state
	// route reports. scripts/css-preview-click-test.py reads them by
	// position, so they stay even though the app no longer samples in Go.
	startedAt  time.Time
	startedSeq uint64
}

func newDemo(p *ownframe.Page, v *view) *demo {
	w, h := p.Size()

	return &demo{p: p, v: v, manual: map[string]*manualSample{}, sizeW: w, sizeH: h}
}

// tick runs once per frame only while the app has work of its own: a manual
// transition or trigger sample, or a catalog that has to be measured again.
// CSS animations are sampled by the page before this callback.
func (d *demo) tick(ctx context.Context) error {
	if d.remeasure && time.Since(d.resizeAt) > resizeSettle {
		d.remeasure = false
		if err := d.remeasureCatalog(ctx); err != nil {
			return err
		}
	}
	// A publish during the tick replaces the operations a manual player
	// points at. Rebind before stepping so Play all and restarts keep writing
	// to the live display.
	if err := d.bindManual(d.v.Rows); err != nil {
		return err
	}
	live := d.stepManual(time.Now())
	switch {
	case live:
		d.p.WantFrameIn(0)
	case d.remeasure:
		d.p.WantFrameIn(remeasureWait)
	default:
		d.p.SetTick(nil)
	}
	return nil
}

// wake arms the tick when the app has work for the frame loop.
func (d *demo) wake() {
	if walkActive() {
		return
	}
	if d.manualLive() || d.remeasure {
		d.p.SetTick(d.tick)
	}
}

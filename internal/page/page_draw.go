package page

import (
	"context"
	"runtime"
	"strings"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

// Redraw renders the current size: a display list for a replayable page,
// a bitmap otherwise. Parsed trees and sheets are reused. Stage timing and
// allocs track only with Perf on.
func (p *Page) Redraw(ctx context.Context) error {
	if err := useContext(ctx); err != nil {
		return err
	}

	track := p.perf
	var m0 runtime.MemStats
	if track {
		m0 = allocStart()
	} else {
		p.clearPerf()
	}

	start := time.Now()
	p.stats.redraws++
	defer func() {
		p.stats.lastRedraw = time.Since(start)
		if track {
			p.stats.allocFrame = allocUsed(m0)
		}
	}()

	ts := time.Now()
	p.applyScrollWindow()
	var body strings.Builder
	if err := p.tpl.Execute(&body, p.data); err != nil {
		return err
	}
	if track {
		p.stats.lastTemplate = time.Since(ts)
	}

	p.source = p.syncForm(body.String())
	p.png = nil

	state := p.renderState()

	ls := time.Now()
	styled, err := p.styledDocument(ctx, p.source, state)
	if track {
		p.stats.layoutTime = time.Since(ls)
	}
	if err != nil {
		return err
	}

	drawStart := time.Now()

	ds := time.Now()
	display, derr := render.DisplayListDocument(ctx, styled, state.Images)
	if track {
		p.stats.displayListTime = time.Since(ds)
	}
	if derr != nil {
		return derr
	}

	if render.Replayable(display) {
		p.dirtyFromDisplay(p.display, display)
		p.stats.layouts++
		p.stats.repaints++
		p.stats.lastDraw = time.Since(drawStart)
		p.img = nil
		p.display = display
		p.fallbackDisplay = nil
		p.setBoxes(display.Boxes)
		p.generation++
		p.applyPending()

		return nil
	}

	return p.paintFallback(display, drawStart, track)
}

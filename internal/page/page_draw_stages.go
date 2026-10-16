package page

import (
	"context"
	"strings"
	"time"

	"github.com/chinmay-sawant/blinkless/css"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

// execTemplate runs the trusted template, timing it when tracked.
func (p *Page) execTemplate(track bool) (string, error) {
	var ts time.Time
	if track {
		ts = time.Now()
	}
	p.applyScrollWindow()
	var body strings.Builder
	if err := p.tpl.Execute(&body, p.data); err != nil {
		return "", err
	}
	executed := body.String()
	p.noteTime(track, &p.stats.lastTemplate, ts)
	return executed, nil
}

// styleStaged styles the source, timing it when tracked.
func (p *Page) styleStaged(ctx context.Context, track bool, source string, state render.State) (*css.Document, error) {
	var ls time.Time
	if track {
		ls = time.Now()
	}
	styled, err := p.styledDocument(ctx, source, state)
	p.noteTime(track, &p.stats.layoutTime, ls)
	return styled, err
}

// drawStaged builds the list and commits either path.
func (p *Page) drawStaged(ctx context.Context, track bool, styled *css.Document, state render.State, executed string) error {
	drawStart := time.Now()
	var ds time.Time
	if track {
		ds = time.Now()
	}
	display, err := render.DisplayListDocument(ctx, styled, state.Images)
	p.noteTime(track, &p.stats.displayListTime, ds)
	if err != nil {
		return err
	}
	p.paintCaret(display)
	if render.Replayable(display) {
		p.commitDisplay(display, executed, drawStart)
		return nil
	}
	if err := p.paintFallback(display, drawStart, track); err != nil {
		return err
	}
	p.commitAccessibility(executed)
	return nil
}

// commitDisplay installs a replayable list and its dirty region.
func (p *Page) commitDisplay(display *Display, executed string, drawStart time.Time) {
	p.dirtyFromDisplay(p.display, display)
	p.stats.layouts++
	p.stats.repaints++
	p.stats.lastDraw = time.Since(drawStart)
	p.img = nil
	p.display = display
	p.fallbackDisplay = nil
	p.setBoxes(display.Boxes)
	p.generation++
	p.commitAccessibility(executed)
	p.applyPending()
}

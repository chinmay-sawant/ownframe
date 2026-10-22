package page

import (
	"context"
	"github.com/chinmay-sawant/blinkless/css"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/page/draw"
	"github.com/chinmay-sawant/ownframe/internal/render"
	"runtime"
	"strings"
	"time"
)

// Redraw renders the current size: a display list for a replayable page,
// a bitmap otherwise. Parsed trees and sheets are reused. Stage timing and
// allocs track only with Perf on. One Page has one owner: never call Redraw
// concurrently on the same Page; keep Tick, MarkRect, and Redraw on the
// page loop the way Run and Serve each own their Page.
func (p *Page) Redraw(ctx context.Context) error {
	return p.redraw(ctx, false)
}

// RedrawIfChanged renders unless the executed source, viewport, focus,
// hover, theme, and images all match the last committed frame. It is for
// callers that redraw after every event and would otherwise lay out the
// same frame again. A skipped call counts no layout or repaint.
func (p *Page) RedrawIfChanged(ctx context.Context) error {
	return p.redraw(ctx, true)
}

func (p *Page) redraw(ctx context.Context, skipSame bool) error {
	if err := useContext(ctx); err != nil {
		return err
	}
	if p.afterRedraw.busy {
		// The after-redraw callback asked for a draw. The commit that runs
		// the callback must not start another one inside it, so the request
		// waits for the commit to finish.
		p.afterRedraw.ask = true

		return nil
	}
	if p.redrawBatch > 0 {
		p.redrawPending = true
		if !skipSame {
			p.redrawForce = true
		}
		return nil
	}
	track, m0, start := p.beginRedraw()
	defer p.finishRedraw(track, m0, start)
	executed, err := p.execTemplate(track)
	if err != nil {
		return err
	}
	p.source = p.syncForm(executed)
	state := p.renderState()
	if skipSame && p.unchanged(state) {
		return nil
	}
	p.draw.DropPNG()
	styled, err := p.styleStaged(ctx, track, p.source, state)
	if err != nil {
		return err
	}
	if err := p.drawStaged(ctx, track, styled, state, executed); err != nil {
		return err
	}

	return p.afterRedrawDraw(ctx)
}

// execTemplate runs the trusted template, timing it when tracked.
func (p *Page) execTemplate(track bool) (string, error) {
	ts := draw.Clock(track)
	p.applyScrollWindow()
	var body strings.Builder
	if len(p.source) > 0 {
		body.Grow(len(p.source))
	}
	if err := p.tpl.Execute(&body, p.data); err != nil {
		return "", err
	}
	executed := body.String()
	draw.Note(track, &p.stats.lastTemplate, ts)
	return executed, nil
}

// styleStaged styles the source, timing it when tracked.
func (p *Page) styleStaged(ctx context.Context, track bool, source string, state render.State) (*css.Document, error) {
	ls := draw.Clock(track)
	styled, err := p.styledDocument(ctx, source, state)
	draw.Note(track, &p.stats.layoutTime, ls)
	return styled, err
}

// drawStaged builds the list and commits either path.
func (p *Page) drawStaged(ctx context.Context, track bool, styled *css.Document, state render.State, executed string) error {
	drawStart := draw.Clock(true)
	ds := draw.Clock(track)
	display, err := render.DisplayListDocumentOptions(ctx, styled,
		layout.Options{Images: state.Images, RetainState: state.RetainState})
	draw.Note(track, &p.stats.displayListTime, ds)
	if err != nil {
		return err
	}
	p.paintCaret(display)
	if draw.Choose(display) == draw.PathList {
		p.commitDisplay(display, executed, drawStart, state)
		return nil
	}
	if err := p.paintFallback(display, drawStart, track); err != nil {
		return err
	}
	p.commitAccessibility(executed)
	return nil
}

// commitDisplay installs a replayable list and its dirty region.
func (p *Page) commitDisplay(display *Display, executed string, drawStart time.Time, state render.State) {
	p.dirtyFromDisplay(p.draw.Display, display)
	p.noteCommit(drawStart)
	p.draw.CommitList(display)
	p.setBoxes(display.Boxes)
	p.hits = hitList(display.Boxes, fixedBoxes(display))
	p.generation++
	p.remember(state)
	p.commitAccessibility(executed)
	p.applyPending()
	p.runAfterRedraw()
}

func (p *Page) noteCommit(drawStart time.Time) {
	p.stats.layouts++
	p.stats.repaints++
	p.stats.lastDraw = time.Since(drawStart)
}

// beginRedraw samples allocs when Perf is on and clears stale numbers
// when off, then counts the redraw. finishRedraw records the frame.
func (p *Page) beginRedraw() (bool, runtime.MemStats, time.Time) {
	track := p.perf
	var m0 runtime.MemStats
	if track {
		m0 = draw.AllocStart()
	} else {
		p.clearPerf()
	}
	start := time.Now()
	p.stats.redraws++
	return track, m0, start
}

func (p *Page) finishRedraw(track bool, m0 runtime.MemStats, start time.Time) {
	p.stats.lastRedraw = time.Since(start)
	if track {
		p.stats.allocFrame = draw.AllocUsed(m0)
	}
}

// paintFallback stores a bitmap for a page the vector path declined. The
// picture is a replay of that same list. blinkless does not paint a page
// bitmap of its own. The hit map is built here too, so a fixed box still
// wins over the content it covers on the bitmap path.
func (p *Page) paintFallback(display *layout.Display, drawStart time.Time, track bool) error {
	ps := draw.Clock(track)
	img, err := draw.Picture(display)
	if err != nil {
		return err
	}
	draw.Note(track, &p.stats.paintTime, ps)

	boxes := display.Boxes
	p.noteCommit(drawStart)
	p.draw.CommitBitmap(display, img)
	p.setBoxes(boxes)
	p.hits = hitList(boxes, fixedBoxes(display))
	p.generation++
	p.markFull()
	p.lastPatch = DisplayPatch{
		Changed: len(display.Ops),
		Added:   len(display.Ops),
		Regions: 1,
		Full:    true,
	}
	if track {
		p.recordFull(len(boxes))
	}
	p.applyPending()
	p.runAfterRedraw()

	return nil
}

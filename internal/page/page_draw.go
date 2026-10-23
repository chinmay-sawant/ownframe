package page

import (
	"context"
	"runtime"
	"strings"
	"time"

	"github.com/chinmay-sawant/blinkless/css"

	"github.com/chinmay-sawant/ownframe/internal/page/draw"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

func (p *Page) redraw(ctx context.Context, skipSame bool) error {
	if err := useContext(ctx); err != nil {
		return err
	}
	if p.afterRedraw.Busy() {
		// The after-redraw callback asked for a draw. The commit that runs
		// the callback must not start another one inside it, so the request
		// waits for the commit to finish.
		p.afterRedraw.Ask()

		return nil
	}
	if draw.QueueBatch(p.redrawBatch, &p.redrawPending, &p.redrawForce, skipSame) {
		return nil
	}

	sample, clear := draw.BeginSample(p.perf)
	if clear {
		p.clearPerf()
	}
	sample.StartClock()
	p.stats.redraws++

	defer p.finishRedraw(sample.Track, sample.Mem, sample.Start)

	executed, err := p.execTemplate(sample.Track)
	if err != nil {
		return err
	}

	p.source = p.syncForm(executed)
	state := p.renderState()
	if skipSame && p.unchanged(state) {
		return nil
	}

	p.draw.DropPNG()
	styled, err := p.styleStaged(ctx, sample.Track, p.source, state)
	if err != nil {
		return err
	}
	if err := p.drawStaged(ctx, sample.Track, styled, state, executed); err != nil {
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
func (p *Page) styleStaged(
	ctx context.Context,
	track bool,
	source string,
	state render.State,
) (*css.Document, error) {
	ls := draw.Clock(track)
	styled, err := p.styledDocument(ctx, source, state)
	draw.Note(track, &p.stats.layoutTime, ls)

	return styled, err
}

// drawStaged builds the list and commits either path.
func (p *Page) drawStaged(
	ctx context.Context,
	track bool,
	styled *css.Document,
	state render.State,
	executed string,
) error {
	drawStart := draw.Clock(true)
	ds := draw.Clock(track)
	display, err := draw.List(ctx, styled, state)
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
	draw.NoteCommit(&p.stats.layouts, &p.stats.repaints, &p.stats.lastDraw, drawStart)
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
func (p *Page) paintFallback(display *Display, drawStart time.Time, track bool) error {
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

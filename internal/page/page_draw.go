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
	return draw.Run(ctx, redrawSurface{p}, skipSame)
}

// redrawSurface adapts Page to draw.RedrawSurface.
// Each method delegates to the page step the sequence already used.
type redrawSurface struct{ p *Page }

func (s redrawSurface) AfterBusy() bool { return s.p.afterRedraw.Busy() }

func (s redrawSurface) AfterAsk() { s.p.afterRedraw.Ask() }

func (s redrawSurface) QueueBatch(skipSame bool) bool {
	return draw.QueueBatch(
		s.p.redrawBatch,
		&s.p.redrawPending,
		&s.p.redrawForce,
		skipSame,
	)
}

func (s redrawSurface) BeginSample() (draw.Sample, bool) {
	return draw.BeginSample(s.p.perf)
}

func (s redrawSurface) ClearPerf() { s.p.clearPerf() }

func (s redrawSurface) CountRedraw() { s.p.stats.redraws++ }

func (s redrawSurface) FinishRedraw(track bool, mem runtime.MemStats, start time.Time) {
	s.p.finishRedraw(track, mem, start)
}

func (s redrawSurface) ExecTemplate(track bool) (string, error) {
	return s.p.execTemplate(track)
}

func (s redrawSurface) SyncForm(executed string) string {
	s.p.source = s.p.syncForm(executed)

	return s.p.source
}

func (s redrawSurface) RenderState() render.State { return s.p.renderState() }

func (s redrawSurface) Unchanged(state render.State) bool { return s.p.unchanged(state) }

func (s redrawSurface) DropPNG() { s.p.draw.DropPNG() }

func (s redrawSurface) StyleStaged(
	ctx context.Context,
	track bool,
	source string,
	state render.State,
) (*css.Document, error) {
	return s.p.styleStaged(ctx, track, source, state)
}

func (s redrawSurface) DrawStaged(
	ctx context.Context,
	track bool,
	styled *css.Document,
	state render.State,
	executed string,
) error {
	return s.p.drawStaged(ctx, track, styled, state, executed)
}

func (s redrawSurface) AfterRedrawDraw(ctx context.Context) error {
	return s.p.afterRedrawDraw(ctx)
}

var _ draw.RedrawSurface = redrawSurface{}

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

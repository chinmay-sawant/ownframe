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
	s.p.stats.lastTemplate = 0
	s.p.stats.styleTime = 0
	s.p.stats.layoutTime = 0
	s.p.stats.paintTime = 0
	s.p.stats.otherTime = 0
	if !s.p.perf {
		return draw.BeginSample(false, false)
	}

	return draw.BeginSample(true, s.p.allocDue(time.Now()))
}

func (s redrawSurface) ClearPerf() { s.p.clearPerf() }

func (s redrawSurface) CountRedraw() { s.p.stats.redraws++ }

func (s redrawSurface) FinishRedraw(track, memOK bool, mem runtime.MemStats, start time.Time) {
	s.p.finishRedraw(track, memOK, mem, start)
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
	p.applyScrollWindow()

	p.dataMu.RLock()
	defer p.dataMu.RUnlock()

	return draw.Execute(
		track,
		len(p.source),
		&p.stats.lastTemplate,
		func(b *strings.Builder) error {
			return p.tpl.Execute(b, p.data)
		},
	)
}

// styleStaged styles the source, timing it when tracked.
func (p *Page) styleStaged(
	ctx context.Context,
	track bool,
	source string,
	state render.State,
) (*css.Document, error) {
	styled, err := draw.StyleDocument(
		ctx,
		track,
		&p.stats.styleTime,
		p.cache,
		source,
		p.viewport(),
		state,
	)
	if err != nil {
		return nil, err
	}

	p.cache = styled.Cache
	if styled.Fresh {
		p.statePseudos = styled.Pseudos
		p.stats.parses++
		p.stats.cascades++
	}

	return styled.Doc, nil
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
	display, err := p.layoutStaged(ctx, styled, state)
	draw.Note(track, &p.stats.layoutTime, ds)
	if err != nil {
		return err
	}

	p.paintCaret(display)
	if draw.Choose(display) == draw.PathList {
		p.commitDisplay(display, executed, drawStart, state)

		return nil
	}
	if err := p.paintFallback(display, drawStart, track, state); err != nil {
		return err
	}

	p.commitAccessibility(executed)

	return nil
}

// commitDisplay installs a replayable list and its dirty region.
func (p *Page) commitDisplay(display *Display, executed string, drawStart time.Time, state render.State) {
	previous := p.draw.Display
	captures := p.captureTransitions(previous, display)
	p.dirtyFromDisplay(previous, display)
	p.noteCommit(drawStart)
	p.draw.CommitList(display)
	p.setDisplayBoxes(display)
	p.generation++
	p.layerGen++
	p.bindMotion(previous, display)
	p.bindTransitions(captures)
	p.remember(state)
	p.commitAccessibility(executed)
	p.applyPending()
	p.runAfterRedraw()
}

func (p *Page) noteCommit(drawStart time.Time) {
	draw.NoteCommit(&p.stats.layouts, &p.stats.repaints, &p.stats.lastDraw, drawStart)
}

func (p *Page) finishRedraw(track, memOK bool, m0 runtime.MemStats, start time.Time) {
	p.stats.lastRedraw = time.Since(start)
	if track {
		p.stats.otherTime = max(0, p.stats.lastRedraw-p.stats.lastTemplate-
			p.stats.styleTime-p.stats.layoutTime-p.stats.paintTime)
	}
	if track && memOK {
		p.stats.allocFrame = draw.AllocUsed(m0)
	}
}

// paintFallback stores a bitmap for a page the vector path declined. The
// picture is a replay of that same list. blinkless does not paint a page
// bitmap of its own. The hit map is built here too, so a fixed box still
// wins over the content it covers on the bitmap path.
func (p *Page) paintFallback(display *Display, drawStart time.Time, track bool, state render.State) error {
	bitmap, err := draw.PaintBitmap(track, &p.stats.paintTime, display)
	if err != nil {
		return err
	}

	previous := p.draw.Display
	p.captureTransitions(previous, display)
	p.noteCommit(drawStart)
	p.draw.CommitBitmap(display, bitmap.Image)
	p.setDisplayBoxes(display)
	p.generation++
	p.layerGen++
	p.bindMotion(previous, nil)
	p.bindTransitions(nil)
	p.remember(state)
	p.markFull()
	p.lastPatch = DisplayPatch{
		Changed: bitmap.Changed,
		Added:   bitmap.Added,
		Regions: bitmap.Regions,
		Full:    bitmap.Full,
	}
	if track {
		p.recordFull(len(display.Boxes))
	}
	p.applyPending()
	p.runAfterRedraw()

	return nil
}

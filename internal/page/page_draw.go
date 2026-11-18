package page

import (
	"context"
	"io"
	"runtime"
	"strings"
	"time"

	"github.com/chinmay-sawant/blinkless/css"
	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
	"github.com/chinmay-sawant/ownframe/internal/page/draw"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

func (p *Page) redraw(ctx context.Context, skipSame bool) error {
	return draw.Run(ctx, p.surf, skipSame)
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
	out := s.p.syncForm(executed)
	s.p.source = out
	// A clean fast-path sync leaves the source byte-identical, so the
	// memoized hash, length, and revision from the last real sync are
	// still valid and the rehash is pure overhead on the skip path.
	if !s.p.syncClean {
		s.p.noteSource(out)
	}

	return out
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
//
// The "{{" probe is memoized to template-write time: the scan runs once
// when p.tpl changes and the verdict is reused, so repeat frames pay no
// scan. A clean form reuses the last merged body with no Execute. A dirty
// frame executes into a pooled buffer and keeps one copy; its hash is taken
// once in the sync path, which pairs it with the body for the next reuse.
func (p *Page) execTemplate(track bool) (string, error) {
	p.applyScrollWindow()
	if p.tplProbeTpl != p.tpl {
		p.tplProbeHas = strings.Contains(p.nav.Current(), "{{")
		p.tplProbeTpl = p.tpl
	}

	if !p.tplProbeHas {
		p.stats.lastTemplate = 0
		p.execSkipped = true

		return p.nav.Current(), nil
	}

	p.dataMu.RLock()
	defer p.dataMu.RUnlock()

	// A Redraw with no new data, form input, or changed row window reuses
	// the last merged template body. Every path that runs application code
	// marks the form dirty first (handlers can mutate the data in place
	// through the SetData pointer), so a clean flag means the body is
	// still current.
	if !p.formDirty && p.lastSyncBody != "" {
		p.stats.lastTemplate = 0
		p.execSkipped = true

		return p.lastSyncBody, nil
	}

	p.execSkipped = false

	out, err := draw.ExecuteWriter(
		track,
		len(p.source),
		&p.stats.lastTemplate,
		func(w io.Writer) error {
			return p.tpl.Execute(w, p.data)
		},
	)
	if err != nil {
		return "", err
	}

	body := string(out)

	return body, nil
}

// styleStaged styles the source, timing it when tracked.
//
// A theme swap forces a fresh Apply: css.Relayout keeps the cached
// document's extra sheets, so the cacheFresh relayout would keep painting
// the old theme. A state-only move the sheets cannot match reuses the cached
// document with no relayout at all.
//
// Otherwise, when the style cache already reflects the current source memo,
// the page relayouts directly: the revision, hash, and length compare with no
// cache lock, and the relayout takes one. Only a new source (or a dropped
// cache) pays the Source memcmp and the fresh parse inside StyleDocument.
func (p *Page) styleStaged(
	ctx context.Context,
	track bool,
	source string,
	state render.State,
) (*css.Document, error) {
	if doc, ok := p.styleNoop(state); ok {
		return doc, nil
	}

	changed := p.themeChanged(state)
	if !changed && p.cacheFresh() {
		// An element patch the hover handler just committed already
		// relayed the cache for this state, so reuse its document.
		if doc, ok := p.relayedDoc(state); ok {
			return doc, nil
		}
		ds := draw.Clock(track)

		doc, err := p.cache.Relayout(ctx, p.width, p.height, state)
		draw.Note(track, &p.stats.styleTime, ds)
		if err != nil {
			return nil, err
		}

		p.noteRelayed(state)
		return doc, nil
	}

	styled, err := draw.StyleDocument(
		ctx,
		track,
		&p.stats.styleTime,
		styleCache(p.cache, changed),
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
	} else if changed {
		// A relayout keeps the old extra sheet, so refresh what the new
		// document can match even without a fresh cascade.
		p.statePseudos = p.cache.StatePseudos()
	}

	p.noteCacheSource()
	p.noteRelayed(state)

	return styled.Doc, nil
}

// styleCache selects the cache a style call starts from. A theme change
// passes nil so the call takes the fresh Apply: css.Relayout keeps the
// cached document's extra sheets and would keep the old theme.
func styleCache(cache *render.Cache, themeChanged bool) *render.Cache {
	if themeChanged {
		return nil
	}

	return cache
}

// themeChanged reports that the theme sheet differs from the one the last
// committed frame was styled with. The cache still reflects the old sheet,
// so the caller must not take the cacheFresh relayout for it.
func (p *Page) themeChanged(state render.State) bool {
	return p.cache != nil && p.last.Theme != state.Theme
}

// styleNoop returns the cached styled document when a redraw changed only
// pointer or focus state the sheets cannot match, so a relayout would
// reproduce the same styles. ok is false when the frame changed anything
// else, or when a changed state can match, and the caller restyles.
func (p *Page) styleNoop(state render.State) (doc *css.Document, ok bool) {
	if p.cache == nil || !p.hasLast {
		return nil, false
	}

	prev := render.State{Hover: p.last.Hover, Active: p.last.Active, Focus: p.last.Focus}
	if !state.SameStyleState(prev, p.statePseudos) {
		return nil, false
	}

	next, last := p.keyFor(state), p.last
	next.Focus, next.Hover, next.Active = "", "", ""
	last.Focus, last.Hover, last.Active = "", "", ""

	if next != last {
		return nil, false
	}

	doc = p.cache.Styled()
	if doc == nil {
		return nil, false
	}

	return doc, true
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
	previousFallback := p.draw.FallbackDisplay
	p.updateMotionIndex(previous, display)
	captures := p.captureTransitions(previous, display)
	p.dirtyFromDisplay(previous, display)
	p.noteCommit(drawStart)
	p.draw.CommitList(display)
	p.setDisplayBoxes(display)
	p.generation++
	p.layerGen++
	p.bindMotion(previous, display)
	p.bindTransitions(captures)
	if previous != nil && previous != display {
		previous.Release()
	}
	if previousFallback != nil && previousFallback != display && previousFallback != previous {
		previousFallback.Release()
	}
	p.remember(state)
	p.commitAccessibility(executed)
	p.applyPending()
	p.runAfterRedraw()
}

func (p *Page) noteCommit(drawStart time.Time) {
	draw.NoteCommit(&p.stats.layouts, &p.stats.repaints, &p.stats.lastDraw, drawStart)
}

func (p *Page) finishRedraw(track, memOK bool, m0 runtime.MemStats, start time.Time) {
	defer p.notifyFrame()
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
	previousFallback := p.draw.FallbackDisplay
	p.updateMotionIndex(previous, nil)
	p.captureTransitions(previous, display)
	p.noteCommit(drawStart)
	p.draw.CommitBitmap(display, bitmap.Image)
	p.setDisplayBoxes(display)
	p.generation++
	p.layerGen++
	p.bindMotion(previous, nil)
	p.bindTransitions(nil)
	if previous != nil && previous != display {
		previous.Release()
	}
	if previousFallback != nil && previousFallback != display && previousFallback != previous {
		previousFallback.Release()
	}
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

// updateMotionIndex updates the cached *cssmotion.Index for display.
// If display.Keyframes matches the previous display's keyframes (pointer equality
// or length == 0), it reuses the previous *cssmotion.Index instead of calling
// NewIndex on every commit.
func (p *Page) updateMotionIndex(previous, display *Display) *cssmotion.Index {
	if display == nil || len(display.Keyframes) == 0 {
		p.motionIndex = cssmotion.EmptyIndex()
		return p.motionIndex
	}
	if previous != nil && sameKeyframes(display.Keyframes, previous.Keyframes) && p.motionIndex != nil && p.motionIndex != cssmotion.EmptyIndex() {
		if p.motionIndex.Display() == display || len(display.Ops) == p.motionIndex.NumOps() {
			p.motionIndex.Repoint(display)
			return p.motionIndex
		}
	}
	p.motionIndex = cssmotion.NewIndex(display)
	return p.motionIndex
}

// sameKeyframes reports whether two keyframe rule slices share the same rules
// by length or pointer equality.
func sameKeyframes(a, b []layout.KeyframesRule) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	return &a[0] == &b[0]
}

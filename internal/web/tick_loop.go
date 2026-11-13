package web

import (
	"context"
	"image"
	"log"
	"sync"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// tickEvery is one frame at about 60 Hz.
const tickEvery = time.Second / 60

// frameDirty is a screen that can report a changed picture after Tick.
type frameDirty interface {
	FrameDirty() bool
}

// dirtyTaker is a screen that can drain its damage. Page implements it.
// The web loop drains it to tell fresh damage from a stale mark the host
// never consumed, so a bare UseFrameDirty opt-in without rects does not
// advance the frame generation.
type dirtyTaker interface {
	TakeDirtyRects() ([]image.Rectangle, bool)
}

// startTick runs app's Tick until ctx ends. A screen without Tick gets a
// no-op stop and no goroutine. The stop function cancels the loop and waits
// for it, so a caller can release the page without a late tick.
func (s *server) startTick(ctx context.Context) func() {
	ticker, ok := s.app.(host.Ticker)
	if !ok {
		return func() {}
	}

	s.mu.Lock()
	s.ticking = true
	s.mu.Unlock()

	tickCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	var once sync.Once
	stop := func() {
		once.Do(func() {
			cancel()
			<-done
		})
	}

	go func() {
		defer close(done)
		defer s.setTicking(false)
		defer func() {
			if r := recover(); r != nil {
				log.Printf("web tick panic: %v", r)
			}
		}()
		s.runTick(tickCtx, ticker)
	}()

	return stop
}

// setTicking updates the value /state reports.
func (s *server) setTicking(on bool) {
	s.mu.Lock()
	s.ticking = on
	s.mu.Unlock()
}

// runTick is the 60 Hz loop. It stops on ctx end or the first tick error.
func (s *server) runTick(ctx context.Context, ticker host.Ticker) {
	t := time.NewTicker(tickEvery)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}

		if ctx.Err() != nil {
			return
		}
		if err := s.tickOnce(ctx, ticker); err != nil {
			if ctx.Err() == nil {
				log.Printf("tick: %s", sanitizeLine(err.Error()))
			}

			return
		}
	}
}

// tickOnce runs one Tick under the server lock and redraws only when the
// screen reports damage or changed its generation. A dirty screen that
// patches retained operations in place keeps its generation, so frameGen
// advances there and /state still sees a new frame. The Take drains the
// damage the Tick just produced: a bare FrameDirty opt-in with no rects
// (an idle UseFrameDirty, or a stale mark the web loop never consumed)
// does not advance frameGen, and a real commit drains its own damage so
// the next idle tick does not bump twice for the same frame.
func (s *server) tickOnce(ctx context.Context, ticker host.Ticker) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	before := s.app.Generation()
	if err := ticker.Tick(ctx); err != nil {
		return err
	}

	dirty := false
	if marker, ok := s.app.(frameDirty); ok {
		dirty = marker.FrameDirty()
	}
	if !dirty && s.app.Generation() == before {
		s.publishIfChangedLocked()
		return nil
	}

	var rects []image.Rectangle
	preDamage := dirty
	tookRects := false
	if taker, ok := s.app.(dirtyTaker); ok {
		drained, ok := taker.TakeDirtyRects()
		preDamage = ok && len(drained) > 0
		tookRects = preDamage
		if tookRects {
			rects = drained
		}
	}

	// In-place tick: the generation is frozen and the drained damage names
	// the changed regions, so advance the frame without a Redraw. The
	// retained operations already carry the new paint; a commit here would
	// only re-execute an unchanged template (or, with an active CSS
	// animation, force a full commit every frame and spam /dom). Ticks that
	// report damage without drainable rects still fall through to Redraw
	// below, preserving today's behavior for those screens.
	if s.app.Generation() == before && tookRects {
		if s.patchCacheGen != before {
			s.patchCache = map[string][2]string{}
			s.patchStyleIDs = nil
			s.patchContainers = nil
			s.patchCacheGen = before
		}
		if s.patchCache == nil {
			s.patchCache = map[string][2]string{}
		}
		if s.patchStyleIDs == nil || s.patchContainers == nil {
			source := executedSourceOf(s.app)
			s.patchStyleIDs = inlineWidthIDs(source)
			s.patchContainers = containerIDs(source)
		}
		patches := collectTickPatches(s.app.Display(), s.app.Boxes(), rects, s.patchCache, s.patchStyleIDs, controlIDs(s.liveControls()), s.patchContainers)
		s.frameGen++
		if s.frameGen <= before {
			s.frameGen = before + 1
		}
		s.publishTickLocked(patches)
		return nil
	}

	if err := s.app.Redraw(ctx); err != nil {
		return err
	}

	if s.app.Generation() == before {
		if preDamage {
			s.frameGen++
			if s.frameGen <= before {
				s.frameGen = before + 1
			}
		}

		s.publishIfChangedLocked()
		return nil
	}

	if taker, ok := s.app.(dirtyTaker); ok {
		_, _ = taker.TakeDirtyRects()
	}

	s.publishIfChangedLocked()
	return nil
}

package web

import (
	"context"
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
// advances there and /state still sees a new frame.
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
		return nil
	}

	if err := s.app.Redraw(ctx); err != nil {
		return err
	}

	if dirty && s.app.Generation() == before {
		s.frameGen++
		if s.frameGen <= before {
			s.frameGen = before + 1
		}
	}

	return nil
}

package window

import (
	"context"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

type updateStage struct {
	name string
	run  func(*shell) error
	// batch shares one redraw batch across the contiguous batched span,
	// so a hover change that coincides with an app refresh lays the page
	// out once instead of twice. The batch opens before the first batched
	// stage runs and flushes after the last one, before syncImage reads
	// the committed generation.
	batch bool
}

func void(f func(*shell)) func(*shell) error {
	return func(s *shell) error { f(s); return nil }
}

var stages = []updateStage{
	{"ctx", func(s *shell) error { return s.ctx.Err() }, false},
	{"close", func(s *shell) error {
		if takeCloseRequest(isWSLg() && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)) {
			return ebiten.Termination
		}
		return nil
	}, false},
	{"devSync", func(s *shell) error { return s.devSync() }, false},
	{"ime", func(s *shell) error { return s.imeUpdate() }, false},
	{"keys", func(s *shell) error { return s.keys() }, false},
	{"resize", func(s *shell) error { return s.resize() }, false},
	{"pass", void((*shell).updatePassthrough), false},
	{"pointer", func(s *shell) error { return s.pointer() }, true},
	{"touch", func(s *shell) error { s.stepTouchScroll(time.Now()); return nil }, true},
	{"hold", func(s *shell) error { return s.holdFrame() }, true},
	{"drop", func(s *shell) error { return s.dropPass(ebiten.DroppedFiles()) }, true},
	{"tick", func(s *shell) error { return s.timedErr(&s.dev.stages.tick, s.tickFrame) }, true},
	{"optin", void((*shell).captureFrameOptIn), false},
	{"reload", func(s *shell) error { return s.pollReload() }, false},
	{"sync", func(s *shell) error {
		if err := s.timedErr(&s.dev.stages.sync, s.syncImage); err != nil {
			return err
		}
		return nil
	}, false},
	{"scrollReq", void((*shell).applyScrollRequest), false},
	{"devRefresh", void((*shell).devRefresh), false},
	{"wheel", (*shell).wheel, false},
	{"scrollWin", func(s *shell) error { return s.syncScrollWindow() }, false},
	{"bitmap", func(s *shell) error { return s.prepareBitmapViewport() }, false},
}

func (s *shell) Update() error {
	if s.drawSyncErr != nil {
		err := s.drawSyncErr
		s.drawSyncErr = nil
		return err
	}

	if s.perf {
		start := time.Now()
		defer func() {
			elapsed := time.Since(start)
			s.dev.stages.update = elapsed
			s.dev.stages.pendingUpdate += elapsed
		}()
	}

	if s.app == nil {
		return errNilApp
	}

	// The pointer span resolves hover and gestures and the tick span runs
	// the frame callback; each redraws, so a hover change that coincides
	// with an app refresh lays the page out twice back to back. The
	// batched stages share one redraw batch, and the single flush before
	// syncImage covers both. Batching defers commits, never input: hover,
	// clicks, picks, drops, and the tick callback itself run in order, and
	// the flush lands in the same frame, so hover still paints within one
	// frame. A screen without the batcher role runs every stage directly,
	// exactly as before.
	batcher, ok := s.app.(host.RedrawBatcher)
	batched := false
	// A stage runs app code and can panic. This does not recover: the panic
	// keeps travelling to the crash report. The open batch must not leak
	// into the next frame, so end it without drawing.
	defer func() {
		if batched {
			batched = false
			s.endBatchQuiet(batcher)
		}
	}()
	flush := func() error {
		batched = false
		return batcher.EndRedrawBatch(s.ctx)
	}
	for _, st := range stages {
		if ok && st.batch && !batched {
			batcher.BeginRedrawBatch()
			batched = true
		}
		if batched && !st.batch {
			if err := flush(); err != nil {
				return err
			}
		}
		if err := st.run(s); err != nil {
			// The frame is doomed, so end the batch without drawing:
			// the queued draw stays queued and the next batch commits
			// it against the state that batch changed.
			if batched {
				batched = false
				s.endBatchQuiet(batcher)
			}
			return err
		}
	}
	if batched {
		if err := flush(); err != nil {
			return err
		}
	}
	s.maybePublishDebug()
	return nil
}

// tickFrame runs the screen's per-frame callback when it has one.
func (s *shell) tickFrame() error {
	ticker, ok := s.app.(host.Ticker)
	if !ok {
		return nil
	}
	return ticker.Tick(s.ctx)
}

// endBatchQuiet ends a batch without drawing, so a stage error or panic
// never leaves the deferral open for the next frame. The draw runs with a
// dead context, so the page records nothing, and a queued draw stays queued
// for the next batch. A panic inside the end must not replace the panic
// that is already travelling.
func (s *shell) endBatchQuiet(batcher host.RedrawBatcher) {
	if batcher == nil || s.ctx == nil {
		return
	}
	dead, cancel := context.WithCancel(s.ctx)
	cancel()
	defer func() { _ = recover() }()
	_ = batcher.EndRedrawBatch(dead)
}

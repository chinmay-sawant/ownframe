package window

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

type updateStage struct {
	name string
	run  func(*shell) error
}

func void(f func(*shell)) func(*shell) error {
	return func(s *shell) error { f(s); return nil }
}

var stages = []updateStage{
	{"ctx", func(s *shell) error { return s.ctx.Err() }},
	{"devSync", func(s *shell) error { return s.devSync() }},
	{"ime", func(s *shell) error { return s.imeUpdate() }},
	{"keys", func(s *shell) error { return s.keys() }},
	{"resize", func(s *shell) error { return s.resize() }},
	{"pass", void((*shell).updatePassthrough)},
	{"pointer", func(s *shell) error { return s.pointer() }},
	{"touch", func(s *shell) error { s.stepTouchScroll(time.Now()); return nil }},
	{"hold", func(s *shell) error { return s.holdFrame() }},
	{"drop", func(s *shell) error { return s.dropPass(ebiten.DroppedFiles()) }},
	{"tick", func(s *shell) error { return s.timedErr(&s.dev.stages.tick, s.tickFrame) }},
	{"optin", void((*shell).captureFrameOptIn)},
	{"reload", func(s *shell) error { return s.pollReload() }},
	{"sync", func(s *shell) error {
		if err := s.timedErr(&s.dev.stages.sync, s.syncImage); err != nil {
			return err
		}
		return nil
	}},
	{"scrollReq", void((*shell).applyScrollRequest)},
	{"devRefresh", void((*shell).devRefresh)},
	{"wheel", (*shell).wheel},
	{"scrollWin", func(s *shell) error { return s.syncScrollWindow() }},
	{"bitmap", func(s *shell) error { return s.prepareBitmapViewport() }},
}

func (s *shell) Update() error {
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
	for _, st := range stages {
		if err := st.run(s); err != nil {
			return err
		}
	}
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

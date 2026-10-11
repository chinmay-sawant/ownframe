package window

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

func (s *shell) Update() error {
	if s.app == nil {
		return errNilApp
	}

	if err := s.ctx.Err(); err != nil {
		return err
	}

	if err := s.devSync(); err != nil {
		return err
	}

	if err := s.imeUpdate(); err != nil {
		return err
	}

	if err := s.keys(); err != nil {
		return err
	}

	if err := s.resize(); err != nil {
		return err
	}

	s.updatePassthrough()
	if err := s.pointer(); err != nil {
		return err
	}
	s.stepTouchScroll(time.Now())

	if err := s.holdFrame(); err != nil {
		return err
	}

	if err := s.dropPass(ebiten.DroppedFiles()); err != nil {
		return err
	}

	if err := s.timedErr(&s.dev.stages.tick, s.tickFrame); err != nil {
		return err
	}

	if err := s.pollReload(); err != nil {
		return err
	}

	if err := s.timedErr(&s.dev.stages.sync, s.syncImage); err != nil {
		return err
	}
	if s.perf {
		s.dev.stages.update = s.dev.stages.tick + s.dev.stages.sync
	}

	s.applyScrollRequest()
	s.devRefresh()
	s.wheel()
	if err := s.syncScrollWindow(); err != nil {
		return err
	}

	return s.prepareBitmapViewport()
}

// tickFrame runs the screen's per-frame callback when it has one.
func (s *shell) tickFrame() error {
	ticker, ok := s.app.(host.Ticker)
	if !ok {
		return nil
	}

	return ticker.Tick(s.ctx)
}

//go:build linux && !android && !ios && !js

package window

import (
	"sync"

	"github.com/chinmay-sawant/ownframe/internal/atspi"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/hajimehoshi/ebiten/v2"
)

// a11yHost holds the Linux AT-SPI publisher. The render loop snapshots a
// changed tree and hands it to one worker, so a slow D-Bus call never
// stalls Update.
type a11yHost struct {
	actions   *atspi.Actions
	pub       *atspi.Publisher
	gen       uint64
	published bool
	retry     bool

	mu        sync.Mutex
	queued    int
	running   bool
	closed    bool
	shutdown  bool
	jobs      chan a11yJob
	done      chan struct{}
	closeOnce sync.Once
	wg        sync.WaitGroup
}

type a11yJob struct {
	tree host.AccessibilityTree
	gen  uint64
}

func (s *shell) initA11y() {
	s.a11y.actions = atspi.NewActions(s.ctx, func() {
		//nolint:staticcheck // reason: concurrent-safe wake for an AT request while the desktop is parked in WaitEvents.
		ebiten.ScheduleFrame()
	})
	s.a11y.pub = atspi.NewPublisher(atspi.NewLiveBus(s.a11y.actions.Request))
	s.a11y.jobs = make(chan a11yJob, 1)
	s.a11y.done = make(chan struct{})
	if s.ctxDone != nil {
		go s.watchA11y(s.ctxDone)
	}
}

// watchA11y closes the publisher when the shell context ends, which
// covers NewGame hosts that never call runWindow. It also exits when
// closeA11y runs on its own, so window tests do not leak it.
func (s *shell) watchA11y(ctxDone <-chan struct{}) {
	select {
	case <-ctxDone:
		s.closeA11y()
	case <-s.a11y.done:
	}
}

func (s *shell) closeA11y() {
	a := &s.a11y
	a.closeOnce.Do(func() {
		if a.actions != nil {
			a.actions.Close()
		}
		a.mu.Lock()
		a.closed = true
		a.mu.Unlock()
		if a.done != nil {
			close(a.done)
		}
		a.wg.Wait()
		if a.pub != nil {
			_ = a.pub.Close()
		}
		a.mu.Lock()
		a.shutdown = true
		a.mu.Unlock()
	})
}

func (s *shell) runA11yActions() {
	if s.a11y.actions == nil {
		return
	}
	if acc, ok := s.app.(host.Accessible); ok {
		s.a11y.actions.Drain(acc.AccessibilityPerform)
	}
}

func (s *shell) syncA11y() {
	if s.a11y.pub == nil {
		return
	}
	// Coordinates do not touch D-Bus. Keep them current on the render
	// loop so moving the window updates the root extents even when the
	// page tree itself did not change.
	x, y := 0, 0
	if s.a11yOrigin != nil {
		x, y = s.a11yOrigin()
	}
	s.a11y.pub.SetOrigin(x, y)
	s.a11y.pub.SetSize(s.screenW, s.screenH)
	if !s.a11y.pub.Ready() {
		return
	}
	s.publishA11y()
}

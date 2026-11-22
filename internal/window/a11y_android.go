//go:build android

package window

import (
	"sync"

	"github.com/chinmay-sawant/ownframe/internal/a11ybridge"
	"github.com/chinmay-sawant/ownframe/internal/atspi"
	"github.com/chinmay-sawant/ownframe/internal/host"
)

// a11yHost serves the page to Android's TalkBack. The host view reads the
// snapshot this loop publishes through a11ybridge and asks for actions from
// its own threads; the actions come back through the queue and run here, on
// the page loop, as they do for AT-SPI and UI Automation.
type a11yHost struct {
	actions   *atspi.Actions
	gen       uint64
	published bool
	closeOnce sync.Once
}

func (s *shell) initA11y() {
	// A phone loop runs every frame while the activity is visible, so a queued
	// action needs no wake-up call.
	s.a11y.actions = atspi.NewActions(s.ctx, func() {})
	a11ybridge.SetPerformer(s.a11y.actions.Request)
}

func (s *shell) runA11yActions() {
	if s.a11y.actions == nil {
		return
	}

	if acc, ok := s.app.(host.Accessible); ok {
		s.a11y.actions.Drain(acc.AccessibilityPerform)
	}
}

// syncA11y publishes a new snapshot when the page redrew.
func (s *shell) syncA11y() {
	a := &s.a11y
	if a.actions == nil {
		return
	}

	acc, ok := s.app.(host.Accessible)
	if !ok {
		return
	}

	gen := s.app.Generation()
	if a.published && a.gen == gen {
		return
	}

	tree, err := acc.AccessibilityTree()
	if err != nil {
		return
	}

	a11ybridge.SetTree(a11ybridge.Encode(tree))

	a.gen, a.published = gen, true
}

func (s *shell) closeA11y() {
	a := &s.a11y
	a.closeOnce.Do(func() {
		if a.actions != nil {
			a.actions.Close()
		}

		a11ybridge.Reset()
	})
}

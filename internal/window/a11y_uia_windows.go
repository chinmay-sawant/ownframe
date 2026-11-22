//go:build windows && (amd64 || arm64)

package window

import (
	"sync"

	"github.com/chinmay-sawant/ownframe/internal/atspi"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/uia"
	"github.com/hajimehoshi/ebiten/v2"
)

// uiaRetryFrames spaces attach attempts while the native window or the first
// tree is not ready, so a failing attach never costs a frame.
const uiaRetryFrames = 30

// a11yHost serves the page to Windows UI Automation. Every method except
// runA11yActions runs on the render loop; UIA clients call into
// uia.Provider from their own threads, and actions come back through
// actions.Drain on the loop.
type a11yHost struct {
	actions *atspi.Actions
	prov    *uia.Provider
	host    *uia.Host

	gen       uint64
	published bool
	skip      int
	closeOnce sync.Once
}

func (s *shell) initA11y() {
	s.a11y.actions = atspi.NewActions(s.ctx, func() {
		//nolint:staticcheck // reason: concurrent-safe wake for a UIA request while the desktop is parked in WaitEvents.
		ebiten.ScheduleFrame()
	})
	s.a11y.prov = uia.NewProvider(s.a11y.actions.Request)
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
	a := &s.a11y
	if a.prov == nil {
		return
	}
	acc, ok := s.app.(host.Accessible)
	if !ok {
		return
	}
	gen := s.app.Generation()
	if a.host == nil {
		if a.skip > 0 {
			a.skip--
			return
		}
		if !s.attachUIA(acc, gen) {
			a.skip = uiaRetryFrames
			return
		}
	} else if !a.published || a.gen != gen {
		tree, err := acc.AccessibilityTree()
		if err != nil {
			return
		}
		a.host.Publish(tree)
		a.gen, a.published = gen, true
	}
	s.placeUIA()
}

// attachUIA publishes the first tree and subclasses the window.
func (s *shell) attachUIA(acc host.Accessible, gen uint64) bool {
	a := &s.a11y
	tree, err := acc.AccessibilityTree()
	if err != nil {
		return false
	}
	a.prov.Update(tree)
	h, err := uia.Attach(a.prov)
	if err != nil {
		return false
	}
	a.host, a.gen, a.published = h, gen, true
	return true
}

// placeUIA tells the provider where page content sits on screen. The client
// area comes from the native window in physical pixels; the built-in title
// bar, drawn inside it, shifts the page down.
func (s *shell) placeUIA() {
	g, ok := s.a11y.host.ClientGeometry()
	if !ok {
		return
	}
	inset := float64(s.titlebarInset()) * g.Scale
	g.OriginY += inset
	g.Height = max(0, g.Height-inset)
	s.a11y.prov.SetGeometry(g)
}

func (s *shell) closeA11y() {
	a := &s.a11y
	a.closeOnce.Do(func() {
		if a.actions != nil {
			a.actions.Close()
		}
		if a.prov != nil {
			a.prov.Stop()
		}
		if a.host != nil {
			a.host.Close()
		}
	})
}

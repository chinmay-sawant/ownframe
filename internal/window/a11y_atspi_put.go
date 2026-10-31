//go:build linux && !android && !ios && !js

package window

import "github.com/chinmay-sawant/ownframe/internal/host"

func (s *shell) publishA11y() {
	if s.a11y.pub == nil {
		return
	}
	acc, ok := s.app.(host.Accessible)
	if !ok {
		return
	}
	gen := s.app.Generation()
	if s.a11y.published && s.a11y.gen == gen && s.a11y.pub.Live() {
		return
	}
	tree, err := acc.AccessibilityTree()
	if err != nil {
		return
	}
	if err := s.a11y.pub.Publish(tree); err != nil {
		return
	}
	s.a11y.gen = gen
	s.a11y.published = true
}

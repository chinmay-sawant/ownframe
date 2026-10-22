//go:build linux && !android && !ios && !js

package window

import "github.com/chinmay-sawant/ownframe/internal/host"

func (s *shell) publishA11y() {
	acc, ok := s.app.(host.Accessible)
	if !ok {
		return
	}
	tree, err := acc.AccessibilityTree()
	if err != nil {
		return
	}
	if err := s.a11y.pub.Publish(tree); err != nil {
		return
	}
}

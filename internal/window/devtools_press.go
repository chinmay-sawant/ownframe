package window

import "github.com/chinmay-sawant/ownframe/internal/clipboard"

// devPanelPress runs the action of the hit under the pointer. A click on
// empty panel space does nothing, so a pinned box stays pinned.
func (s *shell) devPanelPress(x, y int) {
	for _, hit := range s.dev.hits {
		if devInRect(hit.rect, float64(x), float64(y)) {
			s.devRun(hit)

			return
		}
	}
}

// devOpHitText is the displayed row for one operation pick.
func (s *shell) devOpHitText(arg int) string {
	for _, row := range s.devOpsRows() {
		if row.act == devActOp && row.arg == arg {
			return row.line.plain()
		}
	}

	return ""
}

// devRun applies one panel hit. An Ops row pick writes its text to the
// OS clipboard; the write always follows that explicit click, and there
// is no setting to turn it off, so keep the dock closed for sensitive
// output. See documentation/devtools.md and security.md.
func (s *shell) devRun(hit devHit) {
	switch hit.act {
	case devActMobileExpand:
		if s.dev.mobile {
			s.dev.mobileExpanded = !s.dev.mobileExpanded
			s.dev.scroll = 0
		}
	case devActTab:
		s.dev.tab = devTabAt(hit.arg)
		s.dev.scroll = 0
	case devActOpsToggle:
		s.dev.ops = !s.dev.ops
	case devActJSON:
		if s.dev.collapsed == nil {
			s.dev.collapsed = map[string]bool{}
		}

		s.dev.collapsed[hit.key] = !s.dev.collapsed[hit.key]
	case devActOp:
		text := s.devOpHitText(hit.arg)
		s.dev.opPin = false
		if s.dev.haveOp && s.dev.opPick == hit.arg {
			s.dev.haveOp = false
			if text != "" {
				clipboard.Write(text)
			}

			return
		}

		s.dev.opPick, s.dev.haveOp = hit.arg, true
		if text != "" {
			clipboard.Write(text)
		}
	case devActPicker:
		s.devSetPicker(!s.devPickerOn())
	case devActClose:
		_ = s.devToggle()
	}
}

// devResize follows the pointer with the dock's left edge. The page viewport
// follows on the next update, throttled like a window drag.
func (s *shell) devResize(x int) {
	s.moving = true

	w := float64(s.screenW - x)
	if w < devDockMin {
		w = devDockMin
	}

	if w > float64(s.screenW) {
		w = float64(s.screenW)
	}

	s.dev.dockW = w
}

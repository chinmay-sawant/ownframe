package window

import "github.com/hajimehoshi/ebiten/v2"

// devPickerOn reports whether the element picker mode is active.
func (s *shell) devPickerOn() bool {
	return s.dev.picker || s.dev.pickerMode || s.dev.inspect ||
		s.dev.inspectMode || s.dev.pick || s.dev.pickMode
}

// devSetPicker turns element picker mode on or off, keeping the public boolean
// aliases in sync for tests and panel state. Toggling off drops the hovered
// and pinned picks, so the picker's outline effect (and the entries it drives)
// disappears with the mode.
func (s *shell) devSetPicker(on bool) {
	s.dev.picker = on
	s.dev.pickerMode = on
	s.dev.inspect = on
	s.dev.inspectMode = on
	s.dev.pick = on
	s.dev.pickMode = on

	if !on {
		s.dev.haveHov = false
		s.dev.havePin = false

		if s.dev.opPin {
			s.dev.haveOp = false
			s.dev.opPin = false
		}
	}
}

// drawDevPick outlines the hovered and pinned boxes, one colour each, with a
// label over each.
func (s *shell) drawDevPick(screen *ebiten.Image) {
	if !s.devPickerOn() {
		return
	}

	if s.dev.haveHov {
		s.devOutline(screen, s.dev.hovered, devHoverInk)
	}

	if s.dev.havePin {
		s.devOutline(screen, s.dev.pinned, devPinInk)
	}
}

func (s *shell) devPickerHit() {
	pw := devTextWidth("Picker")
	_, px := devHeaderRight(s.dev.panel)
	s.dev.hits = append(s.dev.hits, devHit{
		rect: devRect{X: px, Y: s.dev.panel.Y, W: pw, H: devHeaderH},
		act:  devActPicker,
	})
}

package window

import "testing"

// TestDevHeaderHitsTouch checks the picker hit ends where the close hit
// starts, so no dead gap sits between Picker and F12, and the F12 label
// itself closes the dock.
func TestDevHeaderHitsTouch(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)
	s.devLayout()

	var picker, closer *devHit
	for i := range s.dev.hits {
		switch s.dev.hits[i].act {
		case devActPicker:
			hit := s.dev.hits[i]
			picker = &hit
		case devActClose:
			hit := s.dev.hits[i]
			closer = &hit
		}
	}

	if picker == nil || closer == nil {
		t.Fatalf("hits = %+v, want picker and close", s.dev.hits)
	}

	if picker.rect.X+picker.rect.W != closer.rect.X {
		t.Fatalf("picker [%v %v] and close [%v %v] leave a gap",
			picker.rect.X, picker.rect.X+picker.rect.W, closer.rect.X, closer.rect.X+closer.rect.W)
	}

	fx := closer.rect.X
	f12w := devTextWidth("F12")
	f12Center := fx + f12w/2
	if !devInRect(closer.rect, f12Center, closer.rect.Y+closer.rect.H/2) {
		t.Fatalf("F12 center %v not in close rect %+v", f12Center, closer.rect)
	}

	s.devPanelPress(int(f12Center), int(closer.rect.Y+closer.rect.H/2))
	if err := s.devSync(); err != nil {
		t.Fatal(err)
	}
	if s.dev.on || d.on {
		t.Fatal("pressing the F12 hint did not close the dock")
	}
}

// TestDevOwnsPointBlocksTitlebarDrag checks the dock owns its header and
// resize edge, so a custom-chrome draggable strip cannot start a window
// drag over the close key.
func TestDevOwnsPointBlocksTitlebarDrag(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)
	s.screenW, s.screenH = 1500, 900
	s.dev.on = true
	s.devLayout()

	cls, _ := devHeaderRight(s.dev.panel)
	xCenter := s.dev.panel.X + s.dev.panel.W - devPanelPad - devTextWidth("X")/2
	yCenter := s.dev.panel.Y + devHeaderH/2
	if !s.devOwnsPoint(int(xCenter), int(yCenter)) {
		t.Fatalf("close key %v,%v not owned by dock %+v", xCenter, yCenter, s.dev.panel)
	}
	_ = cls

	edgeX := int(s.dev.panel.X - 1)
	if !s.devOwnsPoint(edgeX, int(yCenter)) {
		t.Fatalf("resize edge %v not owned by dock %+v", edgeX, s.dev.panel)
	}

	pageX := int(s.dev.panel.X) - int(devDockGrabWidth) - 10
	if s.devOwnsPoint(pageX, int(yCenter)) {
		t.Fatalf("page titlebar %v owned by dock %+v", pageX, s.dev.panel)
	}

	s.dev.on = false
	if s.devOwnsPoint(int(xCenter), int(yCenter)) {
		t.Fatal("closed dock still owns the header point")
	}
}

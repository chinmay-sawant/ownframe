package window

import (
	"testing"
	"time"
)

// TestDevToolsPickSkipsClick checks the other Phase 3 exit: a click pins the
// box under the cursor without calling the page's click handler.
func TestDevToolsPickSkipsClick(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)

	s.devSetPicker(true)

	if err := s.devPick(20, 20, false); err != nil {
		t.Fatal(err)
	}

	if !s.dev.havePin || s.dev.pinned.ID != "known" {
		t.Fatalf("pinned = %+v", s.dev.pinned)
	}

	if d.clicked != 0 {
		t.Fatalf("page clicks = %d, want 0", d.clicked)
	}

	if err := s.devPick(20, 20, false); err != nil {
		t.Fatal(err)
	}

	if s.dev.havePin {
		t.Fatal("the same box did not clear the pin")
	}

	if err := s.devPick(20, 20, true); err != nil {
		t.Fatal(err)
	}

	if d.clicked != 1 {
		t.Fatalf("page clicks = %d, want 1 after alt", d.clicked)
	}
}

// TestDevToolsRefreshRecordsDrawTime checks the Phase 1 hook: the shell's
// draw time reaches the page stats through SetDrawTime.
func TestDevToolsRefreshRecordsDrawTime(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)
	s.dev.draw = 5 * time.Millisecond

	s.devRefresh()

	if d.draw != 5*time.Millisecond {
		t.Fatalf("draw time = %v", d.draw)
	}

	if s.dev.stats.LastDraw != 5*time.Millisecond {
		t.Fatalf("stats draw = %v", s.dev.stats.LastDraw)
	}
}

// TestDevToolsDropsStalePicks checks that a new generation drops a box that
// its boxes no longer hold.
func TestDevToolsDropsStalePicks(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)
	s.dev.haveHov, s.dev.hovered = true, d.boxes[0]
	s.dev.havePin, s.dev.pinned = true, d.boxes[0]

	d.boxes = nil
	d.redraws = 1

	s.devRefresh()

	if s.dev.haveHov || s.dev.havePin {
		t.Fatal("a stale pick survived the generation change")
	}
}

func TestDevToolsRefreshWithoutPicksSkipsBoxes(t *testing.T) {
	d := newDevScreen()
	s := newDevShell(d)
	d.redraws++
	s.devRefresh()
	if d.boxReads != 0 {
		t.Fatalf("refresh read boxes %d times without a selection", d.boxReads)
	}
}

func TestDevToolsPickerRequiresExplicitSelection(t *testing.T) {
	d := newDevScreen()
	s := newDevShell(d)
	s.devLayout()
	if s.devPickerOn() {
		t.Fatal("opening DevTools enabled the picker")
	}
	if err := s.devPick(20, 20, false); err != nil {
		t.Fatal(err)
	}
	if d.clicked != 1 || s.dev.havePin {
		t.Fatal("default DevTools state intercepted a page click")
	}
}

func TestDevToolsCloseDisarmsPicker(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		name := "unselected"
		if enabled {
			name = "selected"
		}
		t.Run(name, func(t *testing.T) {
			d := newDevScreen()
			s := newDevShell(d)
			s.devLayout()
			if enabled {
				for _, hit := range s.dev.hits {
					if hit.act == devActPicker {
						s.devPanelPress(int(hit.rect.X+hit.rect.W/2), int(hit.rect.Y+hit.rect.H/2))
						break
					}
				}
				if !s.devPickerOn() {
					t.Fatal("explicit Picker click did not enable picking")
				}
				if err := s.devPick(20, 20, false); err != nil {
					t.Fatal(err)
				}
				if !s.dev.havePin || d.clicked != 0 {
					t.Fatal("selected picker did not intercept the click")
				}
			}
			close, _ := devHeaderRight(s.dev.panel)
			s.devPanelPress(int(close.X+devClosePad+devTextWidth("X")/2), int(close.Y+close.H/2))
			if d.on {
				t.Fatal("the drawn X did not request close")
			}
			if err := s.devSync(); err != nil {
				t.Fatal(err)
			}
			if s.devPickerOn() || s.dev.havePin || s.dev.haveHov {
				t.Fatal("closing DevTools left picking active")
			}
			// Finish the X press, then use the normal closed-dock pointer path.
			s.mouseDown = true
			if _, err := s.pointerMouse(
				20, 20, 20, 20, false, true, false,
			); err != nil {
				t.Fatal(err)
			}
			if _, err := s.pointerMouse(
				20, 20, 20, 20, true, false, false,
			); err != nil {
				t.Fatal(err)
			}
			if _, err := s.pointerMouse(
				20, 20, 20, 20, false, true, false,
			); err != nil {
				t.Fatal(err)
			}
			if d.clicked != 1 {
				t.Fatal("closed DevTools intercepted a normal page click")
			}
			if err := s.devToggle(); err != nil {
				t.Fatal(err)
			}
			if err := s.devSync(); err != nil {
				t.Fatal(err)
			}
			if s.devPickerOn() {
				t.Fatal("reopening DevTools enabled picking without a Picker click")
			}
			if err := s.devPick(20, 20, false); err != nil {
				t.Fatal(err)
			}
			if d.clicked != 2 || s.dev.havePin {
				t.Fatal("reopened DevTools intercepted a normal page click")
			}
		})
	}
}

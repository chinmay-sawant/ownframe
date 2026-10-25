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

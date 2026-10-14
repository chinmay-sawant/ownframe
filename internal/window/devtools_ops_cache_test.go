package window

import (
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestDevOpsRowsStillGenerationNoAlloc checks the Ops tab's per-frame build:
// a still generation returns the same rows and allocates nothing, because the
// list holds one row per painted operation. Not parallel, so no sibling
// allocates during AllocsPerRun.
func TestDevOpsRowsStillGenerationNoAlloc(t *testing.T) {
	d := newDevScreen()
	d.redraws = 1
	s := newDevShell(d)
	s.display = devTestDisplay()
	s.dev.tab = devTabOps

	first := s.devOpsRows()
	got := first

	allocs := testing.AllocsPerRun(100, func() { got = s.devOpsRows() })

	if allocs != 0 || &got[0] != &first[0] {
		t.Fatalf("still generation allocs = %v, rebuilt = %v", allocs, &got[0] != &first[0])
	}

	s.drawDevTools(ebiten.NewImage(s.screenW, s.screenH))

	if &s.dev.opsCache.rows[0] != &first[0] {
		t.Fatal("the draw rebuilt the rows")
	}
}

// TestDevOpsRowsCacheKeys checks the list follows the generation and the
// outline toggle, and that generation zero never caches, so a fake screen
// keeps seeing fresh rows.
func TestDevOpsRowsCacheKeys(t *testing.T) {
	d := newDevScreen()
	d.redraws = 1
	s := newDevShell(d)
	s.display = devTestDisplay()

	off := s.devOpsRows()
	if !strings.HasPrefix(off[0].line.plain(), "[ ]") {
		t.Fatalf("toggle row = %q, want [ ]", off[0].line.plain())
	}

	if got := s.devOpsRows(); &got[0] != &off[0] {
		t.Fatal("the same inputs rebuilt the rows")
	}

	s.dev.ops = true

	on := s.devOpsRows()
	if &on[0] == &off[0] || !strings.HasPrefix(on[0].line.plain(), "[x]") {
		t.Fatalf("toggle row = %q, want [x]", on[0].line.plain())
	}

	d.redraws++
	s.display = nil

	if text := devRowsText(s.devOpsRows()); !strings.Contains(text, "bitmap fallback") {
		t.Fatalf("new generation rows = %q, want the fallback", text)
	}

	d.redraws = 0
	s.display = devTestDisplay()
	first, second := s.devOpsRows(), s.devOpsRows()
	if &first[0] == &second[0] {
		t.Fatal("generation zero cached the rows")
	}
}

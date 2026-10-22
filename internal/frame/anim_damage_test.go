package frame

import (
	"image"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

func TestDamageCoversOldAndNew(t *testing.T) {
	t.Parallel()

	d, op := cardDisplay()
	a := TranslateTo(d, []*page.DisplayOp{op}, 30, 0, 100*time.Millisecond, Linear)
	a.Step(t0)
	dmg, _ := a.Step(t0.Add(100 * time.Millisecond))

	// The card starts at 10..30 points, 12..41 CSS px, and ends at
	// 40..60 points, 52..81 CSS px.
	if !image.Pt(20, 20).In(dmg.Old) {
		t.Fatalf("old box %v misses the old card", dmg.Old)
	}

	if !image.Pt(70, 20).In(dmg.New) {
		t.Fatalf("new box %v misses the new card", dmg.New)
	}

	if image.Pt(70, 20).In(dmg.Old) || image.Pt(20, 20).In(dmg.New) {
		t.Fatalf("boxes cover the wrong sides: %+v", dmg)
	}
}

// marker records the boxes a tick asked to repaint.
type marker struct {
	rect image.Rectangle
	n    int
}

func (m *marker) MarkRect(r image.Rectangle) {
	m.rect = r
	m.n++
}

func TestTickMarksOldAndNew(t *testing.T) {
	t.Parallel()

	d, op := cardDisplay()
	a := TranslateTo(d, []*page.DisplayOp{op}, 30, 0, 100*time.Millisecond, Linear)

	var m marker
	if !a.Tick(&m, t0) {
		t.Fatal("finished at the start")
	}

	if a.Tick(&m, t0.Add(100*time.Millisecond)) {
		t.Fatal("still live at the end")
	}

	if m.n != 2 {
		t.Fatalf("marks = %d", m.n)
	}

	if !image.Pt(20, 20).In(m.rect) || !image.Pt(70, 20).In(m.rect) {
		t.Fatalf("marked %v", m.rect)
	}

	if a.Tick(&m, t0.Add(200*time.Millisecond)) {
		t.Fatal("live after the end")
	}

	if m.n != 2 {
		t.Fatalf("a finished tick marked again: %d", m.n)
	}
}

func TestZeroDurationFinishesNow(t *testing.T) {
	t.Parallel()

	d, op := cardDisplay()
	a := TranslateTo(d, []*page.DisplayOp{op}, 5, 0, 0, Linear)

	dmg, live := a.Step(t0)
	if live {
		t.Fatal("still live")
	}

	if op.X != 15 {
		t.Fatalf("x = %v", op.X)
	}

	if dmg.Empty() {
		t.Fatal("no damage")
	}
}

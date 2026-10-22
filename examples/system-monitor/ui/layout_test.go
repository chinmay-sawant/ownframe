package ui

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// boxByAction returns the first box whose action matches.
func boxByAction(t *testing.T, app *App, action string) ownframe.Box {
	t.Helper()

	for _, b := range app.page.Boxes() {
		if b.Action == action {
			return b
		}
	}

	t.Fatalf("no box with action %q", action)

	return ownframe.Box{}
}

// clickBox clicks the center of a box.
func clickBox(t *testing.T, app *App, b ownframe.Box) {
	t.Helper()

	if err := app.page.Click(context.Background(), b.X+b.W/2, b.Y+b.H/2); err != nil {
		t.Fatal(err)
	}
}

func TestOverviewLayout(t *testing.T) {
	app := newTestApp(t, &fakeSource{}, nil)
	d := app.page.Display()

	for _, id := range panelOrder {
		box, ok := boxByID(app.page.Boxes(), id+"-graph")
		if !ok {
			t.Fatalf("graph box %s missing", id)
		}

		if box.W < 100 || box.H < 40 {
			t.Fatalf("%s graph too small: %vx%v", id, box.W, box.H)
		}

		if bars := ownframe.Fills(d, box, accent); len(bars) != graphColsMax {
			t.Fatalf("%s graph bars = %d, want %d", id, len(bars), graphColsMax)
		}
	}

	cpu, _ := boxByID(app.page.Boxes(), "cpu-graph")
	mem, _ := boxByID(app.page.Boxes(), "mem-graph")
	disk, _ := boxByID(app.page.Boxes(), "disk-graph")

	if cpu.Y != mem.Y || disk.Y <= cpu.Y {
		t.Fatalf("cards are not a two-column grid: cpu.y=%v mem.y=%v disk.y=%v",
			cpu.Y, mem.Y, disk.Y)
	}
}

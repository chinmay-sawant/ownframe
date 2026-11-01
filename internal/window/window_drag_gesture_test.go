package window

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestDoubleClickPair(t *testing.T) {
	t.Parallel()

	now := time.Now()
	if !doubleClickPair(now.Add(-300*time.Millisecond), 100, 40, now, 103, 42) {
		t.Fatal("close press in time should pair")
	}
	if doubleClickPair(time.Time{}, 100, 40, now, 100, 40) {
		t.Fatal("zero previous press must not pair")
	}
	if doubleClickPair(now.Add(-600*time.Millisecond), 100, 40, now, 100, 40) {
		t.Fatal("slow second press is two single clicks")
	}
	if doubleClickPair(now.Add(-100*time.Millisecond), 100, 40, now, 200, 40) {
		t.Fatal("far second press is a new drag, not a pair")
	}
}

func TestDragPointControl(t *testing.T) {
	t.Parallel()

	// Settings-card shape: nested bare text under an id-bearing card
	// without an action must stay chrome, while the same nesting under
	// an action button is a control.
	boxes := []layout.Box{
		{ID: "settings-card", Tag: "div", X: 400, Y: 200, W: 580, H: 340, ParentIndex: -1},
		{ID: "settings-panel", Tag: "div", X: 450, Y: 260, W: 480, H: 220, ParentIndex: 0},
		{Tag: "span", Text: "Appearance", X: 470, Y: 300, W: 90, H: 16, ParentIndex: 1},
		{ID: "theme-dark", Tag: "button", Action: "theme:dark", X: 470, Y: 380, W: 440, H: 60, ParentIndex: 1},
		{Tag: "span", Text: "Dark", X: 490, Y: 392, W: 34, H: 16, ParentIndex: 3},
		{ID: "mail-search", Tag: "input", X: 900, Y: 30, W: 240, H: 36, ParentIndex: -1},
	}
	cases := []struct {
		name string
		x, y float64
		want bool
	}{
		{"card title text is chrome", 500, 308, false},
		{"card padding is chrome", 920, 500, false},
		{"option label resolves to its button", 500, 400, true},
		{"search field is a control", 950, 48, true},
		{"outside everything is chrome", 10, 10, false},
	}
	for _, tc := range cases {
		if got := dragPointControl(boxes, tc.x, tc.y); got != tc.want {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestDragPointControlBadParentIndex(t *testing.T) {
	t.Parallel()

	boxes := []layout.Box{
		{Tag: "span", Text: "x", X: 0, Y: 0, W: 10, H: 10, ParentIndex: 9},
	}
	if dragPointControl(boxes, 5, 5) {
		t.Fatal("dangling parent index must not count as a control")
	}
}

func TestDragPointControlIDLessAction(t *testing.T) {
	t.Parallel()

	// Titlebar shape: the close/min/max buttons carry a data-action but no
	// id. They must still count as controls so a press on them never starts
	// a window drag or a double-click maximize.
	boxes := []layout.Box{
		{Tag: "header", X: 0, Y: 0, W: 1120, H: 68, ParentIndex: -1},
		{Tag: "div", X: 900, Y: 10, W: 210, H: 48, ParentIndex: 0},
		{Tag: "button", Action: "win-close", X: 1060, Y: 16, W: 40, H: 36, ParentIndex: 1},
		{Tag: "span", Text: "×", X: 1072, Y: 24, W: 16, H: 20, ParentIndex: 2},
	}
	if !dragPointControl(boxes, 1065, 20) {
		t.Fatal("id-less action button must count as a control")
	}
	if !dragPointControl(boxes, 1080, 34) {
		t.Fatal("nested label must resolve to its id-less action button")
	}
	if dragPointControl(boxes, 500, 34) {
		t.Fatal("bare titlebar padding must stay chrome")
	}
}

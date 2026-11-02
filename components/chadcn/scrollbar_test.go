package chadcn_test

import (
	"context"
	"math"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/pagetest"
	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/theme"
)

func TestScrollGeometryUsesLogicalUnits(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name               string
		metrics            chadcn.ScrollMetrics
		track, top, height int
		visible            bool
	}{
		{"pixels", chadcn.ScrollMetrics{Content: 1000, Viewport: 200, Offset: 400}, 200, 80, 40, true},
		{"rows", chadcn.ScrollMetrics{Content: 100, Viewport: 20, Offset: 40}, 200, 80, 40, true},
		{"minimum", chadcn.ScrollMetrics{Content: 10000, Viewport: 1, Offset: 9999}, 200, 176, 24, true},
		{"tiny track", chadcn.ScrollMetrics{Content: 100, Viewport: 1}, 10, 0, 10, true},
		{"fits", chadcn.ScrollMetrics{Content: 200, Viewport: 200}, 200, 0, 0, false},
		{"unmeasured", chadcn.ScrollMetrics{Content: 100}, 200, 0, 0, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			top, height, visible := tt.metrics.Thumb(tt.track)
			if top != tt.top || height != tt.height || visible != tt.visible {
				t.Fatalf("thumb=(%d,%d,%v)", top, height, visible)
			}
		})
	}
}

func TestScrollFractionalInputAndBounds(t *testing.T) {
	t.Parallel()
	var state chadcn.ScrollState
	state.SetRange(20, 5)
	for range 3 {
		if state.Wheel(-.25, 1) {
			t.Fatal("fraction advanced too early")
		}
	}
	if !state.Wheel(-.25, 1) || state.Offset() != 1 {
		t.Fatal("fractional wheel did not accumulate")
	}
	snapshot := state.Metrics()
	state.ScrollTo(15)
	state.Wheel(-.75, 1)
	if !state.Wheel(1, 1) || state.Offset() != 14 {
		t.Fatal("outward fraction delayed movement away from end")
	}
	if snapshot.Offset != 1 {
		t.Fatal("metrics snapshot changed")
	}
	for _, input := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		if state.Wheel(input, 1) || state.Wheel(1, input) {
			t.Fatal("nonfinite input advanced range")
		}
	}
	state.ScrollBy(int(^uint(0) >> 1))
	if state.Offset() != 15 {
		t.Fatal("large positive delta did not clamp")
	}
	state.ScrollBy(-int(^uint(0)>>1) - 1)
	if state.Offset() != 0 {
		t.Fatal("large negative delta did not clamp")
	}
	state.ScrollTo(15)
	state.SetRange(7, 5)
	if state.Offset() != 2 {
		t.Fatal("shrinking range did not clamp")
	}
	state.SetRange(-1, -1)
	if state.Metrics() != (chadcn.ScrollMetrics{}) {
		t.Fatal("negative extents not normalized")
	}
}

func TestScrollDragKeyboardAndResize(t *testing.T) {
	t.Parallel()
	var state chadcn.ScrollState
	state.SetRange(100, 20)
	if !state.BeginDrag(200) || !state.DragBy(80, 200) || state.Offset() != 40 {
		t.Fatal("drag did not map thumb travel to logical offset")
	}
	state.EndDrag()
	if state.DragBy(20, 200) {
		t.Fatal("released thumb still moved")
	}
	state.BeginDrag(200)
	if state.DragBy(10, 180) {
		t.Fatal("resized track retained stale drag")
	}
	state.BeginDrag(200)
	state.SetRange(200, 20)
	if state.DragBy(10, 200) {
		t.Fatal("changed extents retained stale drag")
	}
	for _, tt := range []struct {
		key  string
		want int
	}{{"end", 180}, {"pageup", 160}, {"arrowup", 157}, {"home", 0}, {"pagedown", 20}, {"arrowdown", 23}} {
		if !state.Key(tt.key, 3) || state.Offset() != tt.want {
			t.Fatalf("key %s: offset=%d want=%d", tt.key, state.Offset(), tt.want)
		}
	}
	if state.Key("escape", 3) {
		t.Fatal("unrelated key changed range")
	}
}

func TestScrollbarTrackClickPublishesStateAndTheme(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var state chadcn.ScrollState
	state.SetRange(1000, 200)
	var view *ui.View
	var tree func() *ui.Element
	tree = func() *ui.Element {
		return ui.Div().Position(ui.PositionRelative).Width(ui.Px(300)).Height(ui.Px(200)).Children(
			chadcn.Scrollbar(chadcn.ScrollbarOptions{ID: "scroll", ControlsID: "content", Label: "Messages", State: &state, TrackPixels: 200,
				Overrides: theme.Overrides{"--of-scrollbar": "#ff0000", "--of-scrollbar-width": "12px"},
				OnChange:  func(ctx context.Context) error { return view.Update(ctx, tree()) },
			}),
		)
	}
	var err error
	view, err = ui.Mount(ownframe.Config{Width: 400, Height: 300, Theme: theme.Dark}, tree())
	if err != nil {
		t.Fatal(err)
	}
	if err := view.Page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	thumb, ok := pagetest.BoxByID(view.Page, "scroll-thumb")
	if !ok || thumb.W != 12 || thumb.H != 40 {
		t.Fatalf("thumb=%+v present=%v", thumb, ok)
	}
	if err := pagetest.Click(ctx, view.Page, "scroll-after"); err != nil {
		t.Fatal(err)
	}
	if state.Offset() != 200 || !strings.Contains(view.Page.HTML(), `aria-valuenow="200"`) {
		t.Fatal("track paging failed to publish new offset")
	}
	for _, attr := range []string{`role="scrollbar"`, `aria-controls="content"`, `aria-valuemax="800"`, `aria-label="Messages"`} {
		if !strings.Contains(view.Page.HTML(), attr) {
			t.Fatalf("missing %s", attr)
		}
	}
	state.SetRange(100, 200)
	if chadcn.Scrollbar(chadcn.ScrollbarOptions{State: &state, TrackPixels: 200}) != nil {
		t.Fatal("fitting range rendered a scrollbar")
	}
}

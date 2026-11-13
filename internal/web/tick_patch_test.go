package web

import (
	"context"
	"image"
	"strings"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// patchScreen is a retained-op ticker: damage with a frozen generation,
// like a seek-bar tick that patches ops in place.
type patchScreen struct {
	host.Screen
	generation uint64
	redraws    int
	dirty      bool
	rect       image.Rectangle
	text       string
}

func (f *patchScreen) Tick(context.Context) error { return nil }

func (f *patchScreen) FrameDirty() bool {
	dirty := f.dirty
	f.dirty = false
	return dirty
}

func (f *patchScreen) TakeDirtyRects() ([]image.Rectangle, bool) {
	if f.rect.Empty() {
		return nil, false
	}
	return []image.Rectangle{f.rect}, true
}

func (f *patchScreen) Generation() uint64 { return f.generation }

func (f *patchScreen) Size() (int, int) { return 320, 200 }

func (f *patchScreen) Display() *layout.Display {
	return &layout.Display{
		Width:  320,
		Height: 200,
		Ops: []layout.DisplayOp{
			{ID: 7, Kind: layout.DisplayOpText, Text: f.text, X: 11, Y: 21, W: 50, H: 10},
		},
	}
}

func (f *patchScreen) Boxes() []layout.Box {
	return []layout.Box{{ID: "elapsed", X: 10, Y: 20, W: 100, H: 30, ParentIndex: -1, OperationIDs: []uint64{7}}}
}

func (f *patchScreen) Redraw(context.Context) error {
	f.redraws++
	return nil
}

func TestTickOnceSkipsRedrawOnInPlaceDamage(t *testing.T) {
	t.Parallel()

	app := &patchScreen{generation: 5, dirty: true, rect: image.Rect(0, 0, 200, 200), text: "1:24"}
	srv := &server{app: app, subs: map[chan []byte]struct{}{}}
	ch := srv.subscribe(1)
	defer srv.unsubscribe(ch)
	select {
	case <-ch:
	default:
	}

	if err := srv.tickOnce(t.Context(), app); err != nil {
		t.Fatalf("tickOnce: %v", err)
	}

	if app.redraws != 0 {
		t.Fatalf("redraws = %d, want 0 (in-place tick must not Redraw)", app.redraws)
	}
	if srv.frameGen == 0 {
		t.Fatal("frameGen did not advance on in-place damage")
	}

	select {
	case msg := <-ch:
		body := string(msg)
		if !strings.Contains(body, `"domGeneration":5`) {
			t.Fatalf("event moved domGeneration: %q", body)
		}
		if !strings.Contains(body, `"elapsed"`) || !strings.Contains(body, "1:24") {
			t.Fatalf("event missing text patch: %q", body)
		}
	default:
		t.Fatal("in-place tick published no event")
	}
}

func TestCollectTickPatchesDeltaCache(t *testing.T) {
	t.Parallel()

	display := &layout.Display{
		Ops: []layout.DisplayOp{
			{ID: 7, Kind: layout.DisplayOpText, Text: "1:24", X: 11, Y: 21, W: 50, H: 10},
		},
	}
	boxes := []layout.Box{{ID: "elapsed", X: 10, Y: 20, W: 100, H: 30, ParentIndex: -1, OperationIDs: []uint64{7}}}
	rects := []image.Rectangle{image.Rect(0, 0, 200, 200)}
	cache := map[string][2]string{}

	got := collectTickPatches(display, boxes, rects, cache, map[string]bool{"elapsed": true}, nil, map[string]bool{})
	if len(got) != 1 || got[0].ID != "elapsed" || got[0].Text == nil || *got[0].Text != "1:24" {
		t.Fatalf("patches = %+v", got)
	}

	if again := collectTickPatches(display, boxes, rects, cache, map[string]bool{"elapsed": true}, nil, map[string]bool{}); len(again) != 0 {
		t.Fatalf("unchanged tick re-emitted: %+v", again)
	}

	display.Ops[0].Text = "1:25"
	got = collectTickPatches(display, boxes, rects, cache, map[string]bool{"elapsed": true}, nil, map[string]bool{})
	if len(got) != 1 || got[0].Text == nil || *got[0].Text != "1:25" {
		t.Fatalf("changed tick patches = %+v", got)
	}
}

func TestCollectTickPatchesSkipsContainers(t *testing.T) {
	t.Parallel()

	display := &layout.Display{
		Ops: []layout.DisplayOp{
			{ID: 7, Kind: layout.DisplayOpText, Text: "1:24", X: 11, Y: 21, W: 50, H: 10},
			{ID: 8, Kind: layout.DisplayOpFillRect, X: 10, Y: 20, W: 200, H: 30},
		},
	}
	boxes := []layout.Box{
		{ID: "row", X: 0, Y: 0, W: 400, H: 60, ParentIndex: -1, OperationIDs: []uint64{7, 8}},
		{ID: "elapsed", X: 10, Y: 20, W: 100, H: 30, ParentIndex: 0, OperationIDs: []uint64{7}},
	}
	rects := []image.Rectangle{image.Rect(0, 0, 400, 60)}
	cache := map[string][2]string{}
	styleIDs := map[string]bool{"row": true, "elapsed": true}

	got := collectTickPatches(display, boxes, rects, cache, styleIDs, nil, map[string]bool{})
	for _, p := range got {
		if p.ID == "row" {
			t.Fatalf("container box patched: %+v", got)
		}
	}
}

func TestCollectTickPatchesSkipsNestedSourceContainer(t *testing.T) {
	t.Parallel()

	// row-4 is not box-parented to its tiles, but its source element
	// contains child elements: patching it would destroy the tiles.
	source := `<div id="row-4"><span id="t-4-stat-0">On time</span></div>`
	display := &layout.Display{
		Ops: []layout.DisplayOp{
			{ID: 9, Kind: layout.DisplayOpText, Text: "On time", X: 11, Y: 21, W: 50, H: 10},
		},
	}
	boxes := []layout.Box{
		{ID: "row-4", X: 0, Y: 0, W: 400, H: 60, ParentIndex: -1, OperationIDs: []uint64{9}},
		{ID: "t-4-stat-0", X: 10, Y: 20, W: 100, H: 30, ParentIndex: -1, OperationIDs: []uint64{9}},
	}
	rects := []image.Rectangle{image.Rect(0, 0, 400, 60)}
	containers := containerIDs(source)
	if !containers["row-4"] {
		t.Fatal("row-4 not detected as container")
	}
	if containers["t-4-stat-0"] {
		t.Fatal("tile span detected as container")
	}

	got := collectTickPatches(display, boxes, rects, map[string][2]string{}, map[string]bool{}, nil, containers)
	if len(got) != 1 || got[0].ID != "t-4-stat-0" {
		t.Fatalf("patches = %+v, want only the tile", got)
	}
}

func TestCollectTickPatchesSkipsUnstyledFill(t *testing.T) {
	t.Parallel()

	display := &layout.Display{
		Ops: []layout.DisplayOp{
			{ID: 8, Kind: layout.DisplayOpFillRect, X: 10, Y: 20, W: 60, H: 30},
		},
	}
	boxes := []layout.Box{
		{ID: "cover", X: 10, Y: 20, W: 100, H: 30, ParentIndex: -1, OperationIDs: []uint64{8}},
	}
	rects := []image.Rectangle{image.Rect(0, 0, 400, 60)}

	if got := collectTickPatches(display, boxes, rects, map[string][2]string{}, map[string]bool{}, nil, map[string]bool{}); len(got) != 0 {
		t.Fatalf("unstyled fill patched: %+v", got)
	}
}

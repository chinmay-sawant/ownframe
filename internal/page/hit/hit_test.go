package hit

import (
	"math"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

func overlapBoxes() []layout.Box {
	return []layout.Box{
		{ID: "outer", X: 0, Y: 0, W: 100, H: 100},
		{ID: "inner", X: 10, Y: 10, W: 20, H: 20},
		{ID: "", X: 12, Y: 12, W: 2, H: 2},
	}
}

func TestScanOverlappingBoxes(t *testing.T) {
	t.Parallel()

	boxes := overlapBoxes()
	spot := Scan(boxes, 13, 13)
	if !spot.OK || spot.Box.ID != "" || spot.ID != "inner" {
		t.Fatalf("scan = %+v, want the anonymous box and id inner", spot)
	}

	box, ok := Hit(boxes, 13, 13)
	if !ok || box.ID != "" {
		t.Fatalf("hit = %+v, %v, want the anonymous box", box, ok)
	}

	spot = Scan(boxes, 11, 11)
	if !spot.OK || spot.Box.ID != "inner" || spot.ID != "inner" {
		t.Fatalf("scan = %+v, want inner", spot)
	}

	if spot := Scan(boxes, 200, 200); spot.OK || spot.ID != "" {
		t.Fatalf("miss = %+v, want nothing", spot)
	}
}

func TestHitKeepsFarEdge(t *testing.T) {
	t.Parallel()

	boxes := []layout.Box{{ID: "e", X: 0, Y: 0, W: 10, H: 10}}
	if spot := Scan(boxes, 10, 10); !spot.OK || spot.ID != "e" {
		t.Fatalf("far edge = %+v, want inside", spot)
	}
	if spot := Scan(boxes, 10.1, 10); spot.OK {
		t.Fatalf("past far edge = %+v, want outside", spot)
	}
	if spot := Scan(boxes, math.NaN(), 5); !spot.OK || spot.ID != "e" {
		t.Fatalf("NaN = %+v, want the original reject to keep it inside", spot)
	}
}

func TestIndexLastBox(t *testing.T) {
	t.Parallel()

	first := []layout.Box{
		{ID: "dup", X: 1, W: 1, H: 1},
		{ID: "dup", X: 2, W: 1, H: 1},
		{ID: "k", W: 1, H: 1},
	}
	ix := Index(nil, first)
	if !Has(ix, "dup") || Has(ix, "missing") {
		t.Fatal("Has did not answer from the index")
	}
	if got := ByID(ix, "dup"); got.X != 2 {
		t.Fatalf("ByID = %+v, want the last box", got)
	}
	if got := ByID(ix, "missing"); got.ID != "missing" || got.X != 0 {
		t.Fatalf("ByID = %+v, want Box{ID: missing}", got)
	}

	ix = Index(ix, []layout.Box{
		{ID: "", X: 7, W: 1, H: 1},
		{ID: "k", W: 1, H: 1},
		{ID: "", X: 8, W: 1, H: 1},
	})
	if Has(ix, "dup") {
		t.Fatal("the index kept a stale entry")
	}
	if got := ByID(ix, ""); got.X != 8 {
		t.Fatalf("anonymous lookup = %+v, want the last anonymous box", got)
	}
}

func TestIndexRebuildReusesStorage(t *testing.T) {
	boxes := []layout.Box{{ID: "first"}, {ID: "second"}}
	ix := Index(nil, boxes)
	allocs := testing.AllocsPerRun(100, func() {
		ix = Index(ix, boxes)
	})
	if allocs != 0 {
		t.Fatalf("index rebuild allocated %g times, want zero", allocs)
	}
}

func TestListRebuildReusesStorage(t *testing.T) {
	boxes, items := fixedStack()
	first := &items[0]
	boxes[0].X = 100
	allocs := testing.AllocsPerRun(100, func() {
		items = List(items, boxes, nil)
	})
	if allocs != 0 {
		t.Fatalf("list rebuild allocated %g times, want zero", allocs)
	}
	if &items[0] != first || items[0].Box.X != 100 {
		t.Fatal("list rebuild did not reuse storage with updated geometry")
	}
	for _, item := range items {
		if item.Fixed {
			t.Fatal("list rebuild kept a stale fixed flag")
		}
	}
	items = List(items, boxes[:1], boxes[:1])
	if len(items) != 1 || !items[0].Fixed {
		t.Fatal("shortened list did not update its fixed root")
	}
	for _, item := range items[:cap(items)][1:] {
		if item.Box.ID != "" || item.Fixed {
			t.Fatal("shortened list retained a removed box")
		}
	}
}

func fixedStack() ([]layout.Box, []Item) {
	boxes := []layout.Box{
		{ID: "head", ParentIndex: -1, X: 0, Y: 0, W: 100, H: 50},
		{ID: "go", ParentIndex: 0, X: 10, Y: 10, W: 40, H: 20},
		{ID: "body", ParentIndex: -1, X: 0, Y: 0, W: 100, H: 80},
	}

	return boxes, List(nil, boxes, []layout.Box{boxes[0]})
}

func TestListPinsFixedSubtree(t *testing.T) {
	t.Parallel()

	_, items := fixedStack()
	if len(items) != 3 || !items[0].Fixed || !items[1].Fixed || items[2].Fixed {
		t.Fatalf("fixed flags = %v %v %v, want head and its child", items[0].Fixed, items[1].Fixed, items[2].Fixed)
	}
	if !IsFixed(items, items[1].Box) || IsFixed(items, items[2].Box) || IsFixed(items, layout.Box{}) {
		t.Fatal("IsFixed disagreed with the list")
	}
}

func TestScanListFixedWins(t *testing.T) {
	t.Parallel()

	_, items := fixedStack()
	spot := ScanList(items, 13, 13, 0, 0)
	if !spot.OK || spot.ID != "go" || spot.Box.ID != "go" {
		t.Fatalf("pinned = %+v, want the control inside the fixed box", spot)
	}

	spot = ScanList(items, 10, 60, 0, 0)
	if !spot.OK || spot.ID != "body" {
		t.Fatalf("content = %+v, want body", spot)
	}

	// The fixed header is at the viewport origin. Scrolling it off the
	// content point leaves the content box, which is later in the list.
	spot = ScanList(items, 13, 13, 0, 100)
	if !spot.OK || spot.ID != "body" {
		t.Fatalf("scrolled = %+v, want body", spot)
	}
}

func TestFixedBoxesRoots(t *testing.T) {
	t.Parallel()

	if FixedBoxes(nil) != nil {
		t.Fatal("nil display returned boxes")
	}

	d := &layout.Display{
		Boxes: []layout.Box{
			{ID: "page", ParentIndex: -1},
			{ID: "head", ParentIndex: 0},
			{ID: "go", ParentIndex: 1},
			{ID: "body", ParentIndex: 0},
		},
		Ops: []layout.DisplayOp{
			{},
			{Fixed: true},
			{Fixed: true},
			{Fixed: true},
			{},
			{Fixed: true},
		},
		OpOwners: []int{0, 1, 2, 0, 3, 3},
	}
	got := FixedBoxes(d)
	if len(got) != 2 || got[0].ID != "head" || got[1].ID != "body" {
		t.Fatalf("roots = %+v, want head then body", got)
	}
}

func TestListWithoutFixedCopiesDocumentOrder(t *testing.T) {
	t.Parallel()

	boxes := overlapBoxes()
	items := List(nil, boxes, nil)
	if len(items) != len(boxes) {
		t.Fatalf("len = %d, want %d", len(items), len(boxes))
	}
	for i := range boxes {
		if items[i].Fixed || items[i].Box.ID != boxes[i].ID {
			t.Fatalf("item %d = %+v", i, items[i])
		}
	}
}

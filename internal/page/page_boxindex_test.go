package page

import (
	"strconv"
	"testing"
)

// TestBoxIndexLeavesNoScan puts thousands of boxes in the page, then clears
// the slice a per-lookup scan would walk. Only the id index can answer. It
// must keep the last-box-wins rule and the old answers for anonymous boxes
// and unknown ids, and a later setBoxes must drop stale entries.
func TestBoxIndexLeavesNoScan(t *testing.T) {
	p, err := New(Config{HTML: `<div id="x"></div>`, Width: 100, Height: 100})
	if err != nil {
		t.Fatal(err)
	}

	boxes := make([]Box, 0, 4000)
	for i := 0; i < 4000; i++ {
		boxes = append(boxes, Box{ID: "b" + strconv.Itoa(i), W: 1, H: 1})
	}

	boxes = append(boxes, Box{ID: "dup", X: 1, W: 1, H: 1})
	boxes = append(boxes, Box{ID: "dup", X: 2, W: 1, H: 1})
	p.setBoxes(boxes)
	p.boxes = nil

	if !p.hasBox("dup") || p.hasBox("missing") {
		t.Fatal("hasBox did not answer from the index")
	}

	if got := p.boxByID("dup"); got.X != 2 {
		t.Fatalf("boxByID = %+v, want the last box", got)
	}

	if got := p.boxByID("b3999"); got.ID != "b3999" {
		t.Fatalf("boxByID = %+v, want b3999", got)
	}

	if got := p.boxByID("missing"); got.ID != "missing" || got.X != 0 {
		t.Fatalf("boxByID = %+v, want Box{ID: missing}", got)
	}

	p.setBoxes([]Box{
		{ID: "", X: 7, W: 1, H: 1},
		{ID: "k", W: 1, H: 1},
		{ID: "", X: 8, W: 1, H: 1},
	})

	if got := p.boxByID(""); got.X != 8 {
		t.Fatalf("anonymous lookup = %+v, want the last anonymous box", got)
	}

	if p.hasBox("dup") {
		t.Fatal("the index kept a stale entry")
	}
}

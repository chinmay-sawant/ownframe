package frame

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

// pathFillPage is a bevel-cornered box with a uniform border. The engine paints
// it as two DisplayOpPath ops, so the fill helper has to see a shaped fill.
const pathFillPage = `<body style="margin:0;background:#ffffff">` +
	`<div id="shaped" style="position:absolute;left:20px;top:20px;width:120px;height:80px;` +
	`background:#3060c0;border:4px solid #102030;border-radius:24px;` +
	`corner-shape:bevel"></div></body>`

// TestFillFindsACornerPathFill checks that a Go-driven helper can find the
// shaped background and that a path border is not mistaken for a fill.
func TestFillFindsACornerPathFill(t *testing.T) {
	t.Parallel()

	display, err := render.DisplayList(context.Background(), pathFillPage, 200, 140)
	if err != nil {
		t.Fatal(err)
	}

	boxIndex := -1

	for i := range display.Boxes {
		if display.Boxes[i].ID == "shaped" {
			boxIndex = i
		}
	}

	if boxIndex < 0 {
		t.Fatal("shaped box is missing from the display")
	}

	var fill, stroke *page.DisplayOp

	for i := range display.Ops {
		op := &display.Ops[i]

		corner, ok := op.CornerPathValue()
		if op.Kind != render.OpPath || !ok || corner == nil {
			continue
		}

		if corner.Stroke {
			stroke = op
		} else {
			fill = op
		}
	}

	if fill == nil || stroke == nil {
		t.Fatalf("want a fill and a stroke path, got fill %v and stroke %v", fill, stroke)
	}

	box := display.Boxes[boxIndex]

	if got := Fill(display, box, [3]float64{fill.R, fill.G, fill.B}); got != fill {
		t.Fatalf("Fill = %p, want the path fill %p", got, fill)
	}

	if got := Fill(display, box, [3]float64{stroke.R, stroke.G, stroke.B}); got != nil {
		t.Fatalf("a path border matched a fill query: %+v", got)
	}
}

package cssmotion_test

import (
	"context"
	"slices"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

// pathOwnerPage is a bevel-cornered box with a uniform border. The engine paints
// it as two DisplayOpPath ops, one background fill and one border stroke.
const pathOwnerPage = `<body style="margin:0;background:#ffffff">` +
	`<div id="shaped" style="position:absolute;left:20px;top:20px;width:120px;height:80px;` +
	`background:#3060c0;border:4px solid #102030;border-radius:24px;` +
	`corner-shape:bevel"></div></body>`

// TestSelectSeparatesPathFillAndStroke checks that a colour property reaches a
// corner-shape path: a background recolours the fill half and a border or
// outline recolours the stroke half, even though both carry DisplayOpPath.
func TestSelectSeparatesPathFillAndStroke(t *testing.T) {
	t.Parallel()

	display, err := render.DisplayList(context.Background(), pathOwnerPage, 200, 140)
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

	fill, stroke := -1, -1

	for i := range display.Ops {
		corner, ok := display.Ops[i].CornerPathValue()
		if display.Ops[i].Kind != render.OpPath || !ok || corner == nil {
			continue
		}

		if corner.Stroke {
			stroke = i
		} else {
			fill = i
		}
	}

	if fill < 0 || stroke < 0 {
		t.Fatalf("want a fill and a stroke path, got fill %d and stroke %d", fill, stroke)
	}

	box := display.Boxes[boxIndex]

	if got := cssmotion.Select(display, box, boxIndex, cssmotion.KindFill); !slices.Equal(got, []int{fill}) {
		t.Fatalf("background targets = %v, want the path fill %d", got, fill)
	}

	if got := cssmotion.Select(display, box, boxIndex, cssmotion.KindStroke); !slices.Equal(got, []int{stroke}) {
		t.Fatalf("border targets = %v, want the path stroke %d", got, stroke)
	}

	if got := cssmotion.Select(display, box, boxIndex, cssmotion.KindInk); len(got) != 0 {
		t.Fatalf("a shaped box with no text targeted ink %v", got)
	}

	if got := cssmotion.Select(display, box, boxIndex, cssmotion.KindSubtree); !slices.Contains(got, fill) || !slices.Contains(got, stroke) {
		t.Fatalf("subtree targets = %v, want both path halves", got)
	}
}

package render_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/page/draw"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

// cornerShapePage is a bevel-cornered box with a uniform border, the shape the
// engine paints as DisplayOpPath instead of a radiused fill and stroke pair.
const cornerShapePage = `<body style="margin:0;background:#ffffff">` +
	`<div style="position:absolute;left:20px;top:20px;width:120px;height:80px;` +
	`background:#3060c0;border:4px solid #102030;border-radius:24px;` +
	`corner-shape:bevel"></div></body>`

// TestCornerShapePageStaysOnTheListPath checks that a corner-shape page keeps a
// non-round box on the live list: the display carries path ops with a command
// list, Replayable accepts them, and the page chooser selects the list path.
// Before path replay, one OpPath sent the whole page to the bitmap.
func TestCornerShapePageStaysOnTheListPath(t *testing.T) {
	t.Parallel()

	display, err := render.DisplayList(context.Background(), cornerShapePage, 200, 140)
	if err != nil {
		t.Fatal(err)
	}

	var fills, strokes int

	for index := range display.Ops {
		op := &display.Ops[index]
		if op.Kind != render.OpPath {
			continue
		}

		corner, ok := op.CornerPathValue()
		if !ok || corner == nil {
			t.Fatalf("path op %d carries no command list", index)
		}

		if len(corner.Cmds) < 4 {
			t.Fatalf("path op %d has %d commands", index, len(corner.Cmds))
		}

		if corner.Cmds[0].Verb != layout.PathMoveTo || corner.Cmds[len(corner.Cmds)-1].Verb != layout.PathClose {
			t.Fatalf("path op %d is not a closed contour", index)
		}

		if corner.Stroke {
			strokes++

			if op.Width <= 0 {
				t.Errorf("border path %d has width %v", index, op.Width)
			}

			continue
		}

		fills++
	}

	if fills == 0 || strokes == 0 {
		t.Fatalf("want a fill and a border path, got %d fills and %d strokes", fills, strokes)
	}

	if !render.Replayable(display) {
		t.Fatal("corner-shape page should replay")
	}

	if got := draw.Choose(display); got != draw.PathList {
		t.Fatalf("corner-shape page chose %v, want the list path", got)
	}
}

// TestReplayableRejectsPathWithoutCommands keeps an empty payload on the
// bitmap path: there is no geometry to replay, and guessing one would paint a
// shape the engine never emitted.
func TestReplayableRejectsPathWithoutCommands(t *testing.T) {
	t.Parallel()

	display, err := render.DisplayList(
		context.Background(),
		`<div style="width:40px;height:40px;background:#3060c0"></div>`,
		120,
		80,
	)
	if err != nil {
		t.Fatal(err)
	}

	display.Ops = append(display.Ops, render.DisplayOp{Kind: render.OpPath})

	if render.Replayable(display) {
		t.Fatal("a path without commands should not replay")
	}
}

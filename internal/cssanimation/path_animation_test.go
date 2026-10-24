package cssanimation_test

import (
	"context"
	"slices"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/cssanimation"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

// pathAnimationPage is a bevel-cornered box with a uniform border, the page the
// engine paints as two corner paths.
const pathAnimationPage = `<body style="margin:0;background:#ffffff">` +
	`<div id="shaped" style="position:absolute;left:20px;top:20px;width:120px;height:80px;` +
	`background:#3060c0;border:4px solid #102030;border-radius:24px;` +
	`corner-shape:bevel"></div></body>`

// TestBackgroundRecoloursACornerPath checks the shaped contract end to end: a
// background-color keyframe binds to the corner path's fill op, the sample
// rewrites its colour, and the border path and the command list stay put.
func TestBackgroundRecoloursACornerPath(t *testing.T) {
	t.Parallel()

	display, err := render.DisplayList(context.Background(), pathAnimationPage, 200, 140)
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

	display.Boxes[boxIndex].Animations = []layout.Animation{running("tint")}
	display.Keyframes = colorRule("tint", "background-color", "#ff0000", "#00ff00")

	bounds, gaps, err := cssanimation.BindBox(display, display.Boxes[boxIndex], boxIndex, started())
	if err != nil {
		t.Fatalf("bind: %v", err)
	}

	if len(gaps) != 0 {
		t.Fatalf("gaps = %v, want none", gaps)
	}

	if len(bounds) != 1 {
		t.Fatalf("players = %d, want one fill player", len(bounds))
	}

	if !slices.Equal(bounds[0].Indices, []int{fill}) {
		t.Fatalf("background targets = %v, want the path fill %d", bounds[0].Indices, fill)
	}

	if bounds[0].Mask != cssmotion.MaskRGB {
		t.Fatalf("background mask = %b, want RGB only", bounds[0].Mask)
	}

	strokeR, strokeG, strokeB := display.Ops[stroke].R, display.Ops[stroke].G, display.Ops[stroke].B

	bounds[0].Player.Step(halfWay())

	if got := display.Ops[fill]; got.R != .5 || got.G != .5 || got.B != 0 {
		t.Fatalf("path fill = %v,%v,%v, want the sRGB midpoint .5,.5,0", got.R, got.G, got.B)
	}

	if got := display.Ops[stroke]; got.R != strokeR || got.G != strokeG || got.B != strokeB {
		t.Fatalf("path border = %v,%v,%v, want its own %v,%v,%v",
			got.R, got.G, got.B, strokeR, strokeG, strokeB)
	}

	corner, ok := display.Ops[fill].CornerPathValue()
	if !ok || corner == nil || len(corner.Cmds) == 0 {
		t.Fatal("the colour sample dropped the path command list")
	}
}

package render_test

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

// TestReplayAcceptsTransforms pins the transformed page: a rotated box's
// fill, its four border lines, and its text all carry the baked matrix, and
// the replay painter maps each through it.
func TestReplayAcceptsTransforms(t *testing.T) {
	t.Parallel()

	display := acceptDisplay(t, `<div style="width:120px;height:40px;background:#eee;`+
		`border:2px solid #333;transform:rotate(12deg)">rotated</div>`)

	if !render.Replayable(display) {
		t.Fatal("transformed fill, stroke, and text should replay")
	}

	kinds := map[render.Kind]bool{}

	for i := range display.Ops {
		op := &display.Ops[i]
		if op.XformSet {
			kinds[op.Kind] = true
		}
	}

	if !kinds[render.OpFillRect] || !kinds[render.OpLine] || !kinds[render.OpText] {
		t.Fatalf("transformed kinds = %v, want fill, line, and text", kinds)
	}
}

// TestReplayAcceptsNestedBlendGroups pins a two-layer group: an element with
// mix-blend-mode:multiply holding a child with mix-blend-mode:screen. Every
// mode is separable, so the group compositor draws both layers.
func TestReplayAcceptsNestedBlendGroups(t *testing.T) {
	t.Parallel()

	display := acceptDisplay(t, `<body style="margin:0;background:#fff">`+
		`<div style="width:80px;height:80px;background:#369;mix-blend-mode:multiply">`+
		`<div style="width:40px;height:40px;background:#f33;mix-blend-mode:screen">`+
		`inner</div></div></body>`)

	if !render.Replayable(display) {
		t.Fatal("nested separable blend groups should replay")
	}

	seen := map[int]string{}

	for i := range display.Ops {
		for g := display.Ops[i].Group(); g != nil; g = g.Parent {
			seen[g.ID] = g.Mode
		}
	}

	if len(seen) != 2 {
		t.Fatalf("groups = %v, want the multiply and screen layers", seen)
	}

	for _, mode := range seen {
		if mode != "multiply" && mode != "screen" {
			t.Fatalf("group mode %q is not the page's blend", mode)
		}
	}
}

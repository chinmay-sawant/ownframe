package render_test

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

// TestReplayRejectsNonSeparableModes pins the blend modes that need channel
// mixing the per-channel compositor cannot do. The bitmap keeps them.
func TestReplayRejectsNonSeparableModes(t *testing.T) {
	t.Parallel()

	for _, mode := range []string{"hue", "saturation", "color", "luminosity"} {
		display := acceptDisplay(t, `<div style="width:80px;height:80px;`+
			`background:#369;mix-blend-mode:`+mode+`">x</div>`)

		if render.Replayable(display) {
			t.Fatalf("mode %q should not replay", mode)
		}
	}
}

// TestReplayRejectsGroupBandCrossings pins the groups whose members would
// split across the window's paint bands: a z-indexed or fixed member inside a
// blended element cannot buffer into one layer.
func TestReplayRejectsGroupBandCrossings(t *testing.T) {
	t.Parallel()

	zIndexed := acceptDisplay(t, `<div style="width:80px;height:80px;background:#369;`+
		`mix-blend-mode:multiply"><div style="position:relative;z-index:3;`+
		`width:40px;height:40px;background:#f33"></div></div>`)

	if render.Replayable(zIndexed) {
		t.Fatal("a z-indexed member should not replay")
	}

	fixed := acceptDisplay(t, `<div style="width:80px;height:80px;background:#369;`+
		`mix-blend-mode:multiply"><div style="position:fixed;left:4px;top:4px;`+
		`width:40px;height:40px;background:#f33"></div></div>`)

	if render.Replayable(fixed) {
		t.Fatal("a fixed member should not replay")
	}
}

func TestReplayBackgroundBlendModeSeparability(t *testing.T) {
	t.Parallel()

	for _, mode := range []string{"normal", "multiply", "screen", "overlay", "darken", "lighten"} {
		display := acceptDisplay(t, `<div style="width:80px;height:60px;background-color:#f1c40f;`+
			`background-image:linear-gradient(#c0392b,#2980b9);background-blend-mode:`+mode+`"></div>`)
		if !render.Replayable(display) {
			t.Fatalf("background-blend-mode %q should replay", mode)
		}
	}

	for _, mode := range []string{"hue", "saturation", "color", "luminosity"} {
		display := acceptDisplay(t, `<div style="width:80px;height:60px;background-color:#f1c40f;`+
			`background-image:linear-gradient(#c0392b,#2980b9);background-blend-mode:`+mode+`"></div>`)
		if render.Replayable(display) {
			t.Fatalf("non-separable background-blend-mode %q should not replay", mode)
		}
	}
}

func TestReplayAcceptsFilterBlurAndMaskOriginClip(t *testing.T) {
	t.Parallel()

	blurDisplay := acceptDisplay(t, `<div style="width:80px;height:40px;background:#369;filter:blur(2px)">hello</div>`)
	if !render.Replayable(blurDisplay) {
		t.Fatal("filter:blur(2px) should replay")
	}

	maskDisplay := acceptDisplay(t, `<div style="width:80px;height:60px;border:6px solid #000;padding:8px;`+
		`background:#369;mask-image:linear-gradient(#000,transparent);`+
		`mask-origin:padding-box;mask-clip:content-box"></div>`)
	if !render.Replayable(maskDisplay) {
		t.Fatal("mask-origin:padding-box + mask-clip:content-box should replay")
	}
}

package render_test

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

// TestReplayRejectsDeepGroupNesting pins the layer stack bound: five nested
// blend elements exceed maxGroupDepth and stay on the bitmap.
func TestReplayRejectsDeepGroupNesting(t *testing.T) {
	t.Parallel()

	display := acceptDisplay(t, `<div style="width:99px;height:99px;background:#111;mix-blend-mode:multiply">`+
		`<div style="width:88px;height:88px;background:#222;mix-blend-mode:multiply">`+
		`<div style="width:77px;height:77px;background:#333;mix-blend-mode:multiply">`+
		`<div style="width:66px;height:66px;background:#444;mix-blend-mode:multiply">`+
		`<div style="width:55px;height:55px;background:#555;mix-blend-mode:multiply"></div>`+
		`</div></div></div></div>`)

	if render.Replayable(display) {
		t.Fatal("five nested groups should not replay")
	}
}

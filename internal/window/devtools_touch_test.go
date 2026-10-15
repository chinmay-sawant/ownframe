package window

import "testing"

// TestDevPanelStepTapsOnLift fires a panel tap when a still finger lifts,
// and never for content fingers or a zero-size panel.
func TestDevPanelStepTapsOnLift(t *testing.T) {
	t.Parallel()
	panel := devRect{X: 1600, W: 800, H: 1080}
	tr := &panelTrack{}
	all := []touchPos{{id: 1, x: 1700, y: 100}, {id: 2, x: 500, y: 500}}
	content, taps, dy := devPanelStep(all, panel, tr)
	if len(content) != 1 || content[0].id != 2 {
		t.Fatalf("content = %v, want the finger at 500,500", content)
	}
	if len(taps) != 0 || dy != 0 {
		t.Fatalf("landing fired taps = %v dy = %v", taps, dy)
	}
	content, taps, _ = devPanelStep(nil, panel, tr)
	if len(taps) != 1 || taps[0].id != 1 {
		t.Fatalf("lift taps = %v, want the still panel finger", taps)
	}
	if len(content) != 0 {
		t.Fatalf("lift content = %v, want none", content)
	}
	content, taps, _ = devPanelStep(all, devRect{}, &panelTrack{})
	if len(content) != 2 || len(taps) != 0 {
		t.Fatal("a zero panel must not own touches")
	}
}

// TestDevPanelStepDragScrolls scrolls on a dragged panel finger and fires
// no tap for it when it lifts.
func TestDevPanelStepDragScrolls(t *testing.T) {
	t.Parallel()
	panel := devRect{X: 1600, W: 800, H: 1080}
	tr := &panelTrack{}
	devPanelStep([]touchPos{{id: 3, x: 1700, y: 500}}, panel, tr)
	_, _, dy := devPanelStep([]touchPos{{id: 3, x: 1700, y: 460}}, panel, tr)
	if dy != 40 {
		t.Fatalf("drag dy = %v, want 40 up", dy)
	}
	_, taps, _ := devPanelStep(nil, panel, tr)
	if len(taps) != 0 {
		t.Fatalf("dragged lift tapped: %v", taps)
	}
}

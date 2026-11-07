// Probe-patch tests for the scroll trio (TestProbePatchFix2Scroll*).
//
// What patches and what stays FULL (see ui/diff.go: same tag, same child
// count, same order at every node; every changed node needs a stable id;
// a removed attribute forces a rebuild):
//
//   - ScrollArea offset change stays FULL (skipped below): 0 -> 80 mounts the
//     content wrapper and the leading track region (child-count changes), so
//     the diff takes the full-render path. Even with an always-mounted stable
//     structure the move rewrites transform/top/height geometry, which the
//     pinned engine refuses (errSubtreeResized). Needs the blinkless
//     size-change escalation pinned (unpinned work exists but cannot help
//     these tests today).
//   - Scrollbar value change stays FULL (skipped below): 0 -> 200 mounts the
//     leading track region (bar child-count change) and rewrites thumb/region
//     top/height geometry, refused the same way. Same engine dependency.
//   - MessageScroller identical rebuild is a no-op: nothing visible changes.
//   - Transcript append stays FULL by design: one more message changes the
//     transcript's child count, which the diff cannot patch. The scroller
//     keeps stable root/viewport/content part ids from its ID for any future
//     same-shape proof.
//   - Resizable ratio change patches (lock-in): the panes and divider keep
//     stable part ids, so a ratio move is flex-grow and aria-valuenow edits
//     only. No code change was needed.
package chadcn_test

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// TestProbePatchFix2ScrollAreaOffset would prove an offset change on an
// identical child list takes the patch path. Blocked: 0 -> 80 mounts the
// content wrapper and the leading track region and rewrites transform and
// thumb/track geometry, so the update reparses.
func TestProbePatchFix2ScrollAreaOffset(t *testing.T) {
	var state chadcn.ScrollState
	state.SetRange(1000, 200)
	build := func() *ui.Element {
		return ui.Div().ID("probe-root").Children(chadcn.ScrollArea(chadcn.ScrollAreaOptions{
			ID:          "probe-scroll",
			State:       &state,
			TrackPixels: 200,
			Height:      ui.Px(200),
			Children:    []*ui.Element{ui.Div().ID("probe-scroll-row").Text("row")},
		}))
	}
	v, ctx := mountProbeView(t, build())
	if !state.ScrollTo(80) {
		t.Fatal("offset did not move")
	}
	s0 := v.Page.Probe().Stats()
	if err := v.Update(ctx, build()); err != nil {
		t.Fatal(err)
	}
	s1 := v.Page.Probe().Stats()
	t.Skipf("blocked: offset 0->80 mounts the content wrapper and leading track region "+
		"(child-count changes) and rewrites transform/top/height geometry; the diff falls back "+
		"to full render and the pinned engine refuses geometry-changing subtree patches "+
		"(errSubtreeResized) (parses %d->%d); needs the blinkless size-change escalation pinned",
		s0.Parses, s1.Parses)
}

// TestProbePatchFix2ScrollbarValue would prove a standalone scrollbar offset
// change takes the patch path. Blocked: 0 -> 200 mounts the leading track
// region and rewrites thumb/region top/height geometry, so the update reparses.
func TestProbePatchFix2ScrollbarValue(t *testing.T) {
	var state chadcn.ScrollState
	state.SetRange(1000, 200)
	build := func() *ui.Element {
		return ui.Div().ID("probe-root").Children(chadcn.Scrollbar(chadcn.ScrollbarOptions{
			ID:          "probe-bar",
			ControlsID:  "probe-scroll-viewport",
			Label:       "Messages",
			State:       &state,
			TrackPixels: 200,
		}))
	}
	v, ctx := mountProbeView(t, build())
	if !state.ScrollTo(200) {
		t.Fatal("offset did not move")
	}
	s0 := v.Page.Probe().Stats()
	if err := v.Update(ctx, build()); err != nil {
		t.Fatal(err)
	}
	s1 := v.Page.Probe().Stats()
	t.Skipf("blocked: offset 0->200 mounts the leading track region (bar child-count change) "+
		"and rewrites thumb/region top/height geometry; the diff falls back to full render and "+
		"the pinned engine refuses geometry-changing subtree patches (errSubtreeResized) "+
		"(parses %d->%d); needs the blinkless size-change escalation pinned",
		s0.Parses, s1.Parses)
}

// TestProbePatchFix2MessageScrollerNoop proves an identical transcript
// rebuild does no page work.
func TestProbePatchFix2MessageScrollerNoop(t *testing.T) {
	build := func() *ui.Element {
		return ui.Div().ID("probe-root").Children(chadcn.MessageScroller(chadcn.MessageScrollerOptions{
			ID: "probe-ms",
			Children: []*ui.Element{
				ui.Div().ID("probe-ms-m1").Text("hello"),
				ui.Div().ID("probe-ms-m2").Text("world"),
			},
		}))
	}
	v, ctx := mountProbeView(t, build())
	expectProbeNoop(t, v, ctx, build())
}

// TestProbePatchFix2MessageScrollerAppendFull documents that appending a
// transcript child stays on the full-render path by design: the content
// node gains a child, which the diff cannot patch.
func TestProbePatchFix2MessageScrollerAppendFull(t *testing.T) {
	build := func(count int) *ui.Element {
		children := []*ui.Element{ui.Div().ID("probe-ms-m1").Text("hello")}
		if count > 1 {
			children = append(children, ui.Div().ID("probe-ms-m2").Text("world"))
		}
		return ui.Div().ID("probe-root").Children(chadcn.MessageScroller(chadcn.MessageScrollerOptions{
			ID:       "probe-ms",
			Children: children,
		}))
	}
	v, ctx := mountProbeView(t, build(1))
	requireProbeFull(t, v, ctx, build(2))
}

// TestProbePatchFix2ResizableRatio locks in that a ratio move on identical
// panes takes the patch path (flex-grow plus aria-valuenow edits).
func TestProbePatchFix2ResizableRatio(t *testing.T) {
	build := func(ratio float64) *ui.Element {
		return ui.Div().ID("probe-root").Children(chadcn.Resizable(chadcn.ResizableOptions{
			ID:     "probe-split",
			Ratio:  ratio,
			First:  ui.Div().ID("probe-split-a").Text("One"),
			Second: ui.Div().ID("probe-split-b").Text("Two"),
		}))
	}
	v, ctx := mountProbeView(t, build(0.25))
	expectProbePatch(t, v, ctx, build(0.4))
}

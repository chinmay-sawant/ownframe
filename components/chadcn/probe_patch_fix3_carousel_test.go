package chadcn_test

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// TestProbePatchFix3Carousel pins the carousel Index 0->1 move (same 3
// slides) to the full-render path (flip-or-pin verdict: PINNED).
//
// FULL causes, all structural (see ui/diff.go: same tag, same child count,
// same order at every node; a renamed id or a removed attribute rebuilds):
//   - rename: the stage mounts only the visible slide, so the move swaps the
//     slide child id at the same position (pp-car3-s0 -> pp-car3-s1) and the
//     diff refuses the rename. Mounting every slide would avoid the rename
//     but breaks the one-visible-slide contract
//     (TestCarouselShowsOneSlide denies the hidden slides), so it stays.
//   - removal: dot 0 drops aria-current while dot 1 gains it, and the
//     previous control drops disabled as it enables; the patch API sets
//     values but cannot remove an attribute, so each removal rebuilds.
//     A dots-only value swap cannot flip the move while the slide rename and
//     the disabled flip remain, and the ends drop their handlers by design
//     (see TestCarouselDisabledEndsDropHandlers).
//   - pinned-test conflict: TestProbePatchTabsAccStructuralFull/CarouselIndex
//     asserts this exact 0->1 move stays FULL, so flipping stops here.
func TestProbePatchFix3Carousel(t *testing.T) {
	t.Parallel()

	build := func(index int) *ui.Element {
		return ui.Div().ID("probe-root").Children(chadcn.Carousel(chadcn.CarouselOptions{
			ID:    "pp-car3",
			Label: "Shots",
			Index: index,
			Slides: []*ui.Element{
				ui.P().ID("pp-car3-s0").Text("s0"),
				ui.P().ID("pp-car3-s1").Text("s1"),
				ui.P().ID("pp-car3-s2").Text("s2"),
			},
		}))
	}

	t.Run("IndexMoveFull", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, build(0))
		gen0 := v.Page.Probe().Generation()
		s0 := v.Page.Probe().Stats()
		if err := v.Update(ctx, build(1)); err != nil {
			t.Fatal(err)
		}
		s1 := v.Page.Probe().Stats()
		if got, want := v.Page.Probe().Generation(), gen0+1; got != want {
			t.Fatalf("generation = %d, want %d", got, want)
		}
		if got, want := s1.Parses-s0.Parses, uint64(1); got != want {
			t.Fatalf("parses delta = %d, want %d (carousel index stays FULL)", got, want)
		}
	})

	t.Run("IdenticalNoop", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, build(0))
		expectProbeNoop(t, v, ctx, build(0))
	})
}

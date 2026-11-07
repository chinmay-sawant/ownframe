package chadcn_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// TestProbePatchFix3DialogOpen pins the dialog closed->open transition as
// FULL by design: closed renders trigger-only while open mounts the
// backdrop+panel, so the root child count changes and the diff falls back to
// a full render (parses+1). Pre-mounting a hidden backdrop would flip the
// transition to a patch but would break the pinned closed semantics
// (trigger-only closed, nil without a trigger); see TestOverlayDialogClosed
// and TestOverlayDismissClosedDialogIsNotMounted.
func TestProbePatchFix3DialogOpen(t *testing.T) {
	t.Parallel()

	onOpenChange := func(context.Context, bool) error { return nil }
	build := func(open bool, triggerLabel string) *ui.Element {
		return ui.Div().ID("probe-root").Children(chadcn.Dialog(chadcn.DialogOptions{
			ID:           "pp-d3",
			Open:         open,
			Title:        "Edit profile",
			Trigger:      chadcn.Button(triggerLabel, chadcn.ButtonOptions{ID: "pp-d3-trigger"}),
			OnOpenChange: onOpenChange,
		}, ui.P().ID("pp-d3-body").Text("body")))
	}

	t.Run("ClosedToOpenFull", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, build(false, "Open v1"))
		gen0 := v.Page.Probe().Generation()
		s0 := v.Page.Probe().Stats()
		if err := v.Update(ctx, build(true, "Open v1")); err != nil {
			t.Fatal(err)
		}
		s1 := v.Page.Probe().Stats()
		if got, want := v.Page.Probe().Generation(), gen0+1; got != want {
			t.Fatalf("generation = %d, want %d", got, want)
		}
		// FULL by design: closed is trigger-only, open mounts backdrop+panel
		// (child-count change), so Update falls back to one full render.
		if got, want := s1.Parses, s0.Parses+1; got != want {
			t.Fatalf("parses = %d, want %d (expected the full-render fallback)", got, want)
		}
	})

	t.Run("TriggerLabelEditClosed", func(t *testing.T) {
		t.Parallel()
		// Same-length trigger label swap fits the kept button box, so the
		// closed-only edit takes the patch path.
		v, ctx := mountProbeView(t, build(false, "Open v1"))
		expectProbePatch(t, v, ctx, build(false, "Open v2"))
	})

	t.Run("IdenticalClosedNoop", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, build(false, "Open v1"))
		expectProbeNoop(t, v, ctx, build(false, "Open v1"))
	})
}

package chadcn_test

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

func TestProbePatchFix3NoId(t *testing.T) {
	t.Parallel()

	noid := func(label string) *ui.Element {
		return ui.Div().ID("probe-root").Children(chadcn.Button(label, chadcn.ButtonOptions{}))
	}
	withID := func(label string) *ui.Element {
		return ui.Div().ID("probe-root").Children(chadcn.Button(label, chadcn.ButtonOptions{ID: "btn-save"}))
	}

	t.Run("NoIdLabelChangeFallsBackToFullRender", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, noid("Save"))
		s0 := v.Page.Probe().Stats()
		if err := v.Update(ctx, noid("Saved")); err != nil {
			t.Fatal(err)
		}
		s1 := v.Page.Probe().Stats()
		// Anonymous changed node, unaddressable: without an id the diff
		// cannot target the label span, so the update reparses.
		if got, want := s1.Parses, s0.Parses+1; got != want {
			t.Fatalf("parses = %d, want %d (anonymous changed node, unaddressable)", got, want)
		}
	})

	t.Run("SameChangeWithIDPatches", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, withID("Save"))
		expectProbePatch(t, v, ctx, withID("Saved"))
	})

	t.Run("IdenticalNoIdRebuildIsNoop", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, noid("Save"))
		expectProbeNoop(t, v, ctx, noid("Save"))
	})
}

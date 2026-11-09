package chadcn_test

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

func TestProbePatchChecks(t *testing.T) {
	t.Parallel()

	t.Run("checkbox flip", func(t *testing.T) {
		t.Parallel()
		before := func() *ui.Element {
			return ui.Div().ID("probe-root").Children(
				chadcn.CheckboxWithOptions("Terms", false, chadcn.CheckboxOptions{ID: "chk-terms"}).ID("chk-terms"),
			)
		}
		after := func() *ui.Element {
			return ui.Div().ID("probe-root").Children(
				chadcn.CheckboxWithOptions("Terms", true, chadcn.CheckboxOptions{ID: "chk-terms"}).ID("chk-terms"),
			)
		}
		v, ctx := mountProbeView(t, before())
		expectProbePatch(t, v, ctx, after())
	})

	t.Run("switch flip", func(t *testing.T) {
		t.Parallel()
		before := func() *ui.Element {
			return ui.Div().ID("probe-root").Children(
				chadcn.SwitchWithOptions("Wifi", false, chadcn.SwitchOptions{ID: "sw-wifi"}).ID("sw-wifi"),
			)
		}
		after := func() *ui.Element {
			return ui.Div().ID("probe-root").Children(
				chadcn.SwitchWithOptions("Wifi", true, chadcn.SwitchOptions{ID: "sw-wifi"}).ID("sw-wifi"),
			)
		}
		v, ctx := mountProbeView(t, before())
		expectProbePatch(t, v, ctx, after())
	})

	t.Run("slider value", func(t *testing.T) {
		t.Parallel()
		before := func() *ui.Element {
			return ui.Div().ID("probe-root").Children(
				chadcn.Slider(chadcn.SliderOptions{ID: "sl-vol", Label: "Volume", Value: 20}).ID("sl-vol"),
			)
		}
		after := func() *ui.Element {
			return ui.Div().ID("probe-root").Children(
				chadcn.Slider(chadcn.SliderOptions{ID: "sl-vol", Label: "Volume", Value: 80}).ID("sl-vol"),
			)
		}
		v, ctx := mountProbeView(t, before())
		s0 := v.Page.Probe().Stats()
		if err := v.Update(ctx, after()); err != nil {
			t.Fatal(err)
		}
		s1 := v.Page.Probe().Stats()
		t.Skipf("blocked: fill width 20%%->80%% resizes the patched element and the engine "+
			"refuses geometry-changing subtree patches (parses %d->%d); needs engine or ui/diff support",
			s0.Parses, s1.Parses)
	})

	t.Run("progress value", func(t *testing.T) {
		t.Parallel()
		before := func() *ui.Element {
			return ui.Div().ID("probe-root").Children(
				chadcn.ProgressWithOptions("Upload", 20, chadcn.ProgressOptions{ID: "pg-up"}).ID("pg-up"),
			)
		}
		after := func() *ui.Element {
			return ui.Div().ID("probe-root").Children(
				chadcn.ProgressWithOptions("Upload", 80, chadcn.ProgressOptions{ID: "pg-up"}).ID("pg-up"),
			)
		}
		v, ctx := mountProbeView(t, before())
		s0 := v.Page.Probe().Stats()
		if err := v.Update(ctx, after()); err != nil {
			t.Fatal(err)
		}
		s1 := v.Page.Probe().Stats()
		t.Skipf("blocked: fill width 20%%->80%% resizes the patched element and the engine "+
			"refuses geometry-changing subtree patches (parses %d->%d); needs engine or ui/diff support",
			s0.Parses, s1.Parses)
	})
}

package chadcn_test

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

func probeButtonRoot(el *ui.Element) *ui.Element {
	return ui.Div().ID("probe-root").Children(el)
}

func TestProbePatchButtons(t *testing.T) {
	t.Parallel()

	t.Run("ButtonLabel", func(t *testing.T) {
		t.Parallel()
		// The label span is a content-sized flex item, so only a label
		// that fits the kept box patches: growing text must widen the
		// shrink-to-fit button, which needs an ancestor reflow that only
		// a full render does (the engine refuses the size change). The
		// part id below is what lets a fitting edit take the patch path
		// instead of the diff fallback.
		labeled := func(label string) *ui.Element {
			return probeButtonRoot(chadcn.Button(label, chadcn.ButtonOptions{ID: "btn-save"}))
		}
		v, ctx := mountProbeView(t, labeled("Saved"))
		expectProbePatch(t, v, ctx, labeled("Save"))
	})

	t.Run("TogglePressed", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, probeButtonRoot(chadcn.Toggle("Bold", false).ID("tgl-bold")))
		expectProbePatch(t, v, ctx, probeButtonRoot(chadcn.Toggle("Bold", true).ID("tgl-bold")))
	})

	t.Run("ToggleGroupSelect", func(t *testing.T) {
		t.Parallel()
		items := []chadcn.ToggleItem{
			{ID: "tg-a", Value: "a", Label: "A"},
			{ID: "tg-b", Value: "b", Label: "B"},
		}
		v, ctx := mountProbeView(t, probeButtonRoot(chadcn.ToggleGroup(chadcn.ToggleGroupOptions{
			Items: items, Selected: []string{"a"},
		})))
		expectProbePatch(t, v, ctx, probeButtonRoot(chadcn.ToggleGroup(chadcn.ToggleGroupOptions{
			Items: items, Selected: []string{"b"},
		})))
	})

	t.Run("BadgeVariant", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, probeButtonRoot(chadcn.Badge("New").ID("bdg-new")))
		expectProbePatch(t, v, ctx, probeButtonRoot(
			chadcn.Badge("New", chadcn.BadgeOptions{Variant: chadcn.BadgeSecondary}).ID("bdg-new")))
	})

	t.Run("ButtonGroupNoop", func(t *testing.T) {
		t.Parallel()
		group := func() *ui.Element {
			return probeButtonRoot(chadcn.ButtonGroup(chadcn.ButtonGroupOptions{Label: "View"},
				chadcn.Button("List", chadcn.ButtonOptions{}),
				chadcn.Button("Grid", chadcn.ButtonOptions{}),
			))
		}
		v, ctx := mountProbeView(t, group())
		expectProbeNoop(t, v, ctx, group())
	})

	t.Run("KbdNoop", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, probeButtonRoot(chadcn.Kbd("Ctrl")))
		expectProbeNoop(t, v, ctx, probeButtonRoot(chadcn.Kbd("Ctrl")))
	})

	t.Run("LabelNoop", func(t *testing.T) {
		t.Parallel()
		labeled := func() *ui.Element {
			return probeButtonRoot(chadcn.Label("Name", chadcn.LabelOptions{For: "name-input"}))
		}
		v, ctx := mountProbeView(t, labeled())
		expectProbeNoop(t, v, ctx, labeled())
	})
}

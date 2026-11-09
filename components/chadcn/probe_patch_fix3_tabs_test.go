package chadcn_test

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// TestProbePatchFix3TabsSwitchFull pins the tabs Active a->b switch to the
// full-render path (flip-or-pin verdict: PINNED).
//
// FULL causes, all by design (see ui/diff.go: same tag, same child count,
// same order at every node; a renamed id or a removed attribute rebuilds):
//   - mount: Tabs renders only the active panel lazily (Render builds just
//     the active item; feedback_test.go forbids inactive panel content), so
//     the switch swaps the panel subtree for host-driven content.
//   - rename: the single mounted panel id derives from the active item
//     (tab-a-panel -> tab-b-panel), and the diff compares children
//     positionally, so the rename alone forces a rebuild. Stable positional
//     panel ids would break tabs_fix_test.go (`id="analytics-panel"`) and
//     TestTabsRendersActiveTabAndPanel.
//   - removal: only the selected trigger carries aria-controls, so the old
//     trigger drops it (attribute removal rebuilds). Keeping it would dangle
//     at an unmounted panel, and always-mounted panels would break lazy
//     mounting. Every flip path needs one of those, so FULL stays.
func TestProbePatchFix3TabsSwitchFull(t *testing.T) {
	t.Parallel()

	build := func(active string) *ui.Element {
		return ui.Div().ID("probe-root").Children(chadcn.Tabs(chadcn.TabsOptions{
			Active: active,
			Items: []chadcn.TabItem{
				{ID: "tab-a", Label: "Alpha"},
				{ID: "tab-b", Label: "Beta"},
			},
			Render: func(id string) *ui.Element {
				return ui.P().ID("pp-fix3-tabs-" + id).Text("panel " + id)
			},
		}))
	}

	v, ctx := mountProbeView(t, build("tab-a"))
	gen0 := v.Page.Probe().Generation()
	s0 := v.Page.Probe().Stats()
	if err := v.Update(ctx, build("tab-b")); err != nil {
		t.Fatal(err)
	}
	s1 := v.Page.Probe().Stats()
	if got, want := v.Page.Probe().Generation(), gen0+1; got != want {
		t.Fatalf("generation = %d, want %d", got, want)
	}
	if got, want := s1.Parses-s0.Parses, uint64(1); got != want {
		t.Fatalf("parses delta = %d, want %d (tabs switch stays FULL)", got, want)
	}
}

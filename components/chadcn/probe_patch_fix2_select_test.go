// Probe-patch tests for the Select open value change
// (TestProbePatchFix2Select*).
//
// What patches and what stays FULL by design (see ui/diff.go: same tag,
// same child count, same order at every node; every changed node needs a
// stable id; a removed attribute forces a rebuild):
//
//   - Open value a->b on an identical item list patches: each option's
//     selection marker now stays mounted (empty when unselected) with a
//     stable ID+"-option-<n>-check" id, so the move is aria-selected edits
//     on the id-bearing options plus text edits on the two markers.
//     Previously the check span mounted/unmounted and changed the row's
//     child count, which forced a full render.
//   - The trigger carries no conditional marker (the label and chevron
//     always render), so its label text edit already patched; no change
//     was needed there.
//   - The marker spans are flex children like the combobox checks, and the
//     same-list move patches the same way the combobox one does
//     (TestProbePatchSelectsComboboxSameList), so no resizing case is
//     engine-blocked here.
//   - Closed->open stays FULL: the panel child mounts and the root child
//     count changes, which no id scheme can paper over.
package chadcn_test

import (
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

func fix2SelectItems() []chadcn.SelectItem {
	return []chadcn.SelectItem{
		{Value: "a", Label: "Alpha"},
		{Value: "b", Label: "Beta"},
	}
}

// TestProbePatchFix2SelectOpenValueChange proves an open Select value change
// on an identical item list takes the patch path (parses+0) instead of the
// full-render fallback.
func TestProbePatchFix2SelectOpenValueChange(t *testing.T) {
	t.Parallel()
	items := fix2SelectItems()
	v, ctx := mountProbeView(t, ui.Div().ID("probe-root").Children(
		chadcn.Select(chadcn.SelectOptions{
			ID: "sel-2", Open: true, Value: "a", Items: items,
		}),
	))
	expectProbePatch(t, v, ctx, ui.Div().ID("probe-root").Children(
		chadcn.Select(chadcn.SelectOptions{
			ID: "sel-2", Open: true, Value: "b", Items: items,
		}),
	))
}

// TestProbePatchFix2SelectIdenticalOpenNoop proves an identical open rebuild
// does no page work.
func TestProbePatchFix2SelectIdenticalOpenNoop(t *testing.T) {
	t.Parallel()
	items := fix2SelectItems()
	v, ctx := mountProbeView(t, ui.Div().ID("probe-root").Children(
		chadcn.Select(chadcn.SelectOptions{
			ID: "sel-2", Open: true, Value: "a", Items: items,
		}),
	))
	expectProbeNoop(t, v, ctx, ui.Div().ID("probe-root").Children(
		chadcn.Select(chadcn.SelectOptions{
			ID: "sel-2", Open: true, Value: "a", Items: items,
		}),
	))
}

// TestProbePatchFix2SelectMarkersStayMounted proves every open option keeps
// its selection marker with a stable id, so unselected options render an
// empty marker instead of omitting the node.
func TestProbePatchFix2SelectMarkersStayMounted(t *testing.T) {
	t.Parallel()
	rendered, err := ui.Render(chadcn.Select(chadcn.SelectOptions{
		ID: "sel-2", Open: true, Value: "a", Items: fix2SelectItems(),
	}))
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	for _, fragment := range []string{
		`id="sel-2-option-0-check"`,
		`id="sel-2-option-1-check"`,
	} {
		if !strings.Contains(rendered, fragment) {
			t.Fatalf("rendered markup is missing %q:\n%s", fragment, rendered)
		}
	}
	if got := strings.Count(rendered, `class="of-select-check"`); got != 2 {
		t.Fatalf("check markers=%d want 2:\n%s", got, rendered)
	}
}

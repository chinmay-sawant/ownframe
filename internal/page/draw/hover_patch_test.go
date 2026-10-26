package draw

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

func hoverTestBoxes() []layout.Box {
	// page
	// ├─ head
	// └─ body
	//    ├─ list
	//    │  ├─ row-a
	//    │  │  ├─ price-a
	//    │  │  └─ change-a
	//    │  └─ row-b
	//    │     └─ price-b
	//    └─ side
	return []layout.Box{
		{ID: "page", ParentIndex: -1},
		{ID: "head", ParentIndex: 0},
		{ID: "body", ParentIndex: 0},
		{ID: "list", ParentIndex: 2},
		{ID: "row-a", ParentIndex: 3},
		{ID: "price-a", ParentIndex: 4},
		{ID: "change-a", ParentIndex: 4},
		{ID: "row-b", ParentIndex: 3},
		{ID: "price-b", ParentIndex: 7},
		{ID: "side", ParentIndex: 2},
	}
}

func TestHoverPatchRootsSiblings(t *testing.T) {
	t.Parallel()

	// Moving between rows patches both rows, not the list.
	roots, ok := HoverPatchRoots(hoverTestBoxes(), "price-a", "price-b")
	if !ok {
		t.Fatal("sibling rows did not scope")
	}

	if len(roots) != 2 || roots[0] != "row-a" || roots[1] != "row-b" {
		t.Fatalf("roots = %v, want [row-a row-b]", roots)
	}
}

func TestHoverPatchRootsWithinRow(t *testing.T) {
	t.Parallel()

	// Moving inside one row patches the two children: the row itself matched
	// :hover in both frames, so its own rules did not change.
	roots, ok := HoverPatchRoots(hoverTestBoxes(), "price-a", "change-a")
	if !ok {
		t.Fatal("same-row move did not scope")
	}

	if len(roots) != 2 || roots[0] != "price-a" || roots[1] != "change-a" {
		t.Fatalf("roots = %v, want [price-a change-a]", roots)
	}
}

func TestHoverPatchRootsAncestorChild(t *testing.T) {
	t.Parallel()

	// Moving from a row to the list that contains it only patches the row:
	// the list matched :hover in both frames.
	roots, ok := HoverPatchRoots(hoverTestBoxes(), "price-a", "list")
	if !ok {
		t.Fatal("row to ancestor did not scope")
	}

	if len(roots) != 1 || roots[0] != "row-a" {
		t.Fatalf("roots = %v, want [row-a]", roots)
	}
}

func TestHoverPatchRootsAcrossPanels(t *testing.T) {
	t.Parallel()

	// Moving from the list to the sidebar reaches the body: patching both
	// would cover the whole page, so the caller falls back.
	if _, ok := HoverPatchRoots(hoverTestBoxes(), "price-a", "side"); ok {
		t.Fatal("cross-panel move scoped instead of falling back")
	}
}

func TestHoverPatchRootsRejects(t *testing.T) {
	t.Parallel()

	boxes := hoverTestBoxes()

	for _, tc := range []struct {
		name string
		old  string
		new  string
	}{
		{"empty old", "", "price-a"},
		{"empty new", "price-a", ""},
		{"same", "price-a", "price-a"},
		{"missing old", "gone", "price-a"},
		{"missing new", "price-a", "gone"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, ok := HoverPatchRoots(boxes, tc.old, tc.new); ok {
				t.Fatalf("roots for %q to %q scoped", tc.old, tc.new)
			}
		})
	}
}

func TestHoverPatchRisky(t *testing.T) {
	t.Parallel()

	risky := []string{
		`.a:has(.b:hover){color:red}`,
		`body:has(> :hover){background:#fff}`,
		`.a:HOVER + .b{color:red}`,
		".a:hover\t~\n.b{color:red}",
		`.a:hover::before + .b{content:""}`,
		// Descendant-then-sibling shapes stay inside the hovered subtree,
		// but the textual scan cannot tell them apart, so they fall back
		// to the full layout and stay correct.
		`.a:hover > .b + .c{color:red}`,
	}

	for _, source := range risky {
		if !HoverPatchRisky(source) {
			t.Errorf("risky source passed: %q", source)
		}
	}

	safe := []string{
		`.a:hover{color:red}`,
		`.a:hover .b{color:red}`,
		`.a:hover, .b + .c{color:red}`,
		`/* .a:hover + .b */ .a:hover{color:red}`,
		`.a{content:":hover + .b"}`,
		`.a{content:':has('}`,
		``,
	}

	for _, source := range safe {
		if HoverPatchRisky(source) {
			t.Errorf("safe source flagged: %q", source)
		}
	}

	if !HoverPatchRisky(`.a:hover{color:red}`, `.x:has(.y){color:blue}`) {
		t.Error("risky theme did not flag")
	}
}

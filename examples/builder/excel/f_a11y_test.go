package main

import (
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// a11yNodes returns the live accessibility snapshot, failing the test on error.
func a11yNodes(t *testing.T, a *app) []ownframe.AccessibilityNode {
	t.Helper()
	tree, err := a.page.Access().AccessibilityTree()
	if err != nil {
		t.Fatal(err)
	}
	return tree.Nodes
}

func a11yFind(nodes []ownframe.AccessibilityNode, role ownframe.AccessibilityRole, name string) (ownframe.AccessibilityNode, bool) {
	for _, n := range nodes {
		if n.Name == name && (role == "" || n.Role == role) {
			return n, true
		}
	}
	return ownframe.AccessibilityNode{}, false
}

// TestAccessibilityAudit pins EXL-78: the Name Box, formula bar, sheet tabs,
// ribbon buttons and visible grid cells all appear in page.Access() with a
// role and a human name, and nothing interactive is anonymous.
func TestAccessibilityAudit(t *testing.T) {
	a, _ := newTestApp(t)
	nodes := a11yNodes(t, a)

	nb, ok := a11yFind(nodes, ownframe.AccessibilityTextbox, "Name Box")
	if !ok || nb.Value != "A1" {
		t.Fatalf("Name Box = %+v ok=%v, want textbox valued A1", nb, ok)
	}
	if fx, ok := a11yFind(nodes, ownframe.AccessibilityTextbox, "Formula bar"); !ok || fx.Value != "Region" {
		t.Fatalf("Formula bar = %+v ok=%v, want textbox valued Region", fx, ok)
	}
	for _, tab := range []string{"File", "Home", "Insert", "Page Layout", "Formulas", "Data", "Review", "View"} {
		if _, ok := a11yFind(nodes, ownframe.AccessibilityButton, tab); !ok {
			t.Errorf("ribbon tab %q is not a named button", tab)
		}
	}
	for _, btn := range []string{"Bold (Ctrl+B)", "Italic (Ctrl+I)", "Paste", "AutoSum", "Number format", "Font size"} {
		if _, ok := a11yFind(nodes, ownframe.AccessibilityButton, btn); !ok {
			t.Errorf("Home ribbon button %q missing", btn)
		}
	}
	for _, g := range []string{"Clipboard", "Font", "Alignment", "Number"} {
		if _, ok := a11yFind(nodes, ownframe.AccessibilityGroup, g); !ok {
			t.Errorf("ribbon group %q missing", g)
		}
	}
	for _, tab := range a.wb.Sheets {
		if _, ok := a11yFind(nodes, ownframe.AccessibilityButton, tab.Name); !ok {
			t.Errorf("sheet tab %q is not in the tree", tab.Name)
		}
	}
	if _, ok := a11yFind(nodes, "", "Add sheet"); !ok {
		t.Errorf("Add sheet button missing")
	}
	// Cells expose their displayed text, so "$12,500" is readable.
	for _, want := range []string{"Region", "North", "$12,500", "26.4%"} {
		if _, ok := a11yFind(nodes, "", want); !ok {
			t.Errorf("grid cell text %q missing from the tree", want)
		}
	}
	for _, n := range nodes {
		if n.Role == ownframe.AccessibilityButton && strings.TrimSpace(n.Name) == "" {
			t.Errorf("anonymous button %s", n.ElementID)
		}
	}
}

// TestAccessibilityTracksSelection checks that the Name Box and formula bar
// values follow the active cell, and the open dialog is announced.
func TestAccessibilityTracksSelection(t *testing.T) {
	a, _ := newTestApp(t)
	a.selectCell(CellRef{2, 4})
	a.redraw()
	nodes := a11yNodes(t, a)
	if nb, _ := a11yFind(nodes, ownframe.AccessibilityTextbox, "Name Box"); nb.Value != "E3" {
		t.Errorf("Name Box value = %q, want E3", nb.Value)
	}
	if fx, _ := a11yFind(nodes, ownframe.AccessibilityTextbox, "Formula bar"); fx.Value != a.sheet().Raw(2, 4) {
		t.Errorf("Formula bar value = %q, want %q", fx.Value, a.sheet().Raw(2, 4))
	}
	a.runAction("xl:dlg:shortcuts")
	nodes = a11yNodes(t, a)
	if _, ok := a11yFind(nodes, "", "Keyboard Shortcuts"); !ok {
		t.Errorf("open dialog title not announced")
	}
}

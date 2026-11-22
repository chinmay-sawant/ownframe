package uia

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

func hitTree() host.AccessibilityTree {
	b := func(x, y, w, h float64) host.AccessibilityBounds {
		return host.AccessibilityBounds{X: x, Y: y, W: w, H: h}
	}

	return host.AccessibilityTree{
		Document: 1, Revision: 1, Root: "root",
		Nodes: []host.AccessibilityNode{
			{ID: "root", Role: host.AccessibilityDocument, Name: "Page", Children: []host.AccessibilityNodeID{"card", "lone"}},
			{ID: "card", Parent: "root", Role: host.AccessibilityGroup, Name: "Card", HasBounds: true, Bounds: b(10, 10, 300, 200),
				Children: []host.AccessibilityNodeID{"name", "ok", "ok2"}},
			{ID: "name", Parent: "card", Role: host.AccessibilityTextbox, Name: "Name", HasBounds: true, Bounds: b(20, 30, 200, 30)},
			{ID: "ok", Parent: "card", Role: host.AccessibilityButton, Name: "OK", HasBounds: true, Bounds: b(20, 100, 80, 30)},
			{ID: "ok2", Parent: "card", Role: host.AccessibilityButton, Name: "OK2", HasBounds: true, Bounds: b(60, 100, 80, 30)},
			{ID: "lone", Parent: "root", Role: host.AccessibilityText, Name: "No geometry"},
		},
	}
}

func hitProvider(g Geometry) *Provider {
	p := NewProvider(nil)
	p.Update(hitTree())
	p.SetGeometry(g)

	return p
}

func TestElementAtPicksTheDeepestElement(t *testing.T) {
	p := hitProvider(Geometry{OriginX: 0, OriginY: 0, Scale: 1, Width: 800, Height: 600})

	cases := []struct {
		name string
		x, y float64
		want host.AccessibilityNodeID
		ok   bool
	}{
		{"inside the textbox, not the card", 50, 45, "name", true},
		{"inside the card only", 250, 150, "card", true},
		{"inside the button", 30, 110, "ok", true},
		{"outside every element falls back to the window", 600, 500, "", false},
		{"just past the right edge of the card", 310, 50, "", false},
		{"the left edge counts", 10, 10, "card", true},
		{"the right edge does not", 310, 10, "", false},
	}

	for _, tc := range cases {
		got, ok := p.ElementAt(tc.x, tc.y)
		if got != tc.want || ok != tc.ok {
			t.Errorf("%s: ElementAt(%v, %v) = %q, %t; want %q, %t", tc.name, tc.x, tc.y, got, ok, tc.want, tc.ok)
		}
	}
}

func TestElementAtLaterSiblingWinsWhereTheyOverlap(t *testing.T) {
	p := hitProvider(Geometry{Scale: 1, Width: 800, Height: 600})

	// OK spans x 20..100 and OK2 spans x 60..140; both are on y 100..130.
	if got, _ := p.ElementAt(80, 110); got != "ok2" {
		t.Fatalf("ElementAt over two overlapping siblings = %q, want the later one, ok2", got)
	}

	if got, _ := p.ElementAt(30, 110); got != "ok" {
		t.Fatalf("ElementAt where only OK is = %q, want ok", got)
	}
}

func TestElementAtAppliesOriginAndScale(t *testing.T) {
	// The window sits at (-1920, 320) on a 150 percent monitor.
	p := hitProvider(Geometry{OriginX: -1920, OriginY: 320, Scale: 1.5, Width: 1200, Height: 800})

	// The OK button is at page (20,100) size 80 by 30: on screen that is
	// x -1920+30 .. -1920+150 and y 320+150 .. 320+195.
	if got, ok := p.ElementAt(-1920+30+5, 320+150+5); !ok || got != "ok" {
		t.Fatalf("ElementAt on the scaled button = %q, %t; want ok", got, ok)
	}

	// The same pixel, read as if there were no scale, would miss it.
	if got, ok := p.ElementAt(-1920+25, 320+105); ok && got == "ok" {
		t.Fatalf("ElementAt ignored the scale and found %q", got)
	}
}

func TestElementAtWithoutATree(t *testing.T) {
	p := NewProvider(nil)

	if got, ok := p.ElementAt(1, 1); ok || got != "" {
		t.Fatalf("ElementAt on a provider with no tree = %q, %t", got, ok)
	}
}

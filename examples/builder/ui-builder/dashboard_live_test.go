package main

import (
	"context"
	"strings"
	"testing"
)

func TestBuilderViewportFitsWindow(t *testing.T) {
	a, err := newBuilderApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	tabs := []struct {
		name   string
		marker string
	}{
		{"widgets", "Component library"},
		{"settings", "Settings"},
		{"layout", "Layout playground"},
		{"todo", "Tasks ("},
		{"list", "File list"},
		{"theme", "Theme"},
		{"compare", "Builder vs HTML"},
		{"dialog", "Backdrop dialog"},
	}
	for _, tab := range tabs {
		a.tab = tab.name
		if err := a.refresh(ctx); err != nil {
			t.Fatal(err)
		}
		if err := a.page.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
		if html := a.page.HTML(); !strings.Contains(html, tab.marker) {
			t.Errorf("tab %q HTML missing marker %q", tab.name, tab.marker)
		}
		for _, b := range a.page.Boxes() {
			if b.X+b.W > 1121 {
				t.Fatalf("tab %q box overflows window width: %+v", tab.name, b)
			}
		}
	}
}

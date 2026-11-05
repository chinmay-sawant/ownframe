package main

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestListShowsRows(t *testing.T) {
	s := &listApp{selected: -1}
	tt := ui.NewTester(s.view, listWidth, listHeight)
	for _, want := range []string{"Compare Table · 1000 rows", listSubtitle, "Item 1", "Item 1000", "Num", "Description"} {
		if !tt.HasText(want) {
			t.Errorf("missing %q, got %q", want, tt.Texts())
		}
	}
}

func TestListSelectRow(t *testing.T) {
	s := &listApp{selected: -1}
	tt := ui.NewTester(s.view, listWidth, listHeight)
	if err := tt.Click("Item 5"); err != nil {
		t.Fatal(err)
	}
	if s.selected != 4 {
		t.Fatalf("selected = %d, want 4", s.selected)
	}
	if !tt.HasText("Selected: Item 5") {
		t.Errorf("missing selection readout, got %q", tt.Texts())
	}
}

// BenchmarkListClick measures the full data-update path: each click on a
// visible row selects it and rebuilds the whole table view with new data.
func BenchmarkListClick(b *testing.B) {
	s := &listApp{selected: -1}
	tt := ui.NewTester(s.view, listWidth, listHeight)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := tt.Click("Item 5"); err != nil {
			b.Fatal(err)
		}
	}
	if s.selected != 4 {
		b.Fatalf("selected = %d, want 4", s.selected)
	}
}

// BenchmarkListTester measures building the whole thousand-row view in
// memory without a window: one tester per iteration, then a text lookup.
func BenchmarkListTester(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		s := &listApp{selected: 500}
		tt := ui.NewTester(s.view, listWidth, listHeight)
		if !tt.HasText("Item 1000") {
			b.Fatal("missing row text")
		}
	}
}

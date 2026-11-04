package main

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestListShowsWindow(t *testing.T) {
	s := &listApp{selected: -1}
	tt := ui.NewTester(s.view, listWidth, listHeight)
	for _, want := range []string{"Compare Table · 1000 rows", listSubtitle, "Item 1", "Num", "Description"} {
		if !tt.HasText(want) {
			t.Errorf("missing %q, got %q", want, tt.Texts())
		}
	}
	// The virtualized list builds only rows in view.
	if tt.HasText("Item 1000") {
		t.Errorf("list should not build Item 1000 outside the view")
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
// visible row selects it and reframes the virtualized list.
func BenchmarkListClick(b *testing.B) {
	s := &listApp{selected: -1}
	tt := ui.NewTester(s.view, listWidth, listHeight)
	labels := []string{"Item 5", "Item 6"}
	want := []int{4, 5}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := tt.Click(labels[i%len(labels)]); err != nil {
			b.Fatal(err)
		}
	}
	if s.selected != want[(b.N-1)%len(want)] {
		b.Fatalf("selected = %d", s.selected)
	}
}

// BenchmarkListTester measures building the virtualized table view in memory
// without a window: one tester per iteration, then a text lookup.
func BenchmarkListTester(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		s := &listApp{selected: 4}
		tt := ui.NewTester(s.view, listWidth, listHeight)
		if !tt.HasText("Item 1") {
			b.Fatal("missing row text")
		}
	}
}

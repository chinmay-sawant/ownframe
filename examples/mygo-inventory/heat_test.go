package main

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestDashboardShowsPanels(t *testing.T) {
	s := newHeatApp()
	tt := ui.NewTester(s.view, winWidth, winHeight)
	for _, want := range []string{
		"Northwind Goods", "Inventory health", "Stock health", "Bin utilization",
		"Stock heat by product", "Inbound purchase orders", "Inventory value by category",
		"Inventory value", "Stockouts", "Hottest aisle E",
	} {
		if !tt.HasText(want) {
			t.Errorf("missing %q, got %q", want, tt.Texts())
		}
	}
}

func TestStockFilterPills(t *testing.T) {
	s := newHeatApp()
	tt := ui.NewTester(s.view, winWidth, winHeight)
	if err := tt.Click("Stockout 7"); err != nil {
		t.Fatal(err)
	}
	if s.statusFilter != "Stockout" {
		t.Fatalf("statusFilter = %q, want Stockout", s.statusFilter)
	}
	if !tt.HasText("Ceremonial matcha") {
		t.Errorf("filtered view missing Stockout row, got %q", tt.Texts())
	}
	if tt.HasText("Nitrile gloves") {
		t.Errorf("filtered view still shows Healthy row, got %q", tt.Texts())
	}
	if err := tt.Click("All SKUs 4,200"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText("Nitrile gloves") {
		t.Errorf("cleared filter missing Healthy row, got %q", tt.Texts())
	}
}

func TestCategorySelectFilters(t *testing.T) {
	s := newHeatApp()
	s.category = "Janitorial"
	tt := ui.NewTester(s.view, winWidth, winHeight)
	if !tt.HasText("Nitrile gloves") {
		t.Errorf("category view missing Janitorial row, got %q", tt.Texts())
	}
	if tt.HasText("Ceremonial matcha") {
		t.Errorf("category view still shows Coffee row, got %q", tt.Texts())
	}
}

func TestSearchFilters(t *testing.T) {
	s := newHeatApp()
	s.query = "matcha"
	tt := ui.NewTester(s.view, winWidth, winHeight)
	if !tt.HasText("Ceremonial matcha") {
		t.Errorf("search missing matcha row, got %q", tt.Texts())
	}
	if tt.HasText("Nitrile gloves · Healthy") {
		t.Errorf("search still shows gloves result, got %q", tt.Texts())
	}
}

func TestPODialog(t *testing.T) {
	s := newHeatApp()
	tt := ui.NewTester(s.view, winWidth, winHeight)
	if err := tt.Click("+ New PO"); err != nil {
		t.Fatal(err)
	}
	if !s.poOpen {
		t.Fatal("poOpen = false after + New PO")
	}
	if !tt.HasText("New purchase order") {
		t.Errorf("dialog missing title, got %q", tt.Texts())
	}
	if err := tt.Click("Cancel"); err != nil {
		t.Fatal(err)
	}
	if s.poOpen {
		t.Fatal("poOpen = true after Cancel")
	}
}

func TestHeatShowsProducts(t *testing.T) {
	s := newHeatApp()
	tt := ui.NewTester(s.view, winWidth, winHeight)
	for _, want := range []string{
		"Stock heat by product", "Cold brew", "Vanilla", "Overfull", "Room", "Selected: none",
	} {
		if !tt.HasText(want) {
			t.Errorf("missing %q, got %q", want, tt.Texts())
		}
	}
}

func TestHeatSelectCell(t *testing.T) {
	s := newHeatApp()
	tt := ui.NewTester(s.view, winWidth, testHeight)
	if err := tt.Click(cellLabel(7, 4)); err != nil {
		t.Fatal(err)
	}
	if s.selP != 7 || s.selA != 4 {
		t.Fatalf("selected = %d,%d, want 7,4", s.selP, s.selA)
	}
	if !tt.HasText("Vanilla syrup") {
		t.Errorf("missing selection readout, got %q", tt.Texts())
	}
}

func TestCellValueUsesSeparateProducts(t *testing.T) {
	// Same aisle, different products: Overstock reads hotter than Stockout.
	lo := cellValue(0, 4)
	hi := cellValue(7, 4)
	if hi <= lo {
		t.Fatalf("overstock cell = %d, want hotter than stockout cell = %d", hi, lo)
	}
	for pi := range products {
		for ai := 0; ai < 10; ai++ {
			if v := cellValue(pi, ai); v < 0 || v > 100 {
				t.Fatalf("cellValue(%d,%d) = %d, want 0..100", pi, ai, v)
			}
		}
	}
}

// BenchmarkHeatClick measures the full data-update path: alternating cell
// clicks select and rebuild the dashboard with new state.
func BenchmarkHeatClick(b *testing.B) {
	s := newHeatApp()
	tt := ui.NewTester(s.view, winWidth, testHeight)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			if err := tt.Click(cellLabel(7, 4)); err != nil {
				b.Fatal(err)
			}
		} else if err := tt.Click(cellLabel(0, 0)); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkHeatTester builds the whole dashboard view in memory without a
// window: one tester per iteration, then a text lookup.
func BenchmarkHeatTester(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		s := newHeatApp()
		s.selP, s.selA = 7, 4
		tt := ui.NewTester(s.view, winWidth, winHeight)
		if !tt.HasText("Inventory health") {
			b.Fatal("missing dashboard text")
		}
	}
}

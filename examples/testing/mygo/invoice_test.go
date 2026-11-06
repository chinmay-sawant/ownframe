package main

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestInvoiceShowsRows(t *testing.T) {
	s := newInvoiceApp()
	tt := ui.NewTester(s.view, invoiceWidth, invoiceHeight)
	for _, want := range []string{"Invoice Table · 1000 rows", invoiceSubtitle, "INV-1001", "INV-2000", "Num", "Invoice", "Client", "Status", "Grand total"} {
		if !tt.HasText(want) {
			t.Errorf("missing %q, got %q", want, tt.Texts())
		}
	}
}

func TestInvoiceSelectRow(t *testing.T) {
	s := newInvoiceApp()
	tt := ui.NewTester(s.view, invoiceWidth, invoiceHeight)
	if err := tt.Click("INV-1005"); err != nil {
		t.Fatal(err)
	}
	if s.selected != 4 {
		t.Fatalf("selected = %d, want 4", s.selected)
	}
	if !tt.HasText("Selected: INV-1005") {
		t.Errorf("missing selection readout, got %q", tt.Texts())
	}
}

func TestInvoiceThemeToggle(t *testing.T) {
	s := newInvoiceApp()
	if s.theme != invoiceThemeGray {
		t.Fatalf("theme = %q, want gray", s.theme)
	}
	tt := ui.NewTester(s.view, invoiceWidth, invoiceHeight)
	if err := tt.Click(s.themeLabel()); err != nil {
		t.Fatal(err)
	}
	if s.theme != invoiceThemeLight {
		t.Fatalf("theme = %q, want light", s.theme)
	}
}

// BenchmarkInvoiceClick measures the full data-update path: each click on a
// visible row selects it and reframes the whole table view.
func BenchmarkInvoiceClick(b *testing.B) {
	s := newInvoiceApp()
	tt := ui.NewTester(s.view, invoiceWidth, invoiceHeight)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := tt.Click("INV-1005"); err != nil {
			b.Fatal(err)
		}
	}
	if s.selected != 4 {
		b.Fatalf("selected = %d, want 4", s.selected)
	}
}

// BenchmarkInvoiceTester measures building the whole thousand-row view in
// memory without a window: one tester per iteration, then a text lookup.
func BenchmarkInvoiceTester(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		s := newInvoiceApp()
		s.selected = 500
		tt := ui.NewTester(s.view, invoiceWidth, invoiceHeight)
		if !tt.HasText("INV-2000") {
			b.Fatal("missing row text")
		}
	}
}

// BenchmarkInvoiceFrame measures steady-state reframes with unchanged state
// and reports FPS so the report can compare directly with ownframe.
func BenchmarkInvoiceFrame(b *testing.B) {
	s := newInvoiceApp()
	tt := ui.NewTester(s.view, invoiceWidth, invoiceHeight)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		tt.Frame()
	}
}

package main

import (
	"fmt"
	"runtime"
	"strings"
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

func TestInvoiceFilter(t *testing.T) {
	s := newInvoiceApp()
	tt := ui.NewTester(s.view, invoiceWidth, invoiceHeight)
	if err := tt.Click("Filter by client..."); err != nil {
		t.Fatal(err)
	}
	tt.Type("acme")
	want := 0
	for _, r := range makeInvoiceRows(invoiceRows) {
		if strings.Contains(strings.ToLower(r.client), "acme") {
			want++
		}
	}
	if s.query != "acme" {
		t.Fatalf("query = %q, want acme", s.query)
	}
	if wantText := "Showing " + fmt.Sprint(want) + " of " + fmt.Sprint(invoiceRows); !tt.HasText(wantText) {
		t.Errorf("missing readout %q", wantText)
	}
}

// BenchmarkInvoiceColdBuild measures a fresh app plus its first headless
// frame: the closest headless proxy for first paint. Toolkit caches persist
// across iterations in-process, so treat this as a cold view on a warm
// engine.
func BenchmarkInvoiceColdBuild(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		s := newInvoiceApp()
		tt := ui.NewTester(s.view, invoiceWidth, invoiceHeight)
		if !tt.HasText("INV-2000") {
			b.Fatal("missing row text")
		}
	}
}

// BenchmarkInvoiceResize measures alternating resizes, each settling a fresh
// frame: the cost of a window resize or orientation change.
func BenchmarkInvoiceResize(b *testing.B) {
	s := newInvoiceApp()
	tt := ui.NewTester(s.view, invoiceWidth, invoiceHeight)
	sizes := [][2]int{{invoiceWidth, invoiceHeight}, {1200, 800}}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		tt.SetSize(sizes[i%len(sizes)][0], sizes[i%len(sizes)][1])
	}
}

// BenchmarkInvoiceTyping measures one keystroke into the filter box plus the
// reframe it triggers, alternating a typed char and backspace so every
// iteration does constant work: the headless proxy for typing latency.
func BenchmarkInvoiceTyping(b *testing.B) {
	s := newInvoiceApp()
	tt := ui.NewTester(s.view, invoiceWidth, invoiceHeight)
	if err := tt.Click("Filter by client..."); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		if i%2 == 0 {
			tt.Type("a")
			continue
		}
		tt.Key(0, ui.KeyBackspace)
	}
	for s.query != "" {
		tt.Key(0, ui.KeyBackspace)
	}
	if s.query != "" {
		b.Fatalf("query = %q, want empty after balanced typing", s.query)
	}
}

// BenchmarkInvoiceHeapGrowth builds full views, collects, and reports the
// retained heap. Run with -benchtime=1x: retention is a steady-state
// property, one measurement suffices.
func BenchmarkInvoiceHeapGrowth(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		runtime.GC()
		var before runtime.MemStats
		runtime.ReadMemStats(&before)
		for i := 0; i < 10; i++ {
			s := newInvoiceApp()
			tt := ui.NewTester(s.view, invoiceWidth, invoiceHeight)
			if !tt.HasText("INV-2000") {
				b.Fatal("missing row text")
			}
		}
		var mem runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&mem)
		b.ReportMetric(float64(mem.HeapAlloc), "heap-retained-bytes")
		b.ReportMetric(float64(mem.HeapObjects), "heap-retained-objects")
		b.ReportMetric(float64(int64(mem.HeapAlloc)-int64(before.HeapAlloc)), "heap-growth-bytes")
		b.ReportMetric(float64(int64(mem.HeapObjects)-int64(before.HeapObjects)), "heap-growth-objects")
	}
}

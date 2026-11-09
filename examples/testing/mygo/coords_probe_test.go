package main

// Probe: report client centers of the first two data rows through the same
// headless layout the live window renders. Measurement scaffolding for
// scripts/mygo-guibench.py, which clicks those points in a real window:
// GOWORK=off go test -run TestPrintRowCenters -v .
import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestPrintRowCenters(t *testing.T) {
	s := newInvoiceApp()
	tt := ui.NewTester(s.view, invoiceWidth, invoiceHeight)
	for _, want := range []string{"INV-1001", "INV-1002"} {
		r, ok := tt.Find(want)
		if !ok {
			t.Fatalf("no rect for %s", want)
		}
		t.Logf("CENTER %s %.1f %.1f", want, r.X+r.W/2, r.Y+r.H/2)
	}
}

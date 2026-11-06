// Windowed-table tests for the invoice app.
//
// The windowed table mode renders a viewport-sized slice of rows framed by
// the invoice-pad-top/invoice-pad-bot spacers once setWindowed(start)
// positions the window. That API lands separately; each test below skips
// until *invoiceApp implements setWindowed, so this file compiles and
// passes (by skipping) both before and after that change.
package main

import (
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/pagetest"
)

// invoiceWindower is the windowed-table API: setWindowed positions the
// viewport-sized row slice, clamping out-of-range starts into range.
type invoiceWindower interface {
	setWindowed(start int)
}

// requireWindowing returns the app as an invoiceWindower, skipping the
// caller while the windowing API has not landed yet.
func requireWindowing(t *testing.T, a *invoiceApp) invoiceWindower {
	t.Helper()
	w, ok := any(a).(invoiceWindower)
	if !ok {
		t.Skip("windowing not yet implemented: invoiceApp.setWindowed missing")
	}
	return w
}

// showWindow positions the window at start and redraws so the page HTML
// reflects the windowed slice.
func showWindow(t *testing.T, a *invoiceApp, w invoiceWindower, start int) {
	t.Helper()
	ctx := context.Background()
	w.setWindowed(start)
	if err := a.refresh(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}
}

// countWindowRows counts rendered table rows by their row-id prefix.
func countWindowRows(html string) int {
	return strings.Count(html, "invoice-row-")
}

// TestInvoiceWindowClamp checks that setWindowed clamps out-of-range starts:
// a negative start shows the first rows, a past-the-end start shows the
// last row.
func TestInvoiceWindowClamp(t *testing.T) {
	t.Parallel()
	a, err := newInvoiceApp()
	if err != nil {
		t.Fatal(err)
	}
	if err := a.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	w := requireWindowing(t, a)
	last := len(a.rows) - 1

	showWindow(t, a, w, -50)
	if html := a.Page().HTML(); !strings.Contains(html, "invoice-row-0") {
		t.Errorf("negative start: windowed HTML missing invoice-row-0, want clamp to 0")
	}

	showWindow(t, a, w, len(a.rows)+500)
	if html := a.Page().HTML(); !strings.Contains(html, "invoice-row-"+strconv.Itoa(last)) {
		t.Errorf("overflow start: windowed HTML missing invoice-row-%d, want clamp to last window", last)
	}
}

// TestInvoiceWindowHTML checks that a mid-table window renders the window
// rows plus both spacers while far rows are sliced out of the document.
func TestInvoiceWindowHTML(t *testing.T) {
	t.Parallel()
	a, err := newInvoiceApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	fullRows := countWindowRows(a.Page().HTML())
	if fullRows != len(a.rows) {
		t.Fatalf("full HTML has %d rows, want %d", fullRows, len(a.rows))
	}
	w := requireWindowing(t, a)
	before := a.Page().Probe().Stats()
	const start = 500
	showWindow(t, a, w, start)
	after := a.Page().Probe().Stats()
	if after.Layouts <= before.Layouts {
		t.Errorf("windowed refresh laid nothing out")
	}
	html := a.Page().HTML()
	for _, want := range []string{
		"invoice-pad-top",
		"invoice-pad-bot",
		"invoice-row-" + strconv.Itoa(start),
		"invoice-row-" + strconv.Itoa(start+1),
	} {
		if !strings.Contains(html, want) {
			t.Errorf("windowed HTML missing %q", want)
		}
	}
	for _, far := range []string{"invoice-row-0", "invoice-row-999"} {
		if strings.Contains(html, far) {
			t.Errorf("windowed HTML contains far row %q, want it sliced out", far)
		}
	}
	if got := countWindowRows(html); got == 0 || got >= fullRows {
		t.Errorf("windowed HTML has %d rows, want a non-empty viewport slice of %d", got, fullRows)
	}
}

// TestInvoiceWindowSelect checks that clicking a visible windowed row
// selects it and updates the readout. Selection patches the retained
// display without rewriting HTML source, so this test pins the rebuild
// path for its HTML assertions.
func TestInvoiceWindowSelect(t *testing.T) {
	t.Parallel()
	a, err := newInvoiceApp()
	if err != nil {
		t.Fatal(err)
	}
	a.patch = false
	ctx := context.Background()
	if err := a.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	w := requireWindowing(t, a)
	const start, target = 500, 505
	showWindow(t, a, w, start)
	if html := a.Page().HTML(); !strings.Contains(html, "invoice-row-"+strconv.Itoa(target)) {
		t.Fatalf("row %d not rendered in window at %d, cannot test visible selection", target, start)
	}
	if err := pagetest.ClickAction(ctx, a.Page(), "row:"+strconv.Itoa(target)); err != nil {
		t.Fatal(err)
	}
	if a.selected != target {
		t.Fatalf("selected = %d, want %d", a.selected, target)
	}
	html := a.Page().HTML()
	if want := a.selectedText(); !strings.Contains(html, want) {
		t.Errorf("windowed HTML missing selection readout %q", want)
	}
	if !strings.Contains(html, "inv-sel") {
		t.Errorf("windowed HTML missing selected-row class")
	}
}

// TestInvoiceWindowTotalParity checks that windowing only slices the view:
// a windowed app reports the same grand total as a full app, both in data
// and in the invoice-total readout.
func TestInvoiceWindowTotalParity(t *testing.T) {
	t.Parallel()
	full, err := newInvoiceApp()
	if err != nil {
		t.Fatal(err)
	}
	windowed, err := newInvoiceApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := full.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := windowed.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	w := requireWindowing(t, windowed)
	showWindow(t, windowed, w, 500)
	if windowed.grandTotal() != full.grandTotal() {
		t.Errorf("windowed total = %d, want full total %d",
			windowed.grandTotal(), full.grandTotal())
	}
	html := windowed.Page().HTML()
	if want := "Grand total: " + formatCents(full.grandTotal()); !strings.Contains(html, want) {
		t.Errorf("windowed HTML missing total readout %q", want)
	}
	if !strings.Contains(html, "invoice-total") {
		t.Errorf("windowed HTML missing invoice-total readout")
	}
}

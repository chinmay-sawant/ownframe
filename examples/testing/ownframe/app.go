// Package main is the ownframe invoice table built with the UI builder.
//
// Thousand-row invoice with deterministic random data, alternating gray and
// light row themes, click-to-select, and a theme toggle. It measures the
// default full-rebuild path against the mygo invoice table.
package main

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/ui"
)

const (
	invoiceTitle         = "Invoice Table"
	invoiceSubtitle      = "ownframe · builder → HTML → blinkless → Ebiten"
	invoiceWidth         = 1000
	invoiceHeight        = 650
	invoiceRows          = 1000
	invoiceNumColWidth   = 48
	invoiceIDColWidth    = 96
	invoiceClientWidth   = 140
	invoiceDateColWidth  = 96
	invoiceItemColWidth  = 170
	invoiceQtyColWidth   = 48
	invoiceTotalColWidth = 96
	invoiceStatusWidth   = 90
)

// invoiceWindowSize is the windowed row count: screen renders at most this
// many rows plus top/bottom spacers. invoiceRowH estimates one row's height
// in CSS pixels (8px padding top and bottom around a ~16px text line); the
// spacers multiply it by the hidden row counts to keep the scroll range at
// the full-table value. Recalibrate against laid-out boxes when scroll
// wiring lands.
const (
	invoiceWindowSize = 100
	invoiceRowH       = 32
)

// Invoice themes: gray is the zebra stripe (white/gray), light is plain white.
const (
	invoiceThemeGray  = "gray"
	invoiceThemeLight = "light"
)

// invoiceThemeCSS paints row backgrounds by class so the patch path can swap
// a class instead of rebuilding the table. Layout stays inline; only the
// background comes from these rules.
const invoiceThemeCSS = `.inv-row{background-color:#ffffff}` +
	`.inv-alt{background-color:#f3f4f6}` +
	`.inv-sel{background-color:#dbeafe}`

var (
	invoiceClients  = [...]string{"Acme Corp", "Globex", "Soylent", "Initech", "Umbrella", "Hooli", "Stark Industries", "Wayne Enterprises", "Massive Dynamic", "Cyberdyne"}
	invoiceItems    = [...]string{"Widget A", "Gadget B", "Service C", "License D", "Support E", "Consulting F", "Hosting G", "Audit H"}
	invoiceStatuses = [...]string{"Paid", "Pending", "Overdue"}
)

// invoiceRow is one invoice line. Keep field order identical to the mygo copy:
// the rng call sequence below must match it exactly.
type invoiceRow struct {
	num        string
	id         string
	client     string
	date       string
	item       string
	qty        int
	unitCents  int
	totalCents int
	status     string
}

// makeInvoiceRows builds n deterministic rows from seed 1.
func makeInvoiceRows(n int) []invoiceRow {
	rng := rand.New(rand.NewSource(1))
	rows := make([]invoiceRow, n)
	for i := range rows {
		client := invoiceClients[rng.Intn(len(invoiceClients))]
		item := invoiceItems[rng.Intn(len(invoiceItems))]
		month := 1 + rng.Intn(12)
		day := 1 + rng.Intn(28)
		qty := 1 + rng.Intn(10)
		unit := 900 + rng.Intn(49101)
		status := invoiceStatuses[rng.Intn(len(invoiceStatuses))]
		rows[i] = invoiceRow{
			num:        "#" + strconv.Itoa(i+1),
			id:         fmt.Sprintf("INV-%04d", 1001+i),
			client:     client,
			date:       fmt.Sprintf("2026-%02d-%02d", month, day),
			item:       item,
			qty:        qty,
			unitCents:  unit,
			totalCents: qty * unit,
			status:     status,
		}
	}
	return rows
}

func formatCents(c int) string {
	return fmt.Sprintf("$%d.%02d", c/100, c%100)
}

func invoiceStatusColor(status string) ui.Color {
	switch status {
	case "Paid":
		return ui.Hex("#15803d")
	case "Pending":
		return ui.Hex("#b45309")
	default:
		return ui.Hex("#b91c1c")
	}
}

// invoiceApp is the thousand-row invoice: header, dark column header row,
// zebra rows with a gray/light theme switch, and click-to-select.
//
// patch switches the click and theme paths from full rebuilds to element
// patches: two class edits plus a text edit for selection, one class edit
// per row for the theme, each falling back to a full refresh on refusal.
// patch is on by default; the rebuild path and refresh stay as the fallback
// and regression baseline.
//
// windowed renders only the windowStart slice of rows plus top/bottom
// spacers that keep the scroll range at the full-table value. Off by
// default: non-windowed output is byte-identical to the full table.
type invoiceApp struct {
	view        *ui.View
	page        *ownframe.Page
	rows        []invoiceRow
	selected    int
	theme       string
	patch       bool
	windowed    bool
	windowStart int
	// version bumps on every memory-state mutation; renderedVersion records
	// the version whose screen() output refresh() last committed. refresh()
	// skips the rebuild when they match. Patch paths that bypass refresh
	// still bump version so a later refresh is not wrongly skipped.
	version         uint64
	renderedVersion uint64
}

func newInvoiceApp() (*invoiceApp, error) {
	return newInvoiceAppWithCount(invoiceRows)
}

// newInvoiceAppWithCount builds the table with n rows. Tests use a small
// count for batch paths; benchmarks use the full thousand.
func newInvoiceAppWithCount(n int) (*invoiceApp, error) {
	a := &invoiceApp{selected: -1, theme: invoiceThemeGray, rows: makeInvoiceRows(n), patch: true}
	view, err := ui.Mount(ownframe.Config{
		Title:       invoiceTitle,
		Width:       invoiceWidth,
		Height:      invoiceHeight,
		RetainState: true,
		Theme:       invoiceThemeCSS,
	}, a.screen())
	if err != nil {
		return nil, fmt.Errorf("new invoice table app: %w", err)
	}
	a.view = view
	a.page = view.Page
	view.Handle(ownframe.Handlers{Click: a.handleClick})
	return a, nil
}

// Page returns the page Run and Serve display.
func (a *invoiceApp) Page() *ownframe.Page { return a.page }

func (a *invoiceApp) selectedText() string {
	if a.selected < 0 || a.selected >= len(a.rows) {
		return "Selected: none"
	}
	r := a.rows[a.selected]
	return "Selected: " + r.id + " · " + r.client + " · " + formatCents(r.totalCents)
}

func (a *invoiceApp) themeText() string {
	if a.theme == invoiceThemeLight {
		return "Theme: light (plain white rows)"
	}
	return "Theme: gray (alternating white/gray rows)"
}

func (a *invoiceApp) themeButtonLabel() string {
	if a.theme == invoiceThemeLight {
		return "Theme: light — switch to gray"
	}
	return "Theme: gray — switch to light"
}

func (a *invoiceApp) grandTotal() int {
	sum := 0
	for _, r := range a.rows {
		sum += r.totalCents
	}
	return sum
}

// rowClass is the class-side twin of the zebra rule: the theme stylesheet
// paints these, so a selection or theme change is a class edit, not a
// rebuild.
func (a *invoiceApp) rowClass(i int) string {
	if i == a.selected {
		return "inv-row inv-sel"
	}
	if a.theme == invoiceThemeLight {
		return "inv-row"
	}
	if i%2 == 1 {
		return "inv-row inv-alt"
	}
	return "inv-row"
}

func (a *invoiceApp) headerCell(text string, width float64) *ui.Element {
	return ui.Text(text).Width(ui.Px(width)).
		FontSize(ui.Px(13)).FontWeight(ui.FontWeightBold).
		Color(ui.Hex("#ffffff"))
}

func (a *invoiceApp) cell(text string, width float64) *ui.Element {
	return ui.Text(text).Width(ui.Px(width)).FontSize(ui.Px(13))
}

func (a *invoiceApp) dimCell(text string, width float64) *ui.Element {
	return ui.Text(text).Width(ui.Px(width)).FontSize(ui.Px(13)).Color(ui.Hex("#6b7280"))
}

func (a *invoiceApp) screen() *ui.Element {
	header := ui.Div().Display(ui.DisplayFlex).
		FlexDirection(ui.FlexDirectionColumn).
		Gap(ui.Px(4)).
		Padding(ui.Px(12)).
		Children(
			ui.Text("Invoice Table · 1000 rows").FontSize(ui.Px(18)).FontWeight(ui.FontWeightBold),
			ui.Text(invoiceSubtitle).FontSize(ui.Px(12)).Color(ui.Hex("#6b7280")),
			ui.Text(a.selectedText()).ID("invoice-selected").FontSize(ui.Px(14)),
			ui.Text(a.themeText()).ID("invoice-theme-label").FontSize(ui.Px(13)).Color(ui.Hex("#6b7280")),
			ui.Text("Grand total: "+formatCents(a.grandTotal())).ID("invoice-total").FontSize(ui.Px(14)).FontWeight(ui.FontWeightBold),
			ui.Button(a.themeButtonLabel()).ID("invoice-theme").Attr("data-action", "theme"),
		)
	columns := ui.Div().Display(ui.DisplayFlex).
		Gap(ui.Px(12)).
		Padding(ui.Px(8)).
		BackgroundColor(ui.Hex("#1c1915")).
		Children(
			a.headerCell("Num", invoiceNumColWidth),
			a.headerCell("Invoice", invoiceIDColWidth),
			a.headerCell("Client", invoiceClientWidth),
			a.headerCell("Date", invoiceDateColWidth),
			a.headerCell("Item", invoiceItemColWidth),
			a.headerCell("Qty", invoiceQtyColWidth),
			a.headerCell("Total", invoiceTotalColWidth),
			a.headerCell("Status", invoiceStatusWidth),
		)
	start, end := 0, len(a.rows)
	if a.windowed {
		start, end = a.windowRange()
	}
	rows := make([]*ui.Element, 0, end-start)
	for i := start; i < end; i++ {
		r := a.rows[i]
		row := ui.Div().ID("invoice-row-"+strconv.Itoa(i)).
			Attr("data-action", "row:"+strconv.Itoa(i)).
			Class(a.rowClass(i)).
			Display(ui.DisplayFlex).
			Gap(ui.Px(12)).
			Padding(ui.Px(8)).
			Children(
				a.dimCell(r.num, invoiceNumColWidth),
				a.cell(r.id, invoiceIDColWidth),
				a.cell(r.client, invoiceClientWidth),
				a.dimCell(r.date, invoiceDateColWidth),
				a.cell(r.item, invoiceItemColWidth),
				a.cell(strconv.Itoa(r.qty), invoiceQtyColWidth),
				a.cell(formatCents(r.totalCents), invoiceTotalColWidth),
				ui.Text(r.status).Width(ui.Px(invoiceStatusWidth)).FontSize(ui.Px(13)).Color(invoiceStatusColor(r.status)),
			)
		rows = append(rows, row)
	}
	grid := rows
	if a.windowed {
		padded := make([]*ui.Element, 0, len(rows)+2)
		padded = append(padded, ui.Div().ID("invoice-pad-top").Height(ui.Px(float64(start*invoiceRowH))))
		padded = append(padded, rows...)
		padded = append(padded, ui.Div().ID("invoice-pad-bot").Height(ui.Px(float64((len(a.rows)-end)*invoiceRowH))))
		grid = padded
	}
	return ui.Div().Display(ui.DisplayFlex).
		FlexDirection(ui.FlexDirectionColumn).
		Width(ui.Percent(100)).
		Children(header, columns, ui.Div().ID("invoice-grid").Children(grid...))
}

// windowRange clamps the visible slice to [start, start+invoiceWindowSize)
// inside the rows. Short tables show every row from zero.
func (a *invoiceApp) windowRange() (int, int) {
	total := len(a.rows)
	if total <= invoiceWindowSize {
		return 0, total
	}
	start := a.windowStart
	if start < 0 {
		start = 0
	}
	if start > total-invoiceWindowSize {
		start = total - invoiceWindowSize
	}
	return start, start + invoiceWindowSize
}

// setWindowed enables windowing and moves the window to start, clamped into
// range so the window always covers a full invoiceWindowSize slice (or the
// whole table when it is shorter).
func (a *invoiceApp) setWindowed(start int) {
	a.windowed = true
	a.version++
	total := len(a.rows)
	if total <= invoiceWindowSize {
		a.windowStart = 0
		return
	}
	if start < 0 {
		start = 0
	}
	if start > total-invoiceWindowSize {
		start = total - invoiceWindowSize
	}
	a.windowStart = start
}

func (a *invoiceApp) handleClick(ctx context.Context, box ownframe.Box) error {
	if box.Action == "theme" {
		if a.patch {
			return a.patchTheme(ctx)
		}
		if a.theme == invoiceThemeGray {
			a.theme = invoiceThemeLight
		} else {
			a.theme = invoiceThemeGray
		}
		a.version++
		return a.refresh(ctx)
	}
	n, ok := strings.CutPrefix(box.Action, "row:")
	if !ok {
		return nil
	}
	i, err := strconv.Atoi(n)
	if err != nil || i < 0 || i >= len(a.rows) {
		return nil
	}
	if a.patch {
		return a.patchSelectBatched(ctx, i)
	}
	a.selected = i
	a.version++
	return a.refresh(ctx)
}

func (a *invoiceApp) refresh(ctx context.Context) error {
	if a.version == a.renderedVersion {
		return nil
	}
	if err := a.view.Update(ctx, a.screen()); err != nil {
		return fmt.Errorf("update invoice table view: %w", err)
	}
	a.renderedVersion = a.version
	return nil
}

// patchSelectBatched moves the selection with two class edits and a readout
// text edit instead of rebuilding the table, committed through PatchElements
// in ONE call: all edits share a single style re-resolve, relayout, and
// draw. Any failure falls back to a full refresh from memory state, so the
// page never goes stale.
func (a *invoiceApp) patchSelectBatched(ctx context.Context, i int) error {
	old := a.selected
	a.selected = i
	a.version++
	patches := make([]ownframe.ElementPatch, 0, 3)
	if old >= 0 && old < len(a.rows) && old != i {
		patches = append(patches, ownframe.ElementPatch{
			ID:    "invoice-row-" + strconv.Itoa(old),
			Edits: []ownframe.ElementEdit{{Attribute: "class", Value: a.rowClass(old)}},
		})
	}
	patches = append(patches, ownframe.ElementPatch{
		ID:    "invoice-row-" + strconv.Itoa(i),
		Edits: []ownframe.ElementEdit{{Attribute: "class", Value: a.rowClass(i)}},
	})
	text := a.selectedText()
	patches = append(patches, ownframe.ElementPatch{
		ID:    "invoice-selected",
		Edits: []ownframe.ElementEdit{{Text: &text}},
	})
	if err := a.page.PatchElements(ctx, patches); err != nil {
		return a.refresh(ctx)
	}
	return nil
}

// themeCSSFor returns the row-background stylesheet for a theme. Swapping
// the stylesheet restyles every row at once: no element edits, no reparse,
// one cascade, layout, and paint.
func themeCSSFor(theme string) string {
	if theme == invoiceThemeLight {
		return `.inv-row{background-color:#ffffff}` +
			`.inv-alt{background-color:#ffffff}` +
			`.inv-sel{background-color:#dbeafe}`
	}
	return invoiceThemeCSS
}

// styleThemeSwap flips the theme by swapping the row stylesheet and drawing
// once. It is the global-change counterpart to patchSelect: one restyle
// instead of a thousand element patches or a full rebuild.
func (a *invoiceApp) styleThemeSwap(ctx context.Context) error {
	if a.theme == invoiceThemeGray {
		a.theme = invoiceThemeLight
	} else {
		a.theme = invoiceThemeGray
	}
	a.version++
	if err := a.page.SetTheme(themeCSSFor(a.theme)); err != nil {
		return err
	}
	return a.page.Redraw(ctx)
}

// patchTheme flips the theme with one class edit per row inside a redraw
// batch, so all thousand edits commit and draw once.
func (a *invoiceApp) patchTheme(ctx context.Context) error {
	if a.theme == invoiceThemeGray {
		a.theme = invoiceThemeLight
	} else {
		a.theme = invoiceThemeGray
	}
	a.version++
	a.page.Probe().BeginRedrawBatch()
	var firstErr error
	for i := range a.rows {
		err := a.page.PatchElement(ctx, "invoice-row-"+strconv.Itoa(i),
			ownframe.ElementEdit{Attribute: "class", Value: a.rowClass(i)})
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	themeText := a.themeText()
	if err := a.page.PatchElement(ctx, "invoice-theme-label", ownframe.ElementEdit{Text: &themeText}); err != nil && firstErr == nil {
		firstErr = err
	}
	buttonText := a.themeButtonLabel()
	if err := a.page.PatchElement(ctx, "invoice-theme", ownframe.ElementEdit{Text: &buttonText}); err != nil && firstErr == nil {
		firstErr = err
	}
	if err := a.page.Probe().EndRedrawBatch(ctx); err != nil && firstErr == nil {
		firstErr = err
	}
	if firstErr != nil {
		return a.refresh(ctx)
	}
	return nil
}

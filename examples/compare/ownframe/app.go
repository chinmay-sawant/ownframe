// Package main is the ownframe compare table built with the UI builder.
//
// Naive form: the builder renders all 1000 rows every time, so this measures
// the default full-rebuild path against the mygo naive table.
package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/ui"
)

const (
	listTitle         = "Compare Table"
	listSubtitle      = "ownframe · builder → HTML → blinkless → Ebiten"
	listWidth         = 800
	listHeight        = 600
	listRows          = 1000
	listBody          = "Lorem ipsum dolor sit amet, consectetur adipiscing elit."
	listNumColWidth   = 64
	listTitleColWidth = 200
)

// listRow is one styled table row.
type listRow struct {
	num   string
	title string
	body  string
}

// listApp is the thousand-row styled table: a header, a dark column header
// row, and one alternating-color row per item. Clicking a row selects it and
// rebuilds the whole view with the new data.
type listApp struct {
	view     *ui.View
	page     *ownframe.Page
	rows     []listRow
	selected int
	// rowCache retains built row elements by index for the selective refresh
	// path. Unchanged rows are reattached by pointer so the builder diff
	// takes its pointer-identity fast path and skips their subtrees.
	// Rebuilt rows are always fresh elements, never mutations of a cached
	// pointer, per the diff contract.
	rowCache []*ui.Element
	// cacheSel is the selection rowCache was built for; -2 means unprimed.
	cacheSel int
}

func makeListRows(n int) []listRow {
	rows := make([]listRow, n)
	for i := range rows {
		rows[i] = listRow{
			num:   "#" + strconv.Itoa(i+1),
			title: "Item " + strconv.Itoa(i+1),
			body:  listBody,
		}
	}
	return rows
}

func newListApp() (*listApp, error) {
	a := &listApp{selected: -1, rows: makeListRows(listRows), cacheSel: -2}
	view, err := ui.Mount(ownframe.Config{
		Title:  listTitle,
		Width:  listWidth,
		Height: listHeight,
	}, a.screen())
	if err != nil {
		return nil, fmt.Errorf("new compare table app: %w", err)
	}
	a.view = view
	a.page = view.Page
	view.Handle(ownframe.Handlers{Click: a.handleClick})
	return a, nil
}

// Page returns the page Run and Serve display.
func (a *listApp) Page() *ownframe.Page { return a.page }

func (a *listApp) selectedText() string {
	if a.selected < 0 || a.selected >= len(a.rows) {
		return "Selected: none"
	}
	return "Selected: " + a.rows[a.selected].title
}

func (a *listApp) headerCell(text string, width float64) *ui.Element {
	return ui.Text(text).Width(ui.Px(width)).
		FontSize(ui.Px(13)).FontWeight(ui.FontWeightBold).
		Color(ui.Hex("#ffffff"))
}

func (a *listApp) screenHeader() *ui.Element {
	return ui.Div().Display(ui.DisplayFlex).
		FlexDirection(ui.FlexDirectionColumn).
		Gap(ui.Px(4)).
		Padding(ui.Px(12)).
		Children(
			ui.Text("Compare Table · 1000 rows").FontSize(ui.Px(18)).FontWeight(ui.FontWeightBold),
			ui.Text(listSubtitle).FontSize(ui.Px(12)).Color(ui.Hex("#6b7280")),
			ui.Text(a.selectedText()).ID("compare-selected").FontSize(ui.Px(14)),
		)
}

func (a *listApp) screenColumns() *ui.Element {
	return ui.Div().Display(ui.DisplayFlex).
		Gap(ui.Px(12)).
		Padding(ui.Px(8)).
		BackgroundColor(ui.Hex("#1c1915")).
		Children(
			a.headerCell("Num", listNumColWidth),
			a.headerCell("Title", listTitleColWidth),
			ui.Text("Description").FontSize(ui.Px(13)).
				FontWeight(ui.FontWeightBold).Color(ui.Hex("#ffffff")),
		)
}

// buildRow returns a fresh element for row i with the background for the
// current selection. The caller may cache the result, but must never mutate
// it afterwards: the diff treats a reused pointer as unchanged.
func (a *listApp) buildRow(i int) *ui.Element {
	r := a.rows[i]
	bg := ui.Hex("#ffffff")
	if i%2 == 1 {
		bg = ui.Hex("#f3f4f6")
	}
	if i == a.selected {
		bg = ui.Hex("#dbeafe")
	}
	return ui.Div().ID("compare-row-"+strconv.Itoa(i)).
		Attr("data-action", "row:"+strconv.Itoa(i)).
		Display(ui.DisplayFlex).
		Gap(ui.Px(12)).
		Padding(ui.Px(8)).
		BackgroundColor(bg).
		Children(
			ui.Text(r.num).Width(ui.Px(listNumColWidth)).
				FontSize(ui.Px(13)).Color(ui.Hex("#8a8578")),
			ui.Text(r.title).Width(ui.Px(listTitleColWidth)).
				FontSize(ui.Px(14)).FontWeight(ui.FontWeightBold),
			ui.Text(r.body).FontSize(ui.Px(13)).Color(ui.Hex("#4a463c")),
		)
}

func (a *listApp) screen() *ui.Element {
	rows := make([]*ui.Element, 0, len(a.rows))
	for i := range a.rows {
		rows = append(rows, a.buildRow(i))
	}
	return ui.Div().Display(ui.DisplayFlex).
		FlexDirection(ui.FlexDirectionColumn).
		Width(ui.Percent(100)).
		Children(a.screenHeader(), a.screenColumns(), ui.Div().ID("compare-grid").Children(rows...))
}

// screenSelective rebuilds only what a selection change can affect: the
// header (fresh, carrying the readout under the stable id
// "compare-selected"), the previously selected row, and the newly selected
// row. Every other row is reattached from rowCache by pointer so the diff
// pointer-identity fast path skips its subtree.
func (a *listApp) screenSelective() *ui.Element {
	if len(a.rowCache) != len(a.rows) {
		a.rowCache = make([]*ui.Element, len(a.rows))
		a.cacheSel = -2
	}
	prev := a.cacheSel
	rows := make([]*ui.Element, 0, len(a.rows))
	for i := range a.rows {
		if cached := a.rowCache[i]; cached != nil && i != prev && i != a.selected {
			rows = append(rows, cached)
			continue
		}
		fresh := a.buildRow(i)
		a.rowCache[i] = fresh
		rows = append(rows, fresh)
	}
	a.cacheSel = a.selected
	return ui.Div().Display(ui.DisplayFlex).
		FlexDirection(ui.FlexDirectionColumn).
		Width(ui.Percent(100)).
		Children(a.screenHeader(), a.screenColumns(), ui.Div().ID("compare-grid").Children(rows...))
}

func (a *listApp) handleClick(ctx context.Context, box ownframe.Box) error {
	n, ok := strings.CutPrefix(box.Action, "row:")
	if !ok {
		return nil
	}
	i, err := strconv.Atoi(n)
	if err != nil || i < 0 || i >= len(a.rows) {
		return nil
	}
	a.selected = i
	return a.refresh(ctx)
}

func (a *listApp) refresh(ctx context.Context) error {
	if err := a.view.Update(ctx, a.screen()); err != nil {
		return fmt.Errorf("update compare table view: %w", err)
	}
	return nil
}

// refreshSelective updates the view through the fingerprint-gated path: a
// repeated selection returns before building anything, while a changed
// selection rebuilds only the header readout and the two affected rows and
// lets the diff fast path skip the other 998 cached rows. The fingerprint
// covers the selection, the only screen input that varies in this harness
// (rows are built once and never mutated).
func (a *listApp) refreshSelective(ctx context.Context) error {
	fingerprint := "sel=" + strconv.Itoa(a.selected)
	if err := a.view.UpdateChanged(ctx, fingerprint, a.screenSelective); err != nil {
		return fmt.Errorf("update compare table view selectively: %w", err)
	}
	return nil
}

// handleClickSelective is the selective twin of handleClick: same action
// parsing and validation, but the view update keeps retained row elements.
func (a *listApp) handleClickSelective(ctx context.Context, box ownframe.Box) error {
	n, ok := strings.CutPrefix(box.Action, "row:")
	if !ok {
		return nil
	}
	i, err := strconv.Atoi(n)
	if err != nil || i < 0 || i >= len(a.rows) {
		return nil
	}
	a.selected = i
	return a.refreshSelective(ctx)
}

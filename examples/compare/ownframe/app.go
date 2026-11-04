// Package main is the ownframe compare table built with the UI builder.
//
// Best-practice form: the builder renders only the viewport window of rows
// (the 1.0 list recipe: SetWindowing plus SetScrollWindow from
// examples/list), selection skips no-op clicks, and retained state stays on
// so repaint reuses caches.
package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/list"
	"github.com/chinmay-sawant/ownframe/ui"
)

const (
	listTitle    = "Compare Table"
	listSubtitle = "ownframe · builder → HTML → blinkless → Ebiten"
	listWidth    = 800
	listHeight   = 600
	listRows     = 1000
	// rowH is the fixed row height every windowing computation assumes.
	// Row content is one line and fits: 14 px type plus 16 px padding.
	rowH     = 40
	overscan = 10
	listBody = "Lorem ipsum dolor sit amet."
)

// listRow is one styled table row.
type listRow struct {
	num   string
	title string
	body  string
}

// listApp owns the full row data in Go and renders only the viewport window:
// rows winStart:winStart+winCount between topPad and botPad spacers, which
// keep the document at full height. A scroll move records a pending window
// and the tick rebuilds once.
type listApp struct {
	view          *ui.View
	page          *ownframe.Page
	rows          []listRow
	selected      int
	winStart      int
	winCount      int
	topPad        int
	botPad        int
	windowPending bool
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
	a := &listApp{selected: -1, rows: makeListRows(listRows)}
	w := list.Window(0, listHeight, 0, rowH, len(a.rows), overscan)
	a.winStart, a.winCount, a.topPad, a.botPad = w.Start, w.Count, w.Top, w.Bottom
	view, err := ui.Mount(ownframe.Config{
		Title:       listTitle,
		Width:       listWidth,
		Height:      listHeight,
		RetainState: true,
	}, a.screen())
	if err != nil {
		return nil, fmt.Errorf("new compare table app: %w", err)
	}
	a.view = view
	a.page = view.Page
	view.Handle(ownframe.Handlers{Click: a.handleClick})
	a.page.Scroll().SetWindowing(true)
	a.page.SetScrollWindow(a.slice)
	// No tick at rest: a registered tick keeps the tall viewport path on
	// the frame clock even when it reports nothing. slice installs the tick
	// only while a reslice is pending and the tick uninstalls itself after.
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

// slice moves the rendered window to the viewport. It records a pending
// rebuild and reports the move so the host repaints; the tick rebuilds once
// instead, because the scroll callback carries no context for View.Update.
func (a *listApp) slice(offsetY, viewH int) bool {
	gridY := 0
	for _, b := range a.page.Boxes() {
		if b.ID == "compare-grid" {
			gridY = int(b.Y)
			break
		}
	}
	w := list.Window(offsetY, viewH, gridY, rowH, len(a.rows), overscan)
	if w.Start == a.winStart && w.Count == a.winCount {
		return false
	}
	a.winStart, a.winCount, a.topPad, a.botPad = w.Start, w.Count, w.Top, w.Bottom
	a.windowPending = true
	a.page.SetTick(a.tick)
	return true
}

func (a *listApp) tick(ctx context.Context) error {
	if !a.windowPending {
		a.page.SetTick(nil)
		return nil
	}
	a.windowPending = false
	a.page.SetTick(nil)
	return a.refresh(ctx)
}

func (a *listApp) screen() *ui.Element {
	header := ui.Div().Display(ui.DisplayFlex).
		FlexDirection(ui.FlexDirectionColumn).
		Gap(ui.Px(4)).
		Padding(ui.Px(12)).
		Children(
			ui.Text("Compare Table · 1000 rows").FontSize(ui.Px(18)).FontWeight(ui.FontWeightBold),
			ui.Text(listSubtitle).FontSize(ui.Px(12)).Color(ui.Hex("#6b7280")),
			ui.Text(a.selectedText()).ID("compare-selected").FontSize(ui.Px(14)),
		)
	columns := ui.Div().Display(ui.DisplayFlex).
		Gap(ui.Px(12)).
		Padding(ui.Px(8)).
		BackgroundColor(ui.Hex("#1c1915")).
		Children(
			a.headerCell("Num", 64),
			a.headerCell("Title", 200),
			ui.Text("Description").FontSize(ui.Px(13)).
				FontWeight(ui.FontWeightBold).Color(ui.Hex("#ffffff")),
		)
	var top, bottom *ui.Element
	if a.topPad > 0 {
		top = ui.Div().ID("compare-top-pad").Height(ui.Px(float64(a.topPad)))
	}
	if a.botPad > 0 {
		bottom = ui.Div().ID("compare-bot-pad").Height(ui.Px(float64(a.botPad)))
	}
	end := min(a.winStart+a.winCount, len(a.rows))
	rows := make([]*ui.Element, 0, end-a.winStart+2)
	if top != nil {
		rows = append(rows, top)
	}
	for i := a.winStart; i < end; i++ {
		r := a.rows[i]
		bg := ui.Hex("#ffffff")
		if i%2 == 1 {
			bg = ui.Hex("#f3f4f6")
		}
		if i == a.selected {
			bg = ui.Hex("#dbeafe")
		}
		row := ui.Div().ID("compare-row-"+strconv.Itoa(i)).
			Attr("data-action", "row:"+strconv.Itoa(i)).
			Display(ui.DisplayFlex).
			AlignItems(ui.AlignItemsCenter).
			Gap(ui.Px(12)).
			Padding(ui.Px(8)).
			Height(ui.Px(rowH)).
			BackgroundColor(bg).
			Children(
				ui.Text(r.num).Width(ui.Px(64)).
					FontSize(ui.Px(13)).Color(ui.Hex("#8a8578")),
				ui.Text(r.title).Width(ui.Px(200)).
					FontSize(ui.Px(14)).FontWeight(ui.FontWeightBold),
				ui.Text(r.body).FontSize(ui.Px(13)).Color(ui.Hex("#4a463c")),
			)
		rows = append(rows, row)
	}
	if bottom != nil {
		rows = append(rows, bottom)
	}
	return ui.Div().Display(ui.DisplayFlex).
		FlexDirection(ui.FlexDirectionColumn).
		Width(ui.Percent(100)).
		Children(header, columns, ui.Div().ID("compare-grid").Children(rows...))
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
	if i == a.selected {
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

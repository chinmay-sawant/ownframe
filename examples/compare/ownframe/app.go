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
			a.headerCell("Num", listNumColWidth),
			a.headerCell("Title", listTitleColWidth),
			ui.Text("Description").FontSize(ui.Px(13)).
				FontWeight(ui.FontWeightBold).Color(ui.Hex("#ffffff")),
		)
	rows := make([]*ui.Element, 0, len(a.rows))
	for i, r := range a.rows {
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
		rows = append(rows, row)
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
	a.selected = i
	return a.refresh(ctx)
}

func (a *listApp) refresh(ctx context.Context) error {
	if err := a.view.Update(ctx, a.screen()); err != nil {
		return fmt.Errorf("update compare table view: %w", err)
	}
	return nil
}

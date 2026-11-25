package main

import (
	"strconv"

	"github.com/chinmay-sawant/ownframe/ui"
)

// sheetBar is the bottom tab strip: one tab per sheet plus an add button.
func (a *app) sheetBar() *ui.Element {
	bar := ui.Div().ID("xl-tabs").Class("xl-tabs")
	for i, sh := range a.wb.Sheets {
		cls := "xl-tab"
		if i == a.wb.Active {
			cls += " xl-tab--active"
		}
		bar.Children(ui.Div().ID("xl-tab-"+itoa(i)).Class(cls).
			Attr("role", "button").Attr("aria-label", sh.Name).Attr("aria-pressed", strconv.FormatBool(i == a.wb.Active)).Attr("data-action", "sheet:"+itoa(i)).Text(sh.Name))
	}
	bar.Children(titleBtn("xl-tab-add", "Add sheet", "+", "sheet:add"))
	return bar
}

// sheetAction handles a "sheet:" action: a tab index or "add".
func (a *app) sheetAction(arg string) error {
	if arg == "add" {
		a.commitEdit()
		a.wb.AddSheet()
		a.switchSheet(len(a.wb.Sheets) - 1)
		return nil
	}
	i, err := strconv.Atoi(arg)
	if err != nil || i < 0 || i >= len(a.wb.Sheets) {
		return nil
	}
	a.switchSheet(i)
	return nil
}

// switchSheet activates sheet i with a fresh selection at A1.
func (a *app) switchSheet(i int) {
	a.commitEdit()
	a.wb.Active = i
	a.scrollRow, a.scrollCol = 0, 0
	a.cutRange, a.copyRange = nil, nil
	a.active, a.anchor = CellRef{}, CellRef{}
	a.sel = Range{}
	a.redraw()
}

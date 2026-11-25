package main

import (
	"strings"

	"github.com/chinmay-sawant/ownframe/ui"
)

// formulaBar is the Name Box, the cancel/commit/insert-function buttons and
// the formula input. Both inputs opt out of Tab focus so Tab reaches the
// grid key handler while a cell is being edited.
func (a *app) formulaBar() *ui.Element {
	if a.x.hideFormulaBar {
		return nil
	}
	name := ui.Input().ID("xl-name").Class("xl-namebox").Attr("type", "text").
		Attr("value", a.nameBoxText()).Attr("tabindex", "-1").Attr("aria-label", "Name Box")
	fx := ui.Input().ID("xl-fx").Class("xl-fx").Attr("type", "text").
		Attr("value", a.formulaText()).Attr("tabindex", "-1").Attr("aria-label", "Formula bar")
	return ui.Div().ID("xl-fbar").Class("xl-fbar").Children(
		name,
		ui.Div().Class("xl-fsep"),
		titleBtn("xl-fbtn", "Cancel", "×", "fx:cancel"),
		titleBtn("xl-fbtn", "Enter", "✓", "fx:commit"),
		titleBtn("xl-fbtn xl-fbtn--fx", "Insert Function", "fx", "fx:insert"),
		fx,
	)
}

// fbarH is the formula bar band height, zero when the user hides it.
func (a *app) fbarH() int {
	if a.x.hideFormulaBar {
		return 0
	}
	return a.zs(formulaBarHeight)
}

// refHighlights draws one colored box per reference in the formula being
// edited, cycling through the six reference colors in order of appearance.
func (a *app) refHighlights(cols, rows []axisSlot) []*ui.Element {
	if !a.editing || !strings.HasPrefix(a.editBuf, "=") {
		return nil
	}
	var out []*ui.Element
	for i, rg := range FormulaRefs(a.editBuf) {
		x, y, w, h, ok := rectOf(cols, rows, rg)
		if !ok {
			continue
		}
		out = append(out, place(
			ui.Div().ID("xl-gov-ref-"+itoa(i)).Class("xl-ref xl-ref-"+itoa(i%6+1)), x, y, w, h))
	}
	return out
}

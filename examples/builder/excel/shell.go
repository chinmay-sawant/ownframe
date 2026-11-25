package main

import (
	"strconv"

	"github.com/chinmay-sawant/ownframe/ui"
)

func itoa(n int) string { return strconv.Itoa(n) }

func pxStr(n int) string { return strconv.Itoa(n) + "px" }

// place pins an element at a pixel box inside its positioned parent.
func place(e *ui.Element, x, y, w, h int) *ui.Element {
	return e.CSSRaw("left", pxStr(x)).CSSRaw("top", pxStr(y)).
		CSSRaw("width", pxStr(w)).CSSRaw("height", pxStr(h))
}

// titleBtn is one clickable control on the title bar. It carries a
// data-action instead of an OnClick so CustomHeader hit-testing lets the
// button win over the window drag.
func titleBtn(class, label, glyph, action string) *ui.Element {
	return ui.Button(glyph).Class(class).
		Attr("aria-label", label).Attr("title", label).Attr("data-action", action)
}

// titleBar is the Excel chrome: badge, Quick Access Toolbar, workbook pill,
// search trigger, theme toggle and window controls.
func (a *app) titleBar() *ui.Element {
	undoClass, redoClass := "xl-tbtn", "xl-tbtn"
	if !a.wb.CanUndo() {
		undoClass += " xl-tbtn--off"
	}
	if !a.wb.CanRedo() {
		redoClass += " xl-tbtn--off"
	}
	themeGlyph, themeLabel := "Dark", "Switch to dark theme"
	if a.dark {
		themeGlyph, themeLabel = "Light", "Switch to light theme"
	}
	bar := ui.Header().ID("xl-titlebar").Class("xl-title").Children(
		ui.Div().Class("xl-badge").Text("X"),
		ui.Text("Excel").Class("xl-brand-name"),
		ui.Div().Class("xl-qat").Children(
			titleBtn("xl-tbtn", "Save", "Save", "qat:save"),
			titleBtn(undoClass, "Undo", "↶", "qat:undo"),
			titleBtn(redoClass, "Redo", "↷", "qat:redo"),
		),
		ui.Div().Class("xl-title-spacer"),
		ui.Div().ID("xl-doc-pill").Class("xl-doc-pill").Text("Quarterly_Model.xlsx"),
		ui.Div().Class("xl-title-spacer"),
		ui.Div().ID("xl-search").Class("xl-search").Attr("data-action", "search:open").
			Text("Search  (Ctrl+P)"),
		titleBtn("xl-tbtn", themeLabel, themeGlyph, "theme:toggle").Attr("id", "xl-theme"),
	)
	if !a.web {
		bar.Children(ui.Div().Class("xl-winctl").Children(
			titleBtn("xl-winbtn", "Minimize", "−", "win-min"),
			titleBtn("xl-winbtn", "Maximize", "□", "win-max"),
			titleBtn("xl-winbtn", "Close", "×", "win-close"),
		))
	}
	return bar
}

// viewportSize is the window size in CSS pixels.
func (a *app) viewportSize() (int, int) {
	w, h := a.page.Size()
	if w <= 0 {
		w = windowWidth
	}
	if h <= 0 {
		h = windowHeight
	}
	return w, h
}

// gridViewportSize is the pixel box left for the grid (headers included)
// after the fixed bands.
func (a *app) gridViewportSize() (int, int) {
	w, h := a.viewportSize()
	h -= a.zs(titleBarHeight) + a.ribbonHeight() + a.fbarH() + a.zs(sheetBarHeight) + a.zs(statusBarHeight)
	if h < 0 {
		h = 0
	}
	return w, h
}

// gridOrigin is the top-left of the grid viewport in window pixels. The laid
// out box wins when it exists, so a ribbon of a different height still hits.
func (a *app) gridOrigin() (float64, float64) {
	for _, box := range a.page.Boxes() {
		if box.ID == "xl-grid" {
			return box.X, box.Y
		}
	}
	return 0, float64(a.zs(titleBarHeight) + a.ribbonHeight() + a.fbarH())
}

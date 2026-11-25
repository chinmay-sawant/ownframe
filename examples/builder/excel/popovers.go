package main

import (
	"context"
	"regexp"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/ui"
)

// Grid-level popovers: the function autocomplete list, the comment hover
// note and the filter chevrons. They are children of the grid element, not of
// the clipped cell canvas, so they can hang over the edge of the last row.

const (
	acMaxItems    = 8
	acRowHeight   = 20
	acWidth       = 300
	noteWidth     = 200
	noteHeight    = 54
	acListPad     = 4  // vertical padding around the autocomplete list
	filterIDParts = 2  // an id of the form prefix-col splits into two parts
	acItemPad     = 6  // right padding of an autocomplete row
	noteGap       = 4  // gap between a comment note and its cell
	chevronSize   = 16 // filter chevron button side
)

var fnTail = regexp.MustCompile(`[A-Za-z][A-Za-z0-9.]*$`)

// acStart finds where the function-name token at the end of the edit buffer
// begins. It reports false unless the buffer is a formula and the token
// follows an operator, an opening parenthesis or a separator.
func acStart(buf string) (int, bool) {
	if !strings.HasPrefix(buf, "=") {
		return 0, false
	}
	loc := fnTail.FindStringIndex(buf)
	if loc == nil || loc[0] == 0 {
		return 0, false
	}
	if !strings.ContainsRune("=(,+-*/^&<>%; ", rune(buf[loc[0]-1])) {
		return 0, false
	}
	return loc[0], true
}

// acMatches lists the function names that complete the token being typed.
func (a *app) acMatches() []string {
	if !a.editing {
		return nil
	}
	at, ok := acStart(a.editBuf)
	if !ok {
		return nil
	}
	tok := strings.ToUpper(a.editBuf[at:])
	var out []string
	for _, n := range FunctionNames() {
		if strings.HasPrefix(n, tok) {
			out = append(out, n)
			if len(out) == acMaxItems {
				break
			}
		}
	}
	return out
}

// acSelected is the highlighted row, clamped to the current list.
func (a *app) acSelected(n int) int {
	return min(max(a.acSel, 0), max(n-1, 0))
}

// acMove moves the highlight by d rows, wrapping at both ends.
func (a *app) acMove(d int) {
	n := len(a.acMatches())
	if n == 0 {
		return
	}
	a.acSel = (a.acSelected(n) + d + n) % n
	a.redraw()
}

// acAccept replaces the typed token with the highlighted function name and an
// opening parenthesis.
func (a *app) acAccept(name string) {
	at, ok := acStart(a.editBuf)
	if !ok {
		return
	}
	a.editBuf = a.editBuf[:at] + name + "("
	a.acSel = 0
	a.refActive = false
	form := a.page.Form()
	form.SetFormValue("xl-editor", a.editBuf)
	form.SetFormValue("xl-fx", a.editBuf)
	a.redraw()
}

// acKey handles the keys the open list owns. It reports whether it consumed
// the key.
func (a *app) acKey(base string) bool {
	m := a.acMatches()
	if len(m) == 0 {
		return false
	}
	switch base {
	case "arrowdown":
		a.acMove(1)
	case "arrowup":
		a.acMove(-1)
	case "tab", "enter":
		a.acAccept(m[a.acSelected(len(m))])
	default:
		return false
	}
	return true
}

// autocompletePopover draws the list beneath the active cell, or above it
// when the list would run past the bottom of the grid.
func (a *app) autocompletePopover(cols, rows []axisSlot, vw, vh int) *ui.Element {
	m := a.acMatches()
	if len(m) == 0 {
		return nil
	}
	col, okc := slotFor(cols, a.active.Col)
	row, okr := slotFor(rows, a.active.Row)
	if !okc || !okr {
		return nil
	}
	hx, hy := a.rowHdrW(), a.colHdrH()
	h := len(m)*acRowHeight + acListPad
	x := min(hx+col.pos, max(vw-acWidth-scrollbarSize, 0))
	y := hy + row.pos + row.size + 1
	if y+h > vh {
		y = max(hy+row.pos-h-1, 0)
	}
	sel := a.acSelected(len(m))
	sigs := FunctionSignatures()
	list := place(ui.Div().ID("xl-ac").Attr("role", "listbox").Attr("aria-label", "Function suggestions"),
		x, y, acWidth, h)
	list.CSSRaw("position", "absolute").CSSRaw("z-index", "30").CSSRaw("box-sizing", "border-box").
		CSSRaw("padding", "2px").CSSRaw("background-color", "var(--xl-formula-bg)").
		CSSRaw("border", "1px solid var(--xl-ribbon-border)").CSSRaw("color", "var(--xl-cell-text)")
	for i, name := range m {
		label := sigs[name]
		if label == "" {
			label = name + "(...)"
		}
		item := place(ui.Div().ID("xl-ac-"+name).Attr("role", "option").Text(label), 0, i*acRowHeight, acWidth-acItemPad, acRowHeight)
		item.CSSRaw("position", "absolute").CSSRaw("box-sizing", "border-box").CSSRaw("padding", "0 6px").
			CSSRaw("line-height", pxStr(acRowHeight)).CSSRaw("font-size", "12px").CSSRaw("overflow", "hidden").
			CSSRaw("white-space", "nowrap")
		if i == sel {
			item.Attr("aria-selected", "true").CSSRaw("background-color", "var(--xl-brand-tint)").
				CSSRaw("color", "var(--xl-brand-strong)")
		}
		list.Children(item)
	}
	return list
}

// clickAutocomplete accepts the function a click landed on.
func (a *app) clickAutocomplete(box ownframe.Box) bool {
	name, ok := strings.CutPrefix(box.ID, "xl-ac-")
	if !ok || !a.editing {
		return false
	}
	a.acAccept(name)
	return true
}

// ---- comment hover note ----

// cellIDRef parses the id of a cell element ("xl-gc-R-C") or of a web-preview
// hit target for an empty cell ("xl-ge-R-C").
func cellIDRef(id string) (CellRef, bool) {
	rest, ok := strings.CutPrefix(id, "xl-gc-")
	if !ok {
		rest, ok = strings.CutPrefix(id, "xl-ge-")
	}
	if !ok {
		return CellRef{}, false
	}
	parts := strings.Split(rest, "-")
	if len(parts) != filterIDParts {
		return CellRef{}, false
	}
	r, errR := strconv.Atoi(parts[0])
	c, errC := strconv.Atoi(parts[1])
	if errR != nil || errC != nil {
		return CellRef{}, false
	}
	return CellRef{Row: r, Col: c}, true
}

// handleHover tracks the cell under the pointer so a commented cell can show
// its note. It never claims the hover: the page keeps its own hover paint.
func (a *app) handleHover(_ context.Context, _, next ownframe.Box) (bool, error) {
	var want *CellRef
	if ref, ok := cellIDRef(next.ID); ok && a.overlay == "" && a.sheet().Comment(ref.Row, ref.Col) != "" {
		want = &ref
	}
	if (want == nil) != (a.hoverNote == nil) || (want != nil && *want != *a.hoverNote) {
		a.hoverNote = want
		a.redraw()
	}
	return false, nil
}

// notePopover draws the hovered cell's comment beside it.
func (a *app) notePopover(cols, rows []axisSlot, vw, vh int) *ui.Element {
	if a.hoverNote == nil {
		return nil
	}
	text := a.sheet().Comment(a.hoverNote.Row, a.hoverNote.Col)
	col, okc := slotFor(cols, a.hoverNote.Col)
	row, okr := slotFor(rows, a.hoverNote.Row)
	if text == "" || !okc || !okr {
		return nil
	}
	hx, hy := a.rowHdrW(), a.colHdrH()
	x := hx + col.pos + col.size + noteGap
	if x+noteWidth > vw {
		x = max(hx+col.pos-noteWidth-noteGap, 0)
	}
	y := min(hy+row.pos, max(vh-noteHeight, 0))
	n := place(ui.Div().ID("xl-gnote-pop").Attr("role", "tooltip").Text(text), x, y, noteWidth, noteHeight)
	return n.CSSRaw("position", "absolute").CSSRaw("z-index", "30").CSSRaw("box-sizing", "border-box").
		CSSRaw("padding", "4px 6px").CSSRaw("overflow", "hidden").CSSRaw("font-size", "12px").
		CSSRaw("background-color", "#ffffe1").CSSRaw("color", "#1f2328").CSSRaw("border", "1px solid #8a8886")
}

// ---- filter chevrons ----

// filterChevronCols lists the columns that show a filter chevron: those with
// an active filter, and every used column while the sheet filter is on.
func (a *app) filterChevronCols() map[int]bool {
	sh := a.sheet()
	out := map[int]bool{}
	for c := 0; c < MaxCols; c++ {
		if sh.FilterAllowed(c) != nil {
			out[c] = true
		}
	}
	if a.x.filterOn {
		used := sh.UsedRange()
		for c := used.Min.Col; c <= used.Max.Col; c++ {
			out[c] = true
		}
	}
	return out
}

// filterChevron is the small button at the right edge of a column header.
func filterChevron(col, x, y int, active bool) *ui.Element {
	btn := place(ui.Div().ID("xl-gfc-"+itoa(col)).Class("xl-chev").Attr("role", "button").
		Attr("aria-label", "Filter column "+ColName(col)), x, y, chevronSize, chevronSize)
	btn.CSSRaw("position", "absolute").CSSRaw("box-sizing", "border-box").CSSRaw("z-index", "2").
		CSSRaw("background-color", "var(--xl-formula-bg)").CSSRaw("border", "1px solid var(--xl-gridline)")
	if active {
		btn.CSSRaw("border-color", "var(--xl-brand-strong)")
	}
	btn.CSSRaw("text-align", "center").CSSRaw("font-size", "10px").CSSRaw("line-height", "14px").
		CSSRaw("color", "var(--xl-header-fg)").Text("v")
	return btn
}

// clickFilterChevron opens the filter popover for the column a chevron
// belongs to.
func (a *app) clickFilterChevron(id string) bool {
	rest, ok := strings.CutPrefix(id, "xl-gfc-")
	if !ok {
		return false
	}
	col, err := strconv.Atoi(rest)
	if err != nil {
		return false
	}
	a.openFilter(col)
	return true
}

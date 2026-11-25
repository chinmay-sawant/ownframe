package main

import (
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe/ui"
)

// Every element inside the grid carries an id with this prefix, so the drag
// guard can tell grid hits from the inputs and buttons around it.
const gridIDPrefix = "xl-g"

// Grid overlay geometry in unzoomed pixels.
const (
	cellFontPx       = 13  // base cell font size before zoom
	govDropdownInset = 17  // dropdown arrow distance from the cell's right edge
	fillHandleInset  = 5   // fill handle offset from the selection corner
	editorPadding    = 24  // editor width added to the text width
	editorCharWidth  = 8   // editor width per character
	chevronInset     = 19  // filter chevron distance from the column's right edge
	tipOffset        = 8   // hover tip offset from the header corner
	lumaThreshold    = 140 // luma above which text on a fill turns dark
	colorByteMask    = 0xff
	govDropdownWidth = 16 // dropdown arrow button width
	fillHandleSize   = 6  // fill handle side
	chevronTop       = 4  // filter chevron top inset in the header
	tipWidth         = 120
	tipHeight        = 20
	thumbInset       = 2 // scrollbar thumb inset from the track edge
	thumbShrink      = 4 // track thickness deducted for the thumb
	frozenLineWidth  = 2 // frozen-pane divider thickness
	denseCellDivisor = 2 // expected share of visible cells that carry content
	dropdownTop      = 1 // dropdown arrow top inset
	dropdownShrink   = 3 // dropdown arrow height deducted from the cell
)

func isGridID(id string) bool {
	return id == "xl-grid" || strings.HasPrefix(id, gridIDPrefix)
}

// gridView draws only the visible window of the sheet: gridlines, the cells
// that carry content or style, the selection overlays, sticky headers and
// scrollbars. Cost follows the viewport, not the sheet size.
func (a *app) gridView() *ui.Element {
	cw, ch := a.canvasSize()
	cols := a.layoutCols(cw)
	rows := a.layoutRows(ch)
	vw, vh := a.gridViewportSize()
	hx, hy := a.rowHdrW(), a.colHdrH()

	canvas := ui.Div().ID("xl-gcanvas").Class("xl-canvas").
		CSSRaw("left", pxStr(hx)).CSSRaw("top", pxStr(hy))
	if !a.x.hideGrid {
		canvas.Children(a.gridLines(cols, rows, cw, ch)...)
	}
	canvas.Children(a.cellElements(cols, rows)...)
	canvas.Children(a.selectionOverlays(cols, rows)...)
	canvas.Children(a.freezeLines(cw, ch)...)

	grid := ui.Div().ID("xl-grid").Class("xl-grid").Children(canvas)
	if !a.x.hideHeads {
		grid.Children(a.headers(cols, rows, vw, vh)...)
	}
	grid.Children(a.scrollbars(vw, vh)...)
	if pop := a.notePopover(cols, rows, vw, vh); pop != nil {
		grid.Children(pop)
	}
	if pop := a.autocompletePopover(cols, rows, vw, vh); pop != nil {
		grid.Children(pop)
	}
	if tip := a.resizeTip(); tip != nil {
		grid.Children(tip)
	}
	return grid
}

func (a *app) gridLines(cols, rows []axisSlot, cw, ch int) []*ui.Element {
	out := make([]*ui.Element, 0, len(cols)+len(rows))
	for _, c := range cols {
		out = append(out, place(
			ui.Div().ID("xl-gl-v-"+itoa(c.idx)).Class("xl-vline"),
			c.pos+c.size-1, 0, 1, ch))
	}
	for _, r := range rows {
		out = append(out, place(
			ui.Div().ID("xl-gl-h-"+itoa(r.idx)).Class("xl-hline"),
			0, r.pos+r.size-1, cw, 1))
	}
	return out
}

func (a *app) freezeLines(cw, ch int) []*ui.Element {
	sh := a.sheet()
	fw, fh := a.frozenExtent()
	var out []*ui.Element
	if sh.FreezeCols() > 0 {
		out = append(out, place(ui.Div().ID("xl-gfz-v").Class("xl-frozen-v"), fw-1, 0, frozenLineWidth, ch))
	}
	if sh.FreezeRows() > 0 {
		out = append(out, place(ui.Div().ID("xl-gfz-h").Class("xl-frozen-h"), 0, fh-1, cw, frozenLineWidth))
	}
	return out
}

// cellElements builds one element per visible cell that has content or a
// style. A merged region is one element at its anchor and its covered cells
// are skipped. Empty default cells paint as gridline background only, except
// in the web preview where a click carries no pointer position, so each empty
// cell gets a transparent hit target.
func (a *app) cellElements(cols, rows []axisSlot) []*ui.Element {
	sh := a.sheet()
	out := make([]*ui.Element, 0, len(cols)*len(rows)/denseCellDivisor)
	for _, r := range rows {
		for _, c := range cols {
			if m, ok := sh.MergeAt(r.idx, c.idx); ok {
				fr, _ := firstIn(rows, m.Min.Row, m.Max.Row)
				fc, _ := firstIn(cols, m.Min.Col, m.Max.Col)
				if r.idx == fr && c.idx == fc {
					out = append(out, a.mergedElement(sh, m, cols, rows))
				}
				continue
			}
			raw := sh.Raw(r.idx, c.idx)
			st := sh.StyleAt(r.idx, c.idx)
			note := sh.Comment(r.idx, c.idx) != ""
			if raw == "" && st.BG == "" && !st.HasBorder() && !note {
				if a.web {
					out = append(out, place(ui.Div().ID("xl-ge-"+itoa(r.idx)+"-"+itoa(c.idx)).CSSRaw("position", "absolute"),
						c.pos, r.pos, c.size-1, r.size-1))
				}
				continue
			}
			out = append(out, a.cellElement(sh, CellRef{Row: r.idx, Col: c.idx}, c.pos, r.pos, c.size, r.size, raw, st, note))
		}
	}
	return out
}

// firstIn returns the first visible slot index inside [lo, hi].
func firstIn(slots []axisSlot, lo, hi int) (int, bool) {
	for _, s := range slots {
		if s.idx >= lo && s.idx <= hi {
			return s.idx, true
		}
	}
	return 0, false
}

// mergedElement draws a merged range as one cell spanning the visible part of
// its width and height. It carries the anchor's id, so a click anywhere in
// the merge selects the anchor.
func (a *app) mergedElement(sh *Sheet, m Range, cols, rows []axisSlot) *ui.Element {
	x, w, _ := spanOf(cols, m.Min.Col, m.Max.Col)
	y, h, _ := spanOf(rows, m.Min.Row, m.Max.Row)
	raw := sh.Raw(m.Min.Row, m.Min.Col)
	st := sh.StyleAt(m.Min.Row, m.Min.Col)
	note := sh.Comment(m.Min.Row, m.Min.Col) != ""
	el := a.cellElement(sh, m.Min, x, y, w, h, raw, st, note)
	if st.BG == "" {
		// Hide the gridlines that run through the merged interior.
		el.CSSRaw("background-color", "var(--xl-sheet-bg)")
	}
	return el.Attr("data-merged", m.String())
}

func (a *app) cellElement(sh *Sheet, ref CellRef, x, y, w, h int, raw string, st Style, note bool) *ui.Element {
	val := sh.Value(ref.Row, ref.Col)
	text := sh.Display(ref.Row, ref.Col)
	if a.x.showFormulas && strings.HasPrefix(raw, "=") {
		text = raw
	}
	classes := []string{"xl-cell"}
	switch st.Align {
	case "left":
		classes = append(classes, "xl-cell--left")
	case "center":
		classes = append(classes, "xl-cell--center")
	case "right":
		classes = append(classes, "xl-cell--right")
	default:
		switch val.Kind {
		case KindNumber:
			classes = append(classes, "xl-cell--num")
		case KindError:
			classes = append(classes, "xl-cell--err")
		case KindBool:
			classes = append(classes, "xl-cell--center")
		default:
			classes = append(classes, "xl-cell--text")
		}
	}
	if st.Bold {
		classes = append(classes, "xl-b")
	}
	if st.Italic {
		classes = append(classes, "xl-i")
	}
	switch {
	case st.Underline && st.Strike:
		classes = append(classes, "xl-us")
	case st.Underline:
		classes = append(classes, "xl-u")
	case st.Strike:
		classes = append(classes, "xl-s")
	}
	if st.Wrap {
		classes = append(classes, "xl-cell--wrap")
	}
	el := ui.Div().ID("xl-gc-" + itoa(ref.Row) + "-" + itoa(ref.Col)).Class(strings.Join(classes, " "))
	place(el, x, y, w-1, h-1)
	z := a.zoomPct()
	if z != zoomBase {
		el.CSSRaw("font-size", pxStr(a.zs(cellFontPx)))
		if !st.Wrap {
			el.CSSRaw("line-height", pxStr(h-1))
		}
	}
	el.Attr("role", "gridcell").Attr("aria-label", ref.A1()+": "+text)
	switch {
	case st.FG != "":
		el.CSSRaw("color", st.FG)
	case st.BG != "":
		// A fill the user chose keeps readable text in both themes.
		el.CSSRaw("color", contrastText(st.BG))
	case val.Kind == KindNumber && val.Num < 0 && (st.NumFmt == "Currency" || st.NumFmt == "Accounting"):
		el.CSSRaw("color", "var(--xl-negative-num)")
	}
	if st.BG != "" {
		el.CSSRaw("background-color", st.BG)
	}
	if st.FontSize > 0 {
		el.CSSRaw("font-size", pxStr(a.zs(st.FontSize)))
	}
	applyStyleExtras(el, st, w-1, h-1, text)
	if frac, color, ok := sh.DataBar(ref.Row, ref.Col); ok && frac > 0 {
		bar := ui.Div().Class("xl-bar").CSSRaw("width", pxStr(int(frac*float64(w-1))))
		if color != "" {
			bar.CSSRaw("background-color", color)
		}
		el.Children(bar)
	}
	el.Children(ui.Span().Text(text))
	if note {
		// The renderer does not draw border triangles, so the corner is a small
		// filled square.
		el.Children(ui.Div().Class("xl-note").CSSRaw("border-top", "0").CSSRaw("border-left", "0").
			CSSRaw("width", "6px").CSSRaw("height", "6px").CSSRaw("background-color", "var(--xl-comment-corner)"))
	}
	return el
}

// rectOf returns the pixel rect of a range, and false when none of it is on
// screen.
func rectOf(cols, rows []axisSlot, rg Range) (x, y, w, h int, ok bool) {
	x, w, okc := spanOf(cols, rg.Min.Col, rg.Max.Col)
	y, h, okr := spanOf(rows, rg.Min.Row, rg.Max.Row)
	return x, y, w, h, okc && okr
}

func (a *app) selectionOverlays(cols, rows []axisSlot) []*ui.Element {
	var out []*ui.Element
	out = append(out, a.refHighlights(cols, rows)...)
	if a.sel.Min != a.sel.Max {
		if x, y, w, h, ok := rectOf(cols, rows, a.sel); ok {
			out = append(out, place(ui.Div().ID("xl-gov-sel").Class("xl-sel"), x, y, w, h))
		}
	}
	if a.cutRange != nil {
		if x, y, w, h, ok := rectOf(cols, rows, *a.cutRange); ok {
			out = append(out, place(ui.Div().ID("xl-gov-cut").Class("xl-cut"), x, y, w, h))
		}
	} else if a.copyRange != nil {
		if x, y, w, h, ok := rectOf(cols, rows, *a.copyRange); ok {
			out = append(out, place(ui.Div().ID("xl-gov-cut").Class("xl-cut"), x, y, w, h))
		}
	}
	act := Range{Min: a.active, Max: a.active}
	if m, ok := a.sheet().MergeAt(a.active.Row, a.active.Col); ok {
		act = m
	}
	ax, ay, aw, ah, aok := rectOf(cols, rows, act)
	if aok && !a.editing {
		bx, by := max(ax-1, 0), max(ay-1, 0)
		out = append(out, place(ui.Div().ID("xl-gov-active").Class("xl-active"), bx, by, ax+aw-bx, ay+ah-by))
		if list := a.sheet().ValidationList(a.active.Row, a.active.Col); len(list) > 0 {
			out = append(out, place(
				ui.Div().ID("xl-gov-dd").Class("xl-dd").Text("v"),
				ax+aw-govDropdownInset, ay+dropdownTop, govDropdownWidth, ah-dropdownShrink))
		}
	}
	if x, y, w, h, ok := rectOf(cols, rows, a.sel); ok && !a.editing {
		out = append(out, place(ui.Div().ID("xl-gov-handle").Class("xl-handle"), x+w-fillHandleInset, y+h-fillHandleInset, fillHandleSize, fillHandleSize))
	}
	if a.editing && aok {
		out = append(out, a.editorElement(ax, ay, aw, ah))
	}
	return out
}

// editorElement is the in-cell editor. A real input takes the typing when
// the edit started in the cell; an edit started in the formula bar shows
// the same text in a static box.
func (a *app) editorElement(x, y, w, h int) *ui.Element {
	cw, _ := a.canvasSize()
	width := max(w, len(a.editBuf)*editorCharWidth+editorPadding)
	width = min(width, max(cw-x, w))
	if a.editFromBar {
		return place(ui.Div().ID("xl-geditor-fake").Class("xl-editor-fake").Text(a.editBuf), x-1, y-1, width+1, h+1)
	}
	in := ui.Input().ID("xl-editor").Attr("type", "text").Attr("value", a.editBuf).
		Attr("tabindex", "-1").Attr("aria-label", "Cell editor").Class("xl-editor")
	return place(in, x-1, y-1, width+1, h+1)
}

// headers draws the sticky column letters, row numbers and select-all corner.
func (a *app) headers(cols, rows []axisSlot, vw, vh int) []*ui.Element {
	hx, hy := a.rowHdrW(), a.colHdrH()
	colhdrs := ui.Div().ID("xl-gcolhdrs").Class("xl-colhdrs").
		CSSRaw("left", pxStr(hx)).CSSRaw("height", pxStr(hy))
	chevrons := a.filterChevronCols()
	for _, c := range cols {
		cls := "xl-colh"
		if c.idx >= a.sel.Min.Col && c.idx <= a.sel.Max.Col {
			cls += " xl-hdr--on"
		}
		colhdrs.Children(place(
			ui.Div().ID("xl-gch-"+itoa(c.idx)).Class(cls).Text(ColName(c.idx)),
			c.pos, 0, c.size, hy))
		if chevrons[c.idx] && c.size >= 32 {
			colhdrs.Children(filterChevron(c.idx, c.pos+c.size-chevronInset, chevronTop, a.sheet().FilterAllowed(c.idx) != nil))
		}
	}
	rowhdrs := ui.Div().ID("xl-growhdrs").Class("xl-rowhdrs").
		CSSRaw("top", pxStr(hy)).CSSRaw("width", pxStr(hx))
	for _, r := range rows {
		cls := "xl-rowh"
		if r.idx >= a.sel.Min.Row && r.idx <= a.sel.Max.Row {
			cls += " xl-hdr--on"
		}
		hdr := place(
			ui.Div().ID("xl-grh-"+itoa(r.idx)).Class(cls).Text(strconv.Itoa(r.idx+1)),
			0, r.pos, hx, r.size)
		hdr.CSSRaw("line-height", pxStr(r.size-1))
		rowhdrs.Children(hdr)
	}
	corner := ui.Div().ID("xl-gcorner").Class("xl-corner").Children(ui.Div().Class("xl-corner-tri"))
	return []*ui.Element{colhdrs, rowhdrs, corner}
}

// scrollbars draws both tracks with proportional thumbs.
func (a *app) scrollbars(vw, vh int) []*ui.Element {
	vt, ht := a.vTrack(vw, vh), a.hTrack(vw, vh)
	mr, mc := a.maxScroll()
	fr, fc := a.sheet().FreezeRows(), a.sheet().FreezeCols()
	_, trackH := vt[2], vt[3]
	voff, vlen := thumbGeom(trackH, a.scrollRow-fr, mr-fr, MaxRows-fr, a.visibleRowCount())
	cw, _ := a.canvasSize()
	fw, _ := a.frozenExtent()
	hvis := max((cw-fw)/a.zs(defaultColWidth), 1)
	hoff, hlen := thumbGeom(ht[2], a.scrollCol-fc, mc-fc, MaxCols-fc, hvis)
	v := place(ui.Div().ID("xl-gvt").Class("xl-vtrack"), vt[0], vt[1], scrollbarSize, vt[3])
	v.Children(place(ui.Div().ID("xl-gvth").Class("xl-thumb"), thumbInset, voff, scrollbarSize-thumbShrink, vlen))
	h := place(ui.Div().ID("xl-ght").Class("xl-htrack"), ht[0], ht[1], ht[2], scrollbarSize)
	h.Children(place(ui.Div().ID("xl-ghth").Class("xl-thumb"), hoff, thumbInset, hlen, scrollbarSize-thumbShrink))
	return []*ui.Element{v, h}
}

// vTrack returns {x, y, w, h} of the vertical track in grid-viewport pixels.
func (a *app) vTrack(vw, vh int) [4]int {
	top := a.colHdrH()
	return [4]int{vw - scrollbarSize, top, scrollbarSize, max(vh-top-scrollbarSize, 0)}
}

// hTrack returns {x, y, length, h} of the horizontal track.
func (a *app) hTrack(vw, vh int) [4]int {
	left := a.rowHdrW()
	return [4]int{left, vh - scrollbarSize, max(vw-left-scrollbarSize, 0), scrollbarSize}
}

func (a *app) resizeTip() *ui.Element {
	if a.drag != dragResizeCol && a.drag != dragResizeRow {
		return nil
	}
	label := ""
	if a.drag == dragResizeCol {
		label = "Width: " + itoa(a.sheet().ColWidth(a.dragIndex)) + " px"
	} else {
		label = "Height: " + itoa(a.sheet().RowHeight(a.dragIndex)) + " px"
	}
	return place(ui.Div().ID("xl-gtip").Class("xl-tip").Text(label), rowHeaderWidth+tipOffset, colHeaderHeight+tipOffset, tipWidth, tipHeight)
}

// contrastText picks dark or light text for a #rrggbb fill. Other color
// notations keep the theme text color.
func contrastText(bg string) string {
	if len(bg) != 7 || bg[0] != '#' {
		return "var(--xl-cell-text)"
	}
	v, err := strconv.ParseUint(bg[1:], 16, 32)
	if err != nil {
		return "var(--xl-cell-text)"
	}
	r, g, b := float64(v>>16&colorByteMask), float64(v>>8&colorByteMask), float64(v&colorByteMask)
	if 0.299*r+0.587*g+0.114*b > lumaThreshold {
		return "#1f2328"
	}
	return "#f4f4f5"
}

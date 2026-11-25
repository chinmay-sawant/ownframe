package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

// argSecond is the index of an action's second argument: "xl:verb:first:second".
const argSecond = 2

// argThird is the index of an action's third argument.
const argThird = 3

// Limits and defaults used by the format and view actions.
const (
	maxDecimals     = 8    // most decimal places a number format keeps
	defaultDecimals = 2    // decimals implied by the currency-like formats
	minFontSize     = 6    // smallest font size in points
	maxFontSize     = 72   // largest font size in points
	zoomFitBase     = 1000 // zoom-to-selection numerator, in percent-columns
	zoomMinPct      = 50   // lowest zoom percentage
	zoomMaxPct      = 200  // highest zoom percentage
	autoFitMinW     = 40   // narrowest auto-fit column width, in pixels
	autoFitCharPx   = 8    // estimated glyph width for auto-fit, in pixels
	autoFitPadPx    = 16   // padding added by auto-fit, in pixels
)

// runAction executes one "xl:" action and redraws. It reports whether the
// action was recognized.
func (a *app) runAction(action string) bool {
	return a.runActionCtx(a.bg(), action)
}

// redrawCtx is redraw for callers that hold a handler context.
func (a *app) redrawCtx(ctx context.Context) {
	if a.view == nil {
		return
	}
	if err := a.view.Update(ctx, a.screen()); err != nil {
		a.lastErr = err
		return
	}
	a.syncInputs()
}

// runActionCtx is runAction with the handler context threaded to the redraw.
func (a *app) runActionCtx(ctx context.Context, action string) bool {
	if !strings.HasPrefix(action, "xl:") {
		return false
	}
	p := strings.Split(strings.TrimPrefix(action, "xl:"), ":")
	arg := func(i int) string {
		if i < len(p) {
			return p[i]
		}
		return ""
	}
	if p[0] != "menu" && p[0] != "opt" && p[0] != "flag" && p[0] != "fc" {
		a.x.menu = ""
	}
	handled := true
	switch p[0] {
	case "tab":
		a.ribbonTab = arg(1)
		a.ribbonCollapsed = false
	case "ribbon":
		a.ribbonCollapsed = !a.ribbonCollapsed
	case "menu":
		if a.x.menu == arg(1) {
			a.x.menu = ""
		} else {
			a.x.menu = arg(1)
		}
	case "close":
		a.overlay = ""
		a.x.menu = ""
		a.redrawCtx(ctx)
		return true
	case "toast":
		a.x.toastOpen = false
	case "undo":
		a.undo()
		return true
	case "redo":
		a.redo()
		return true
	case "cut":
		a.copy(true)
	case "copy":
		a.copy(false)
	case "paste":
		switch {
		case arg(1) == "special":
			a.pasteSpecial()
		case !a.pasteCut():
			a.paste()
		}
	case "painter":
		a.armPainter()
	case "bold", "italic", "underline", "strike", "wrap":
		a.toggleStyle(p[0])
	case "fsize":
		a.changeFontSize(arg(1))
	case "align":
		v := arg(1)
		a.applyFormat(func(s *Style) { s.Align = v })
	case "valign":
		v := arg(1)
		a.applyFormat(func(s *Style) { s.VAlign = v })
	case "indent":
		d := indentDelta(arg(1))
		a.applyFormat(func(s *Style) { s.Indent = clampIndent(s.Indent + d) })
	case "family":
		v := arg(1)
		a.applyFormat(func(s *Style) { s.FontFamily = v })
	case "nf":
		v := arg(1)
		a.applyFormat(func(s *Style) { s.NumFmt = v })
	case "cur":
		a.applyFormat(func(s *Style) {
			s.NumFmt = "Accounting"
			if s.Decimals == 0 {
				s.Decimals = 2
			}
		})
	case "pct":
		a.applyFormat(func(s *Style) { s.NumFmt = "Percent" })
	case "comma":
		a.applyFormat(func(s *Style) { s.NumFmt = "Number"; s.Decimals = 2 })
	case "dec":
		d := 1
		if arg(1) == "-" {
			d = -1
		}
		st := a.activeStyle()
		next := effectiveDecimals(st) + d
		if next < 0 {
			next = 0
		}
		if next > maxDecimals {
			next = maxDecimals
		}
		stored := next
		if next == 0 {
			stored = -1 // the engine reads 0 as "format default" and -1 as zero places
		}
		a.applyFormat(func(s *Style) { s.Decimals = stored })
	case "fillc":
		a.rememberColor(arg(1))
		bg := arg(1)
		a.applyFormat(func(s *Style) { s.BG = bg })
	case "fontc":
		a.rememberColor(arg(1))
		fg := arg(1)
		a.applyFormat(func(s *Style) { s.FG = fg })
	case "border":
		a.wb.ApplyBorder(a.wb.Sheet(), a.sel, arg(1))
	case "autosum":
		a.autoSum()
	case "fn":
		a.beginEdit("=" + arg(1) + "(")
		return true
	case "clear":
		a.clearWhat(arg(1))
	case "cstyle":
		a.applyCellStyle(arg(1))
	case "cells":
		a.cellsAction(arg(1))
	case "merge":
		a.mergeAction(arg(1))
	case "sort":
		a.sortAction(arg(1))
	case "dlg":
		a.openDialogByID(arg(1))
		return true
	case "filter":
		a.filterAction(arg(1), arg(argSecond))
		if arg(1) == "open" || arg(1) == "val" {
			return true // openFilter and the checkbox flip already redraw
		}
	case "cf":
		a.condAction(arg(1))
	case "cfr":
		a.condRuleAction(arg(1), arg(argSecond))
	case "chart":
		a.chartAction(arg(1), arg(argSecond), arg(argThird))
	case "find":
		a.findAction(arg(1), arg(argSecond))
	case "fc":
		a.formatDraftAction(p[1:])
	case "val":
		a.validationAction(arg(1))
	case "comment":
		a.commentAction(arg(1))
	case "file":
		a.fileAction(arg(1))
	case "view":
		a.viewAction(arg(1))
	case "zoom":
		a.zoomAction(arg(1))
	case "freeze":
		a.freezeAction(arg(1))
	case "theme":
		a.toggleTheme()
		return true
	case "opt":
		a.setOpt(arg(1), arg(argSecond))
	case "flag":
		a.setOpt(arg(1), strconv.FormatBool(!a.flag(arg(1), a.flagDefault(arg(1)))))
	case "ctx":
		a.ctxAction(arg(1))
	case "vlist":
		a.pickListValue(arg(1))
	default:
		handled = false
	}
	if handled {
		a.redrawCtx(ctx)
	}
	return handled
}

// effectiveDecimals mirrors how the engine reads Style.Decimals: negative means
// zero places and zero means the number format's own default.
func effectiveDecimals(st Style) int {
	switch {
	case st.Decimals < 0:
		return 0
	case st.Decimals > 0:
		return st.Decimals
	}
	switch st.NumFmt {
	case "Number", "Currency", "Accounting", "Scientific":
		return defaultDecimals
	}
	return 0
}

// flagDefault returns the default of flags whose default is true.
func (a *app) flagDefault(key string) bool { return key == "sorthdr" }

func (a *app) toggleStyle(which string) {
	st := a.activeStyle()
	switch which {
	case "bold":
		v := !st.Bold
		a.applyFormat(func(s *Style) { s.Bold = v })
	case "italic":
		v := !st.Italic
		a.applyFormat(func(s *Style) { s.Italic = v })
	case "underline":
		v := !st.Underline
		a.applyFormat(func(s *Style) { s.Underline = v })
	case "strike":
		v := !st.Strike
		a.applyFormat(func(s *Style) { s.Strike = v })
	case "wrap":
		v := !st.Wrap
		a.applyFormat(func(s *Style) { s.Wrap = v })
	}
}

func (a *app) changeFontSize(arg string) {
	size := a.activeStyle().FontSize
	if size == 0 {
		size = defaultFontSize
	}
	switch arg {
	case "+":
		size++
	case "-":
		size--
	default:
		if n, err := strconv.Atoi(arg); err == nil {
			size = n
		}
	}
	if size < minFontSize {
		size = minFontSize
	}
	if size > maxFontSize {
		size = maxFontSize
	}
	a.applyFormat(func(s *Style) { s.FontSize = size })
}

// pasteCut finishes a pending cut by moving the block to the active cell, so
// formulas that pointed at the block follow it. It reports false when no cut
// is pending.
func (a *app) pasteCut() bool {
	if a.cutRange == nil {
		return false
	}
	sh := a.wb.Sheet()
	buf := sh.CopyRich(*a.cutRange)
	buf.Cut = true
	rng := a.wb.PasteSpecial(sh, a.active, buf, PasteAll, false, false)
	a.cutRange, a.copyRange = nil, nil
	a.selectRange(rng)
	return true
}

func (a *app) armPainter() {
	if a.x.painter {
		// A second click while armed makes it sticky, as a double-click does.
		a.x.painterSticky = true
		return
	}
	a.x.painter = true
	a.x.painterSticky = false
	a.x.painterStyle = a.activeStyle()
	a.x.painterSrc = a.sel
}

func (a *app) clearWhat(what string) {
	sh := a.wb.Sheet()
	switch what {
	case "formats":
		a.applyFormat(func(s *Style) { *s = Style{} })
	case "contents":
		a.wb.ClearRange(sh, a.sel)
	case "comments":
		a.eachCell(func(c CellRef) { sh.SetComment(c.Row, c.Col, "") })
	default:
		a.wb.ClearRange(sh, a.sel)
		a.applyFormat(func(s *Style) { *s = Style{} })
		a.eachCell(func(c CellRef) { sh.SetComment(c.Row, c.Col, "") })
	}
}

func (a *app) eachCell(f func(CellRef)) {
	for r := a.sel.Min.Row; r <= a.sel.Max.Row; r++ {
		for c := a.sel.Min.Col; c <= a.sel.Max.Col; c++ {
			f(CellRef{Row: r, Col: c})
		}
	}
}

func (a *app) applyCellStyle(name string) {
	for _, cs := range cellStyles() {
		if cs.Name != name {
			continue
		}
		a.applyFormat(func(s *Style) {
			s.FG, s.BG, s.Bold, s.FontSize = cs.FG, cs.BG, cs.Bold, cs.Size
			if name == "Total" {
				s.Edges.Top = EdgeThin
			}
		})
		return
	}
}

// autoSum writes =SUM over the contiguous numbers above (or left of) the
// active cell, or under each column of a multi-row selection.
func (a *app) autoSum() {
	sh := a.wb.Sheet()
	if a.sel.Min != a.sel.Max && a.sel.Max.Row > a.sel.Min.Row {
		target := a.sel.Max.Row + 1
		for c := a.sel.Min.Col; c <= a.sel.Max.Col; c++ {
			rng := Range{Min: CellRef{a.sel.Min.Row, c}, Max: CellRef{a.sel.Max.Row, c}}
			a.wb.Edit(sh, target, c, "=SUM("+rng.String()+")")
		}
		return
	}
	at := a.active
	isNum := func(r, c int) bool { return r >= 0 && c >= 0 && sh.Value(r, c).Kind == KindNumber }
	top := at.Row
	for isNum(top-1, at.Col) {
		top--
	}
	if top < at.Row {
		rng := Range{Min: CellRef{top, at.Col}, Max: CellRef{at.Row - 1, at.Col}}
		a.wb.Edit(sh, at.Row, at.Col, "=SUM("+rng.String()+")")
		return
	}
	left := at.Col
	for isNum(at.Row, left-1) {
		left--
	}
	if left < at.Col {
		rng := Range{Min: CellRef{at.Row, left}, Max: CellRef{at.Row, at.Col - 1}}
		a.wb.Edit(sh, at.Row, at.Col, "=SUM("+rng.String()+")")
		return
	}
	a.beginEdit("=SUM(")
}

// sortRange is the selection when it spans several cells, else the used range.
func (a *app) sortRange() Range {
	if a.sel.Min != a.sel.Max {
		return a.sel
	}
	return a.wb.Sheet().UsedRange()
}

func (a *app) sortAction(dir string) {
	asc := dir != "desc"
	a.wb.SortRange(a.wb.Sheet(), a.sortRange(), a.active.Col, asc, true)
}

func (a *app) openDialogByID(id string) {
	switch id {
	case "find":
		a.x.findMsg, a.x.findHits = "", nil
	case "format":
		a.x.fmtDraft = a.activeStyle().migrated()
		a.x.fmtOrig = a.x.fmtDraft
		a.setOpt("fcborder", "")
		a.x.fmtTab = "Number"
	case "sort":
		a.setOpt("sk1", ColName(a.active.Col))
	case "condfmt":
	case "validation":
	case "comment":
	}
	a.x.menu = ""
	a.setOverlay(id)
}

// formatDraftAction edits the Format Cells draft or commits it.
func (a *app) formatDraftAction(p []string) {
	d := &a.x.fmtDraft
	arg := func(i int) string {
		if i < len(p) {
			return p[i]
		}
		return ""
	}
	switch arg(0) {
	case "tab":
		a.x.fmtTab = arg(1)
	case "align":
		d.Align = arg(1)
	case "wrap":
		d.Wrap = !d.Wrap
	case "bold":
		d.Bold = !d.Bold
	case "italic":
		d.Italic = !d.Italic
	case "underline":
		d.Underline = !d.Underline
	case "strike":
		d.Strike = !d.Strike
	case "size":
		d.FontSize, _ = strconv.Atoi(arg(1))
	case "font":
		d.FG = arg(1)
	case "fill":
		d.BG = arg(1)
	case "border":
		a.setOpt("fcborder", arg(1))
	case "family":
		d.FontFamily = arg(1)
	case "valign":
		d.VAlign = arg(1)
	case "indent":
		d.Indent = clampIndent(d.Indent + indentDelta(arg(1)))
	case "nf":
		d.NumFmt = arg(1)
	case "dec":
		if arg(1) == "-" && d.Decimals > 0 {
			d.Decimals--
		} else if arg(1) == "+" && d.Decimals < 8 {
			d.Decimals++
		}
	case "ok":
		orig, draft := a.x.fmtOrig, a.x.fmtDraft
		a.applyFormat(func(s *Style) { diffApply(orig, draft, s) })
		if preset := a.opt("fcborder", ""); preset != "" {
			a.wb.ApplyBorder(a.wb.Sheet(), a.sel, preset)
		}
		a.overlay = ""
	}
}

// diffApply copies only the fields that changed between orig and draft, so a
// cell's own style survives for fields the dialog left alone.
func diffApply(orig, draft Style, s *Style) {
	if orig.Bold != draft.Bold {
		s.Bold = draft.Bold
	}
	if orig.Italic != draft.Italic {
		s.Italic = draft.Italic
	}
	if orig.Underline != draft.Underline {
		s.Underline = draft.Underline
	}
	if orig.Strike != draft.Strike {
		s.Strike = draft.Strike
	}
	if orig.Wrap != draft.Wrap {
		s.Wrap = draft.Wrap
	}
	if orig.Align != draft.Align {
		s.Align = draft.Align
	}
	if orig.FG != draft.FG {
		s.FG = draft.FG
	}
	if orig.BG != draft.BG {
		s.BG = draft.BG
	}
	if orig.NumFmt != draft.NumFmt {
		s.NumFmt = draft.NumFmt
	}
	if orig.Decimals != draft.Decimals {
		s.Decimals = draft.Decimals
	}
	if orig.FontSize != draft.FontSize {
		s.FontSize = draft.FontSize
	}
	if orig.FontFamily != draft.FontFamily {
		s.FontFamily = draft.FontFamily
	}
	if orig.VAlign != draft.VAlign {
		s.VAlign = draft.VAlign
	}
	if orig.Indent != draft.Indent {
		s.Indent = draft.Indent
	}
}

const maxIndent = 15

func clampIndent(n int) int { return max(0, min(n, maxIndent)) }

func indentDelta(sign string) int {
	if sign == "-" {
		return -1
	}
	return 1
}

// --- conditional formatting ---

func (a *app) condAction(kind string) {
	switch kind {
	case "gt", "lt", "between", "eq", "text":
		a.setOpt("cfkind", kind)
		a.openDialogByID("condfmt")
	case "dupes":
		a.addCond(CondRule{Kind: "dupes", Preset: "red"})
	case "top":
		a.addCond(CondRule{Kind: "top", A: "10", Preset: "green"})
	case "bottom":
		a.addCond(CondRule{Kind: "bottom", A: "10", Preset: "yellow"})
	case "above", "below":
		a.addCond(CondRule{Kind: kind, Preset: "green"})
	case "databar":
		a.addCond(CondRule{Kind: "databar", Preset: "blue"})
	case "scale3":
		a.addCond(CondRule{Kind: "scale3"})
	case "clear":
		a.wb.Sheet().ClearCondFormat(a.sel)
	case "apply":
		rule := CondRule{
			Kind: a.opt("cfkind", "gt"), A: a.input("rb-cf-a"), B: a.input("rb-cf-b"),
			Preset: a.opt("cfpreset", "red"),
		}
		a.addCond(rule)
	}
}

func (a *app) addCond(rule CondRule) {
	a.wb.Sheet().AddCondFormat(a.sel, rule)
}

func (a *app) condRuleAction(verb, idx string) {
	i, err := strconv.Atoi(idx)
	sh := a.wb.Sheet()
	if err != nil || i < 0 || i >= len(sh.CondRules()) {
		return
	}
	switch verb {
	case "del":
		sh.DeleteCond(i)
	case "toggle":
		sh.SetCondOff(i, !sh.CondRules()[i].Off)
	case "up":
		sh.MoveCondUp(i)
	}
}

// --- filter ---

func (a *app) filterAction(verb, arg string) {
	sh := a.wb.Sheet()
	switch verb {
	case "toggle":
		a.x.filterOn = !a.x.filterOn
		sh.SetFilterHeader(sh.UsedRange().Min.Row)
		if !a.x.filterOn {
			for c := 0; c < MaxCols; c++ {
				sh.SetFilter(c, nil)
			}
		}
	case "clear":
		for c := 0; c < MaxCols; c++ {
			sh.SetFilter(c, nil)
		}
	case "open":
		col, _ := strconv.Atoi(arg)
		a.openFilter(col)
	case "val":
		i, _ := strconv.Atoi(arg)
		if i >= 0 && i < len(a.x.filterVals) {
			v := a.x.filterVals[i]
			a.x.filterAllowed[v] = !a.x.filterAllowed[v]
		}
		a.redraw()
	case "selall":
		for _, v := range a.x.filterVals {
			a.x.filterAllowed[v] = filterMatches(v, a.opt("fsearch", ""))
		}
	case "sort":
		a.filterSort(arg != "desc")
		a.overlay = ""
	case "clearcol":
		sh.SetFilter(a.x.filterCol, nil)
		a.overlay = ""
	case "apply":
		all := true
		for _, v := range a.x.filterVals {
			if !a.x.filterAllowed[v] {
				all = false
			}
		}
		if all {
			sh.SetFilter(a.x.filterCol, nil)
		} else {
			m := map[string]bool{}
			for v, on := range a.x.filterAllowed {
				if on {
					m[v] = true
				}
			}
			sh.SetFilter(a.x.filterCol, m)
		}
		a.x.filterOn = true
		a.overlay = ""
	}
}

// filterMatches reports whether a value passes the popover search text.
func filterMatches(v, query string) bool {
	return query == "" || strings.Contains(strings.ToLower(v), strings.ToLower(query))
}

// filterSearch applies the popover search: only values matching the text stay
// checked, as in Excel.
func (a *app) filterSearch(query string) {
	a.setOpt("fsearch", query)
	for _, v := range a.x.filterVals {
		a.x.filterAllowed[v] = filterMatches(v, query)
	}
}

// filterSort sorts the filtered table by the popover's column, keeping the
// header row in place.
func (a *app) filterSort(asc bool) {
	sh := a.wb.Sheet()
	r := sh.UsedRange()
	r.Min.Row = max(r.Min.Row, sh.FilterHeader())
	a.wb.SortRange(sh, r, a.x.filterCol, asc, true)
}

// openFilter opens the filter popover for one column. The grid calls it from
// the header chevron.
func (a *app) openFilter(col int) {
	sh := a.wb.Sheet()
	vals := sh.FilterValues(col)
	a.x.filterCol, a.x.filterVals = col, vals
	a.setOpt("fsearch", "")
	a.x.filterAllowed = map[string]bool{}
	allowed := sh.FilterAllowed(col)
	for _, v := range vals {
		a.x.filterAllowed[v] = allowed == nil || allowed[v]
	}
	a.setOverlay("filter")
}

// --- charts ---

func (a *app) chartAction(verb, arg, extra string) {
	switch verb {
	case "new":
		a.addChart(arg)
		a.ribbonTab = "Insert"
	case "select":
		if id, err := strconv.Atoi(arg); err == nil {
			a.x.selChart = id
		}
	case "del":
		if id, err := strconv.Atoi(arg); err == nil {
			a.deleteChart(id)
		}
	}
}

// --- find and replace ---

func (a *app) findOptions() (q string, matchCase, entire, book, values bool) {
	return a.input("rb-find-q"), a.flag("matchcase", false), a.flag("entire", false),
		a.opt("scope", "sheet") == "book", a.opt("lookin", "formulas") == "values"
}

func matchText(text, q string, matchCase, entire bool) bool {
	if !matchCase {
		text, q = strings.ToLower(text), strings.ToLower(q)
	}
	if entire {
		return text == q
	}
	return strings.Contains(text, q)
}

func (a *app) findAll() []findHit {
	q, mc, entire, book, values := a.findOptions()
	if q == "" {
		return nil
	}
	var hits []findHit
	for i, sh := range a.wb.Sheets {
		if !book && i != a.wb.Active {
			continue
		}
		used := sh.UsedRange()
		for r := used.Min.Row; r <= used.Max.Row; r++ {
			for c := used.Min.Col; c <= used.Max.Col; c++ {
				text := sh.Raw(r, c)
				if values {
					text = sh.Display(r, c)
				}
				if text != "" && matchText(text, q, mc, entire) {
					hits = append(hits, findHit{SheetIdx: i, Sheet: sh.Name, Ref: CellRef{r, c}, Text: text})
				}
			}
		}
	}
	return hits
}

func (a *app) gotoHit(h findHit) {
	a.wb.Active = h.SheetIdx
	a.selectCell(h.Ref)
}

func (a *app) findAction(verb, arg string) {
	a.setOpt("findq", a.input("rb-find-q"))
	a.setOpt("findr", a.input("rb-find-r"))
	switch verb {
	case "next", "prev":
		hits := a.findAll()
		a.x.findHits = hits
		if len(hits) == 0 {
			a.x.findMsg = "No matches found."
			return
		}
		idx := -1
		for i, h := range hits {
			if h.SheetIdx == a.wb.Active && h.Ref == a.active {
				idx = i
			}
		}
		if verb == "next" {
			idx = (idx + 1) % len(hits)
		} else {
			idx = (idx - 1 + len(hits)) % len(hits)
			if idx < 0 {
				idx = len(hits) - 1
			}
		}
		a.x.findIdx = idx
		a.x.findMsg = fmt.Sprintf("Match %d of %d.", idx+1, len(hits))
		a.gotoHit(hits[idx])
	case "all":
		a.x.findHits = a.findAll()
		a.x.findMsg = fmt.Sprintf("%d cell(s) found.", len(a.x.findHits))
	case "go":
		if i, err := strconv.Atoi(arg); err == nil && i < len(a.x.findHits) {
			a.gotoHit(a.x.findHits[i])
		}
	case "replace":
		for _, h := range a.findAll() {
			if h.SheetIdx == a.wb.Active && h.Ref == a.active {
				a.replaceAt(h)
				a.x.findMsg = "Replaced 1 cell."
				a.findAction("next", "")
				return
			}
		}
		a.findAction("next", "")
	case "replaceall":
		hits := a.findAll()
		for _, h := range hits {
			a.replaceAt(h)
		}
		a.x.findHits = nil
		a.x.findMsg = fmt.Sprintf("Replaced %d cell(s).", len(hits))
	}
}

func (a *app) replaceAt(h findHit) {
	q, mc, entire, _, values := a.findOptions()
	repl := a.input("rb-find-r")
	sh := a.wb.Sheets[h.SheetIdx]
	raw := sh.Raw(h.Ref.Row, h.Ref.Col)
	if values && strings.HasPrefix(raw, "=") {
		return // a computed cell has no value text to rewrite
	}
	var out string
	if entire {
		out = repl
	} else {
		pat := regexp.QuoteMeta(q)
		if !mc {
			pat = "(?i)" + pat
		}
		out = regexp.MustCompile(pat).ReplaceAllLiteralString(raw, repl)
	}
	a.wb.Edit(sh, h.Ref.Row, h.Ref.Col, out)
}

// --- paste special ---

// pasteSpecial applies the dialog choices. It uses the rich buffer of the last
// copy when there is one (formulas, styles) and the clipboard text otherwise.
func (a *app) pasteSpecial() {
	a.overlay = ""
	sh := a.wb.Sheet()
	mode := PasteAll
	switch a.opt("ps", "all") {
	case "values":
		mode = PasteValues
	case "formulas":
		mode = PasteFormulas
	case "formats":
		mode = PasteFormats
	}
	skip, transpose := a.flag("psskip", false), a.flag("pstranspose", false)
	var buf *ClipBuf
	switch {
	case a.copyRange != nil:
		buf = sh.CopyRich(*a.copyRange)
	case a.cutRange != nil:
		buf = sh.CopyRich(*a.cutRange)
	default:
		text, err := clipboardText()
		if err != nil || text == "" {
			a.notify("Paste Special", "The clipboard is empty.")
			return
		}
		buf = &ClipBuf{Sheet: sh}
		for _, row := range ParseTSV(text) {
			line := make([]ClipCell, len(row))
			for i, f := range row {
				line[i] = ClipCell{Raw: f, Val: Value{Kind: KindString, Str: f}}
				if n, err := strconv.ParseFloat(f, 64); err == nil {
					line[i].Val = Value{Kind: KindNumber, Num: n}
				}
			}
			buf.Cells = append(buf.Cells, line)
		}
		if len(buf.Cells) > 0 {
			buf.Src = Range{Max: CellRef{Row: len(buf.Cells) - 1, Col: len(buf.Cells[0]) - 1}}
		}
	}
	rng := a.wb.PasteSpecial(sh, a.active, buf, mode, skip, transpose)
	a.copyRange = nil
	a.selectRange(rng)
}

// --- validation ---

func (a *app) validationAction(verb string) {
	sh := a.wb.Sheet()
	switch verb {
	case "clear":
		sh.SetValidation(a.sel, nil)
	case "ok":
		kind := a.opt("vtype", "list")
		if kind == "list" {
			var list []string
			for _, s := range strings.Split(a.input("rb-val-list"), ",") {
				if s = strings.TrimSpace(s); s != "" {
					list = append(list, s)
				}
			}
			a.setOpt("vlist", strings.Join(list, ","))
			sh.SetValidation(a.sel, list)
		} else {
			lo, _ := strconv.ParseFloat(strings.TrimSpace(a.input("rb-val-min")), 64)
			hi, _ := strconv.ParseFloat(strings.TrimSpace(a.input("rb-val-max")), 64)
			sh.SetValidationBounds(a.sel, kind, lo, hi)
		}
		a.overlay = ""
	}
}

// validateEntry checks a value about to be committed to row, col. The edit
// commit calls it; ok is false when the value breaks a rule, with msg set.
func (a *app) validateEntry(row, col int, raw string) (ok bool, msg string) {
	return a.wb.Sheet().CheckValidation(row, col, raw)
}

// --- comments ---

func (a *app) commentCells() []CellRef {
	sh := a.wb.Sheet()
	used := sh.UsedRange()
	var out []CellRef
	for r := used.Min.Row; r <= used.Max.Row; r++ {
		for c := used.Min.Col; c <= used.Max.Col; c++ {
			if sh.Comment(r, c) != "" {
				out = append(out, CellRef{r, c})
			}
		}
	}
	return out
}

func (a *app) commentAction(verb string) {
	sh := a.wb.Sheet()
	switch verb {
	case "ok":
		sh.SetComment(a.active.Row, a.active.Col, a.input("rb-comment-text"))
		a.overlay = ""
	case "delete":
		sh.SetComment(a.active.Row, a.active.Col, "")
		if a.overlay == "comment" {
			a.overlay = ""
		}
	case "next", "prev":
		cells := a.commentCells()
		if len(cells) == 0 {
			a.notify("Comments", "This sheet has no comments.")
			return
		}
		idx := -1
		for i, c := range cells {
			if c == a.active {
				idx = i
			}
		}
		if verb == "next" {
			idx = (idx + 1) % len(cells)
		} else {
			idx = (idx - 1 + len(cells)) % len(cells)
		}
		a.selectCell(cells[idx])
	}
}

// --- files ---

// savedFileMode keeps saved workbooks private to the user.
const savedFileMode = 0o600

func clipboardText() (string, error) { return readClipboard() }

func (a *app) fileAction(verb string) {
	switch verb {
	case "new":
		a.replaceWorkbook(NewWorkbook())
		a.overlay = ""
	case "open":
		path := strings.TrimSpace(a.input("rb-open-path"))
		a.setOpt("openpath", path)
		if err := a.openPath(path); err != nil {
			a.x.msg = err.Error()
			return
		}
		a.overlay = ""
	case "save":
		path := strings.TrimSpace(a.input("rb-save-path"))
		a.setOpt("savepath", path)
		if err := a.savePath(path, a.opt("savefmt", "json")); err != nil {
			a.x.msg = err.Error()
			return
		}
		a.overlay = ""
		a.notify("Saved", path)
	}
}

func (a *app) replaceWorkbook(wb *Workbook) {
	a.wb = wb
	a.x.charts = nil
	a.active = CellRef{}
	a.sel = Range{}
	a.anchor = CellRef{}
	a.scrollRow, a.scrollCol = 0, 0
}

// openPath loads a .json, .xlsx or .csv file from disk. A CSV becomes a new
// sheet.
func (a *app) openPath(path string) error {
	if path == "" {
		return fmt.Errorf("enter a file path")
	}
	if strings.EqualFold(filepath.Ext(path), ".xlsx") {
		wb, err := LoadXLSX(path)
		if err != nil {
			return err
		}
		a.replaceWorkbook(wb)
		return nil
	}
	b, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return err
	}
	return a.openBytes(filepath.Base(path), b)
}

// openBytes loads .json or CSV data under a file name.
func (a *app) openBytes(name string, b []byte) error {
	switch ext := strings.ToLower(filepath.Ext(name)); ext {
	case ".json":
		wb, err := UnmarshalWorkbook(b)
		if err != nil {
			return err
		}
		a.replaceWorkbook(wb)
	case ".csv", ".tsv", ".txt":
		if err := a.wb.ImportCSV(bytes.NewReader(b), strings.TrimSuffix(name, filepath.Ext(name))); err != nil {
			return err
		}
		a.wb.Active = len(a.wb.Sheets) - 1
	default:
		return fmt.Errorf("unsupported file type %q", ext)
	}
	return nil
}

// handleDrop opens dropped .json, .xlsx and .csv files. It is the
// Handlers.Drop hook.
func (a *app) handleDrop(ctx context.Context, files []ownframe.Drop) error {
	for _, f := range files {
		if f.IsDir {
			continue
		}
		var err error
		if strings.EqualFold(filepath.Ext(f.Name), ".xlsx") && f.Path != "" {
			err = a.openPath(f.Path)
		} else {
			var b []byte
			if f.Read == nil {
				continue
			}
			if b, err = f.Read(); err == nil {
				err = a.openBytes(f.Name, b)
			}
		}
		if err != nil {
			a.notify("Open failed", err.Error())
		}
		a.redrawCtx(ctx)
		return nil
	}
	return nil
}

func (a *app) savePath(path, format string) error {
	if path == "" {
		return fmt.Errorf("enter a file path")
	}
	switch format {
	case "xlsx":
		return a.wb.SaveXLSX(path)
	case "csv":
		f, err := os.Create(filepath.Clean(path))
		if err != nil {
			return err
		}
		if err := a.wb.Sheet().ExportCSV(f); err != nil {
			_ = f.Close()
			return err
		}
		return f.Close()
	default:
		b, err := a.wb.MarshalJSON()
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Clean(path), b, savedFileMode)
	}
}

// --- view ---

func (a *app) viewAction(which string) {
	switch which {
	case "grid":
		a.x.hideGrid = !a.x.hideGrid
	case "heads":
		a.x.hideHeads = !a.x.hideHeads
	case "fbar":
		a.x.hideFormulaBar = !a.x.hideFormulaBar
	case "formulas":
		a.x.showFormulas = !a.x.showFormulas
	}
}

func (a *app) zoomAction(which string) {
	z := a.zoomPct()
	switch which {
	case "in":
		z += 10
	case "out":
		z -= 10
	case "reset":
		z = zoomDefaultPct
	case "sel":
		cols := a.sel.Max.Col - a.sel.Min.Col + 1
		z = zoomFitBase / cols
	}
	if z < zoomMinPct {
		z = zoomMinPct
	}
	if z > zoomMaxPct {
		z = zoomMaxPct
	}
	a.x.zoom = z
}

func (a *app) freezeAction(which string) {
	sh := a.wb.Sheet()
	switch which {
	case "both":
		sh.SetFreeze(a.active.Row, a.active.Col)
	case "row":
		sh.SetFreeze(1, sh.FreezeCols())
	case "col":
		sh.SetFreeze(sh.FreezeRows(), 1)
	default:
		sh.SetFreeze(0, 0)
	}
}

// --- context menu ---

func (a *app) ctxAction(verb string) {
	a.overlay = ""
	sh := a.wb.Sheet()
	switch verb {
	case "cut":
		a.copy(true)
	case "copy":
		a.copy(false)
	case "paste":
		a.paste()
	case "clear":
		a.wb.ClearRange(sh, a.sel)
	case "hide", "unhide":
		a.hideSelection(verb == "hide")
	case "insert":
		a.insertDelete(true)
	case "delete":
		a.insertDelete(false)
	case "autofit":
		a.autoFit()
	case "addsheet":
		sh := a.wb.AddSheet()
		_ = sh
		a.wb.Active = len(a.wb.Sheets) - 1
	case "deletesheet":
		if len(a.wb.Sheets) > 1 {
			a.wb.DeleteSheet(a.wb.Active)
			if a.wb.Active >= len(a.wb.Sheets) {
				a.wb.Active = len(a.wb.Sheets) - 1
			}
		}
	}
}

// autoFit sizes the selected column or row to its content.
func (a *app) autoFit() {
	sh := a.wb.Sheet()
	if a.x.ctxKind == "row" {
		return
	}
	for c := a.sel.Min.Col; c <= a.sel.Max.Col; c++ {
		w := autoFitMinW
		for r := 0; r <= sh.UsedRange().Max.Row; r++ {
			if n := len([]rune(sh.Display(r, c)))*autoFitCharPx + autoFitPadPx; n > w {
				w = n
			}
		}
		sh.SetColWidth(c, w)
	}
}

// pickListValue writes the chosen validation-list entry into the active cell.
func (a *app) pickListValue(idx string) {
	i, err := strconv.Atoi(idx)
	items := a.wb.Sheet().ValidationList(a.active.Row, a.active.Col)
	a.overlay = ""
	if err != nil || i < 0 || i >= len(items) {
		return
	}
	a.wb.Edit(a.wb.Sheet(), a.active.Row, a.active.Col, items[i])
}

// hideSelection hides or unhides the selected rows or columns, following the
// kind of the open context menu (rows unless it was a column header menu).
func (a *app) hideSelection(hide bool) {
	sh := a.wb.Sheet()
	if a.x.ctxKind == "col" {
		for c := a.sel.Min.Col; c <= a.sel.Max.Col; c++ {
			sh.SetColHidden(c, hide)
		}
		return
	}
	for r := a.sel.Min.Row; r <= a.sel.Max.Row; r++ {
		sh.SetRowHidden(r, hide)
	}
}

// insertDelete inserts or deletes whole rows or columns under the selection.
// The context menu kind picks columns for a column header and rows otherwise.
func (a *app) insertDelete(insert bool) {
	sh := a.wb.Sheet()
	if a.x.ctxKind == "col" || a.x.cellsDim == "col" {
		n := a.sel.Max.Col - a.sel.Min.Col + 1
		if insert {
			a.wb.InsertCols(sh, a.sel.Min.Col, n)
		} else {
			a.wb.DeleteCols(sh, a.sel.Min.Col, n)
		}
		return
	}
	n := a.sel.Max.Row - a.sel.Min.Row + 1
	if insert {
		a.wb.InsertRows(sh, a.sel.Min.Row, n)
	} else {
		a.wb.DeleteRows(sh, a.sel.Min.Row, n)
	}
}

// mergeAction merges or unmerges the selection. Merging keeps the top-left
// value and warns when other cells hold data, which the merge discards.
func (a *app) mergeAction(verb string) {
	sh := a.wb.Sheet()
	if verb == "unmerge" {
		sh.Unmerge(a.sel)
		return
	}
	if a.sel.Min == a.sel.Max {
		return
	}
	lost := 0
	a.eachCell(func(c CellRef) {
		if c != a.sel.Min && sh.Raw(c.Row, c.Col) != "" {
			lost++
		}
	})
	sh.Merge(a.sel)
	if verb == "center" {
		a.applyFormat(func(s *Style) { s.Align = "center" })
	}
	if lost > 0 {
		a.notify("Merge Cells", "Merging kept the top-left value. Other values stay hidden under the merge.")
	}
}

func (a *app) cellsAction(verb string) {
	a.x.cellsDim = "row"
	if strings.HasSuffix(verb, "col") {
		a.x.cellsDim = "col"
	}
	switch verb {
	case "insrow", "inscol":
		a.insertDelete(true)
	case "delrow", "delcol":
		a.insertDelete(false)
	default:
		a.wb.ClearRange(a.wb.Sheet(), a.sel)
	}
	a.x.cellsDim = ""
}

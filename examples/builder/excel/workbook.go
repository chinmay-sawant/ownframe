package main

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Sheet is one worksheet with a sparse cell store.
type Sheet struct {
	Name     string
	TabColor string

	wb         *Workbook
	cells      map[CellRef]*xlCell
	styles     map[CellRef]Style
	comments   map[CellRef]string
	colW, rowH map[int]int
	hiddenRows map[int]bool
	hiddenCols map[int]bool
	freezeR    int
	freezeC    int
	filters    map[int]map[string]bool
	filterHdr  int
	merges     []Range
	conds      []*xlCondEntry
	valid      []xlValid
	ver        int
	autoVer    int
	auto       map[int]int
	usedVer    int
	used       Range
}

// xlValid is one data-validation rule: a dropdown list or numeric/length
// bounds ("list", "whole", "decimal", "length").
type xlValid struct {
	rng      Range
	kind     string
	list     []string
	min, max float64
}

// Workbook holds the sheets, dependency graph and undo history.
type Workbook struct {
	Sheets []*Sheet
	Active int

	rev  map[xlKey]map[xlKey]struct{}
	hist xlHistory
	// clock supplies the time for TODAY and NOW; nil means time.Now (tests inject one).
	clock func() time.Time
	funcs xlRegistry
}

// funcTable returns the worksheet function table, built on first use.
func (w *Workbook) funcTable() xlRegistry {
	if w.funcs == nil {
		w.funcs = xlBuildFuncs()
	}
	return w.funcs
}

// now returns the current time for date functions.
func (w *Workbook) now() time.Time {
	if w.clock != nil {
		return w.clock()
	}
	return time.Now()
}

// Column widths and rows used by the sample workbook.
const (
	sampleColWidth   = 100
	sampleQuarterRow = 10 // first row of the quarterly revenue table
	sampleWideCol    = 96
	sampleLabelCol   = 150
	sampleValueCol   = 120
)

const (
	xlMinWrapLines    = 2 // wrapped text needs at least this many lines to grow a row
	xlWrapRowPadding  = 7 // vertical padding added to an auto-fitted wrapped row
	setColWidthFloor  = 8
	setRowHeightFloor = 8
	defaultColWidth   = 84
	defaultRowHeight  = 22
)

func newSheet(w *Workbook, name string) *Sheet {
	return &Sheet{
		Name: name, wb: w,
		cells: map[CellRef]*xlCell{}, styles: map[CellRef]Style{}, comments: map[CellRef]string{},
		colW: map[int]int{}, rowH: map[int]int{}, hiddenRows: map[int]bool{}, hiddenCols: map[int]bool{},
		filters: map[int]map[string]bool{}, usedVer: -1,
	}
}

// Sheet returns the active sheet.
func (w *Workbook) Sheet() *Sheet {
	if w.Active < 0 || w.Active >= len(w.Sheets) {
		w.Active = 0
	}
	return w.Sheets[w.Active]
}

// SheetByName finds a sheet case-insensitively.
func (w *Workbook) SheetByName(name string) *Sheet { return w.sheetByName(name) }

func (w *Workbook) uniqueName(base string) string {
	name := base
	for i := 2; w.sheetByName(name) != nil; i++ {
		name = fmt.Sprintf("%s (%d)", base, i)
	}
	return name
}

// undoable runs f and records the whole-workbook change as one undo step. It
// records nothing when f reports no change.
func (w *Workbook) undoable(f func() bool) {
	before := w.captureState()
	if !f() {
		return
	}
	after := w.captureState()
	w.push(xlCmd{undo: func() { w.restoreState(before) }, redo: func() { w.restoreState(after) }})
}

// AddSheet appends a sheet named SheetN and returns it (one undo step).
func (w *Workbook) AddSheet() *Sheet {
	var s *Sheet
	w.undoable(func() bool {
		n := len(w.Sheets) + 1
		for w.sheetByName(fmt.Sprintf("Sheet%d", n)) != nil {
			n++
		}
		s = w.addSheetNamed(fmt.Sprintf("Sheet%d", n))
		return true
	})
	return s
}

func (w *Workbook) addSheetNamed(name string) *Sheet {
	s := newSheet(w, w.uniqueName(name))
	w.Sheets = append(w.Sheets, s)
	w.reloadAll()
	return s
}

// DeleteSheet removes sheet i (the last sheet cannot be deleted); formulas that
// pointed at it become #REF!.
func (w *Workbook) DeleteSheet(i int) {
	if i < 0 || i >= len(w.Sheets) || len(w.Sheets) == 1 {
		return
	}
	w.undoable(func() bool { w.deleteSheet(i); return true })
}

func (w *Workbook) deleteSheet(i int) {
	name := w.Sheets[i].Name
	w.Sheets = append(w.Sheets[:i], w.Sheets[i+1:]...)
	if w.Active >= len(w.Sheets) || w.Active > i {
		w.Active = max(0, min(w.Active-1, len(w.Sheets)-1))
	}
	w.rewriteFormulas(func(sp xlRefSpan, cur *Sheet) (string, bool) {
		if sp.hasSheet && strings.EqualFold(sp.sheet, name) {
			return ErrRef, true
		}
		return "", false
	})
	w.reloadAll()
}

// RenameSheet renames sheet i and rewrites references to it.
func (w *Workbook) RenameSheet(i int, name string) {
	name = strings.TrimSpace(name)
	if i < 0 || i >= len(w.Sheets) || name == "" || w.Sheets[i].Name == name {
		return
	}
	if o := w.sheetByName(name); o != nil && o != w.Sheets[i] {
		return
	}
	w.undoable(func() bool { w.renameSheet(i, name); return true })
}

func (w *Workbook) renameSheet(i int, name string) {
	old := w.Sheets[i].Name
	// Rewrite prefixes textually so quoting follows the new name.
	for _, s := range w.Sheets {
		for _, c := range s.cells {
			if len(c.raw) < 2 || c.raw[0] != '=' {
				continue
			}
			body := c.raw[1:]
			spans := xlScanRefs(body)
			var sb strings.Builder
			last := 0
			for _, sp := range spans {
				if sp.hasSheet && strings.EqualFold(sp.sheet, old) {
					sb.WriteString(body[last:sp.from])
					sb.WriteString(xlQuoteSheet(name) + "!")
					last = sp.refFrom
				}
			}
			sb.WriteString(body[last:])
			c.raw = "=" + sb.String()
		}
	}
	w.Sheets[i].Name = name
	w.reloadAll()
}

// rewriteFormulas applies fn to every reference span of every formula.
func (w *Workbook) rewriteFormulas(fn func(sp xlRefSpan, cur *Sheet) (string, bool)) {
	for _, s := range w.Sheets {
		for _, c := range s.cells {
			if len(c.raw) < 2 || c.raw[0] != '=' {
				continue
			}
			cur := s
			c.raw = "=" + xlRewriteSpans(c.raw[1:], func(sp xlRefSpan) (string, bool) { return fn(sp, cur) })
		}
	}
}

// Raw returns the typed text of a cell ("" when empty).
func (s *Sheet) Raw(row, col int) string {
	if c := s.cells[CellRef{row, col}]; c != nil {
		return c.raw
	}
	return ""
}

// Value returns the computed value of a cell.
func (s *Sheet) Value(row, col int) Value {
	if c := s.cells[CellRef{row, col}]; c != nil {
		return c.val
	}
	return Value{}
}

func (s *Sheet) cellVal(ref CellRef) Value {
	if c := s.cells[ref]; c != nil {
		return c.val
	}
	return Value{}
}

// Display returns the number-formatted text of a cell.
func (s *Sheet) Display(row, col int) string {
	c := s.cells[CellRef{row, col}]
	if c == nil {
		return ""
	}
	return FormatValue(c.val, s.styles[CellRef{row, col}])
}

// BaseStyle returns the style without conditional formatting.
func (s *Sheet) BaseStyle(row, col int) Style { return s.styles[CellRef{row, col}] }

// IsFormula reports whether the cell holds a formula.
func (s *Sheet) IsFormula(row, col int) bool {
	c := s.cells[CellRef{row, col}]
	return c != nil && len(c.raw) > 1 && c.raw[0] == '='
}

// EffectiveAlign returns the horizontal alignment: explicit, else numbers right,
// booleans/errors center, text left.
func (s *Sheet) EffectiveAlign(row, col int) string {
	if a := s.styles[CellRef{row, col}].Align; a != "" {
		return a
	}
	switch s.Value(row, col).Kind {
	case KindNumber:
		return "right"
	case KindBool, KindError:
		return "center"
	}
	return "left"
}

// SetRaw writes typed text ("=..." is a formula) and recalculates dependents.
// It is not recorded in history; use Workbook.Edit for undoable edits.
func (s *Sheet) SetRaw(row, col int, raw string) {
	ref := CellRef{row, col}
	if s.setRawQuiet(ref, raw) {
		s.wb.recalc([]xlKey{{s, ref}})
	}
}

// setRawQuiet stores the cell and its dependency edges without recalculating.
func (s *Sheet) setRawQuiet(ref CellRef, raw string) bool {
	if ref.Row < 0 || ref.Row >= MaxRows || ref.Col < 0 || ref.Col >= MaxCols {
		return false
	}
	w := s.wb
	k := xlKey{s, ref}
	old := s.cells[ref]
	if old != nil {
		if old.raw == raw {
			return false
		}
		w.dropEdges(k, old.deps)
	}
	s.ver++
	if raw == "" {
		delete(s.cells, ref)
		return true
	}
	c := &xlCell{raw: raw}
	if len(raw) > 1 && raw[0] == '=' {
		ast, err := xlParse(raw[1:])
		if err != nil {
			c.val = xlErrV(ErrSyn)
		} else {
			c.ast = ast
			c.deps = w.collectDeps(s, ast)
			w.addEdges(k, c.deps)
		}
	} else {
		lit := xlInfer(raw)
		c.val = lit.val
		if lit.numFmt != "" {
			st := s.styles[ref]
			if st.NumFmt == "" || st.NumFmt == "General" {
				st.NumFmt, st.Decimals = lit.numFmt, lit.dec
				s.styles[ref] = st
			}
		}
	}
	s.cells[ref] = c
	return true
}

// SetStyle applies f to every cell of r.
func (s *Sheet) SetStyle(r Range, f func(*Style)) {
	r = r.clip()
	for row := r.Min.Row; row <= r.Max.Row; row++ {
		for col := r.Min.Col; col <= r.Max.Col; col++ {
			ref := CellRef{row, col}
			st := s.styles[ref].migrated()
			f(&st)
			if st == (Style{}) {
				delete(s.styles, ref)
			} else {
				s.styles[ref] = st
			}
		}
	}
	s.ver++
}

func (s *Sheet) ColWidth(col int) int {
	if w, ok := s.colW[col]; ok {
		return w
	}
	return defaultColWidth
}

// RowHeight returns the row's height: the height set by the user, else the
// height its wrapped text needs, else the default.
func (s *Sheet) RowHeight(row int) int {
	if h, ok := s.rowH[row]; ok {
		return h
	}
	if h, ok := s.autoRows()[row]; ok {
		return h
	}
	return defaultRowHeight
}

// autoRows maps rows without a manual height to the height their wrapped
// cells need. The map is rebuilt when the sheet changes.
func (s *Sheet) autoRows() map[int]int {
	if s.autoVer == s.ver+1 {
		return s.auto
	}
	s.autoVer = s.ver + 1
	s.auto = nil
	for ref, st := range s.styles {
		if !st.Wrap {
			continue
		}
		if _, manual := s.rowH[ref.Row]; manual {
			continue
		}
		text := s.Display(ref.Row, ref.Col)
		lines := wrapLineCount(text, s.ColWidth(ref.Col)-st.Indent*indentStepPx, st.FontSize)
		if lines < xlMinWrapLines {
			continue
		}
		fs := st.FontSize
		if fs == 0 {
			fs = defaultFontSize
		}
		h := lines*wrapLineHeight*fs/defaultFontSize + xlWrapRowPadding
		if h <= defaultRowHeight {
			continue
		}
		if s.auto == nil {
			s.auto = map[int]int{}
		}
		s.auto[ref.Row] = max(s.auto[ref.Row], h)
	}
	return s.auto
}

func (s *Sheet) SetColWidth(col, w int) {
	if w < setColWidthFloor {
		w = setColWidthFloor
	}
	s.colW[col] = w
	s.ver++
}

func (s *Sheet) SetRowHeight(row, h int) {
	if h < setRowHeightFloor {
		h = setRowHeightFloor
	}
	s.rowH[row] = h
}

// SetRowHidden hides or shows a row manually; SetColHidden does the same for columns.
func (s *Sheet) SetRowHidden(row int, hidden bool) { xlSetFlag(s.hiddenRows, row, hidden) }
func (s *Sheet) SetColHidden(col int, hidden bool) { xlSetFlag(s.hiddenCols, col, hidden) }
func (s *Sheet) ColHidden(col int) bool            { return s.hiddenCols[col] }

func xlSetFlag(m map[int]bool, k int, on bool) {
	if on {
		m[k] = true
	} else {
		delete(m, k)
	}
}

// UsedRange is the bounding box of non-empty cells (A1:A1 when empty).
func (s *Sheet) UsedRange() Range {
	if s.usedVer == s.ver {
		return s.used
	}
	r := Range{}
	first := true
	for ref := range s.cells {
		if first {
			r, first = Range{ref, ref}, false
			continue
		}
		r.Min.Row, r.Min.Col = min(r.Min.Row, ref.Row), min(r.Min.Col, ref.Col)
		r.Max.Row, r.Max.Col = max(r.Max.Row, ref.Row), max(r.Max.Col, ref.Col)
	}
	s.used, s.usedVer = r, s.ver
	return r
}

func (s *Sheet) FreezeRows() int { return s.freezeR }
func (s *Sheet) FreezeCols() int { return s.freezeC }
func (s *Sheet) SetFreeze(rows, cols int) {
	s.freezeR, s.freezeC = max(rows, 0), max(cols, 0)
}

// SetFilter restricts rows below the header to the allowed display texts of col;
// nil clears that column's filter.
func (s *Sheet) SetFilter(col int, allowed map[string]bool) {
	if allowed == nil {
		delete(s.filters, col)
	} else {
		s.filters[col] = allowed
	}
	s.ver++
}

// SetFilterHeader sets the header row of the filtered table (default 0).
func (s *Sheet) SetFilterHeader(row int) { s.filterHdr = row }
func (s *Sheet) FilterHeader() int       { return s.filterHdr }
func (s *Sheet) HasFilter() bool         { return len(s.filters) > 0 }

// FilterAllowed returns the allowed set for col (nil when not filtered).
func (s *Sheet) FilterAllowed(col int) map[string]bool { return s.filters[col] }

// FilterValues lists the distinct display texts of col below the header.
func (s *Sheet) FilterValues(col int) []string {
	seen := map[string]bool{}
	var out []string
	u := s.UsedRange()
	for row := s.filterHdr + 1; row <= u.Max.Row; row++ {
		d := s.Display(row, col)
		if !seen[d] {
			seen[d] = true
			out = append(out, d)
		}
	}
	sort.Slice(out, func(i, j int) bool { return xlCompareText(out[i], out[j]) < 0 })
	return out
}

func xlCompareText(a, b string) int {
	fa, oka := xlParseNumber(a)
	fb, okb := xlParseNumber(b)
	if oka && okb {
		switch {
		case fa < fb:
			return -1
		case fa > fb:
			return 1
		}
		return 0
	}
	return strings.Compare(strings.ToLower(a), strings.ToLower(b))
}

// HiddenRow reports whether a row is hidden manually or by an active filter.
func (s *Sheet) HiddenRow(row int) bool {
	if s.hiddenRows[row] {
		return true
	}
	if len(s.filters) == 0 || row <= s.filterHdr {
		return false
	}
	for col, allowed := range s.filters {
		if !allowed[s.Display(row, col)] {
			return true
		}
	}
	return false
}

// VisibleRows counts non-hidden rows in [from, to].
func (s *Sheet) VisibleRows(from, to int) int {
	n := 0
	for r := from; r <= to; r++ {
		if !s.HiddenRow(r) {
			n++
		}
	}
	return n
}

// Find returns cells whose text or formula contains query, in row-major order.
func (s *Sheet) Find(query string, matchCase bool) []CellRef {
	if query == "" {
		return nil
	}
	if !matchCase {
		query = strings.ToLower(query)
	}
	var out []CellRef
	for ref, c := range s.cells {
		hay := FormatValue(c.val, s.styles[ref]) + "\x00" + c.raw
		if !matchCase {
			hay = strings.ToLower(hay)
		}
		if strings.Contains(hay, query) {
			out = append(out, ref)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Row != out[j].Row {
			return out[i].Row < out[j].Row
		}
		return out[i].Col < out[j].Col
	})
	return out
}

// Merge records a merged range (top-left cell is the anchor).
func (s *Sheet) Merge(r Range) {
	if r.Min == r.Max {
		return
	}
	s.Unmerge(r)
	s.merges = append(s.merges, r)
}

// Unmerge removes merges that intersect r.
func (s *Sheet) Unmerge(r Range) {
	keep := s.merges[:0]
	for _, m := range s.merges {
		if !m.Intersects(r) {
			keep = append(keep, m)
		}
	}
	s.merges = keep
}

// MergeAt returns the merged range covering a cell.
func (s *Sheet) MergeAt(row, col int) (Range, bool) {
	for _, m := range s.merges {
		if m.Contains(CellRef{row, col}) {
			return m, true
		}
	}
	return Range{}, false
}

func (s *Sheet) Merges() []Range { return append([]Range(nil), s.merges...) }

// SetValidation attaches a dropdown list to r (empty list clears every rule
// that overlaps r).
func (s *Sheet) SetValidation(r Range, list []string) {
	s.dropValidation(r)
	if len(list) > 0 {
		s.valid = append(s.valid, xlValid{rng: r, kind: "list", list: append([]string(nil), list...)})
	}
}

// SetValidationBounds attaches a whole-number, decimal or text-length bound
// ("whole", "decimal", "length") to r, replacing overlapping rules.
func (s *Sheet) SetValidationBounds(r Range, kind string, lo, hi float64) {
	s.dropValidation(r)
	s.valid = append(s.valid, xlValid{rng: r, kind: kind, min: lo, max: hi})
}

func (s *Sheet) dropValidation(r Range) {
	keep := s.valid[:0]
	for _, v := range s.valid {
		if !v.rng.Intersects(r) {
			keep = append(keep, v)
		}
	}
	s.valid = keep
}

// ValidationList returns the dropdown options for a cell ("" kinds excluded).
func (s *Sheet) ValidationList(row, col int) []string {
	for i := len(s.valid) - 1; i >= 0; i-- {
		if v := s.valid[i]; v.kind == "list" && v.rng.Contains(CellRef{row, col}) {
			return v.list
		}
	}
	return nil
}

// ValidationRules returns each rule's range text and kind, for tests and UI.
func (s *Sheet) ValidationRules() (ranges []Range, kinds []string) {
	for _, v := range s.valid {
		ranges, kinds = append(ranges, v.rng), append(kinds, v.kind)
	}
	return ranges, kinds
}

// CheckValidation reports whether raw may be committed to the cell. Formulas
// and empty text always pass.
func (s *Sheet) CheckValidation(row, col int, raw string) (ok bool, msg string) {
	if raw == "" || strings.HasPrefix(raw, "=") {
		return true, ""
	}
	ref := CellRef{row, col}
	for i := len(s.valid) - 1; i >= 0; i-- {
		v := s.valid[i]
		if !v.rng.Contains(ref) {
			continue
		}
		if v.kind == "list" {
			for _, o := range v.list {
				if strings.EqualFold(o, raw) {
					return true, ""
				}
			}
			return false, "Pick one of: " + strings.Join(v.list, ", ")
		}
		n, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
		switch v.kind {
		case "length":
			n, err = float64(len([]rune(raw))), nil
		case "whole":
			if err == nil && n != math.Trunc(n) {
				return false, "Enter a whole number."
			}
		}
		lo, hi := strconv.FormatFloat(v.min, 'f', -1, 64), strconv.FormatFloat(v.max, 'f', -1, 64)
		if err != nil {
			return false, "Enter a number between " + lo + " and " + hi + "."
		}
		if n < v.min || n > v.max {
			return false, "Value must be between " + lo + " and " + hi + "."
		}
		return true, ""
	}
	return true, ""
}

func (s *Sheet) SetComment(row, col int, text string) {
	if text == "" {
		delete(s.comments, CellRef{row, col})
		return
	}
	s.comments[CellRef{row, col}] = text
}

func (s *Sheet) Comment(row, col int) string { return s.comments[CellRef{row, col}] }

// Stats aggregates a range for the status bar.
func (s *Sheet) Stats(r Range) (count, numCount int, sum, avg, minV, maxV float64) {
	r = r.clip()
	u := s.UsedRange()
	r.Max.Row, r.Max.Col = min(r.Max.Row, u.Max.Row), min(r.Max.Col, u.Max.Col)
	minV, maxV = math.Inf(1), math.Inf(-1)
	for row := r.Min.Row; row <= r.Max.Row; row++ {
		for col := r.Min.Col; col <= r.Max.Col; col++ {
			v := s.cellVal(CellRef{row, col})
			if v.Kind == KindEmpty {
				continue
			}
			count++
			if v.Kind == KindNumber {
				numCount++
				sum += v.Num
				minV, maxV = math.Min(minV, v.Num), math.Max(maxV, v.Num)
			}
		}
	}
	if numCount == 0 {
		return count, 0, 0, 0, 0, 0
	}
	return count, numCount, sum, sum / float64(numCount), minV, maxV
}

// xlCompareValues orders values for sorting: numbers, text, bools, errors; empty last.
func xlSortRank(v Value) int {
	switch v.Kind {
	case KindNumber:
		return xlRankNumber
	case KindString:
		return xlRankString
	case KindBool:
		return xlRankBool
	case KindError:
		return xlRankError
	}
	return xlRankEmpty
}

// Sort orders the rows of r by the absolute column keyCol. Formulas move with
// their rows and their relative references shift accordingly. Not recorded in
// history; use Workbook.SortRange for an undoable sort.
func (s *Sheet) Sort(r Range, keyCol int, asc, hasHeader bool) {
	r = r.clip()
	if hasHeader {
		r.Min.Row++
	}
	if r.Min.Row >= r.Max.Row || keyCol < r.Min.Col || keyCol > r.Max.Col {
		return
	}
	type rowData struct {
		src  int
		raws []string
		sts  []Style
		cmts []string
		key  Value
	}
	n := r.Rows()
	rows := make([]rowData, n)
	for i := range rows {
		row := r.Min.Row + i
		d := rowData{src: row}
		for col := r.Min.Col; col <= r.Max.Col; col++ {
			d.raws = append(d.raws, s.Raw(row, col))
			d.sts = append(d.sts, s.styles[CellRef{row, col}])
			d.cmts = append(d.cmts, s.comments[CellRef{row, col}])
		}
		d.key = s.cellVal(CellRef{row, keyCol})
		rows[i] = d
	}
	sort.SliceStable(rows, func(i, j int) bool {
		a, b := rows[i].key, rows[j].key
		ra, rb := xlSortRank(a), xlSortRank(b)
		if ra == 4 || rb == 4 {
			return ra < rb // empty always last
		}
		var c int
		if ra != rb {
			c = ra - rb
		} else {
			c = xlCompareScalar(a, b)
		}
		if asc {
			return c < 0
		}
		return c > 0
	})
	var seeds []xlKey
	for i, d := range rows {
		row := r.Min.Row + i
		for j := range d.raws {
			col := r.Min.Col + j
			ref := CellRef{row, col}
			raw := ShiftFormula(d.raws[j], 0, row-d.src)
			s.setRawQuiet(ref, raw)
			if d.sts[j] == (Style{}) {
				delete(s.styles, ref)
			} else {
				s.styles[ref] = d.sts[j]
			}
			if d.cmts[j] == "" {
				delete(s.comments, ref)
			} else {
				s.comments[ref] = d.cmts[j]
			}
			seeds = append(seeds, xlKey{s, ref})
		}
	}
	s.ver++
	s.wb.recalc(seeds)
}

// xlCompareScalar compares two non-empty values of the same rank.
func xlCompareScalar(a, b Value) int {
	switch a.Kind {
	case KindNumber:
		switch {
		case a.Num < b.Num:
			return -1
		case a.Num > b.Num:
			return 1
		}
	case KindString:
		return strings.Compare(strings.ToLower(a.Str), strings.ToLower(b.Str))
	case KindBool:
		if a.Bool != b.Bool {
			if b.Bool {
				return -1
			}
			return 1
		}
	case KindError:
		return strings.Compare(a.Err, b.Err)
	}
	return 0
}

// ---- structural edits ----

func xlShiftIdx(i, at, n int) int {
	if i >= at {
		return i + n
	}
	return i
}

// xlShiftRange moves a range through an insert (cnt rows/cols before at) or a
// delete (cnt starting at at). It reports false when the range disappears.
func xlShiftRange(r Range, at, cnt int, del, rows bool, limit int) (Range, bool) {
	get := func(c CellRef) int {
		if rows {
			return c.Row
		}
		return c.Col
	}
	put := func(c *CellRef, v int) {
		if rows {
			c.Row = v
		} else {
			c.Col = v
		}
	}
	lo, hi := get(r.Min), get(r.Max)
	if !del {
		lo, hi = xlShiftIdx(lo, at, cnt), xlShiftIdx(hi, at, cnt)
	} else {
		end := at + cnt
		switch {
		case lo >= at && hi < end:
			return r, false
		case lo >= end:
			lo -= cnt
		case lo >= at:
			lo = at
		}
		switch {
		case hi >= end:
			hi -= cnt
		case hi >= at:
			hi = at - 1
		}
	}
	if lo > hi || lo >= limit {
		return r, false
	}
	put(&r.Min, lo)
	put(&r.Max, min(hi, limit-1))
	return r, true
}

// InsertRows inserts n blank rows before row at.
func (w *Workbook) InsertRows(s *Sheet, at, n int) { w.structural(s, at, n, true) }

// DeleteRows deletes n rows starting at row at.
func (w *Workbook) DeleteRows(s *Sheet, at, n int) { w.structural(s, at, -n, true) }

// InsertCols inserts n blank columns before column at.
func (w *Workbook) InsertCols(s *Sheet, at, n int) { w.structural(s, at, n, false) }

// DeleteCols deletes n columns starting at column at.
func (w *Workbook) DeleteCols(s *Sheet, at, n int) { w.structural(s, at, -n, false) }

// structural inserts (n>0) or deletes (n<0) rows/cols with undo, rewriting
// formulas workbook-wide.
func (w *Workbook) structural(s *Sheet, at, n int, rows bool) {
	if n == 0 || at < 0 {
		return
	}
	before := w.captureState()
	del := n < 0
	cnt := n
	if del {
		cnt = -n
	}
	limit := MaxCols
	if rows {
		limit = MaxRows
	}
	part := func(p xlRefPart) int {
		if rows {
			return p.Row
		}
		return p.Col
	}
	setPart := func(p xlRefPart, v int) xlRefPart {
		if rows {
			p.Row = v
		} else {
			p.Col = v
		}
		return p
	}
	// 1. rewrite formulas that point into s.
	w.rewriteFormulas(func(sp xlRefSpan, cur *Sheet) (string, bool) {
		target := cur
		if sp.hasSheet {
			target = w.sheetByName(sp.sheet)
		}
		if target != s {
			return "", false
		}
		a, b := sp.a, sp.b
		pa, pb := part(a), part(b)
		if !del {
			a = setPart(a, xlShiftIdx(pa, at, cnt))
			b = setPart(b, xlShiftIdx(pb, at, cnt))
			if part(a) >= limit || part(b) >= limit {
				return ErrRef, true
			}
			return xlSpanText(sp, a, b), true
		}
		lo, hi := min(pa, pb), max(pa, pb)
		end := at + cnt // first index after the deleted block
		if lo >= at && hi < end {
			return ErrRef, true
		}
		fix := func(v int) int {
			switch {
			case v >= end:
				return v - cnt
			case v >= at:
				return at
			}
			return v
		}
		if sp.isRange {
			// shrink: start inside block moves to block start; end inside moves before it
			na, nb := pa, pb
			if pa >= at && pa < end {
				na = at
			} else if pa >= end {
				na = pa - cnt
			}
			if pb >= at && pb < end {
				nb = at - 1
			} else if pb >= end {
				nb = pb - cnt
			}
			if na > nb {
				return ErrRef, true
			}
			return xlSpanText(sp, setPart(a, na), setPart(b, nb)), true
		}
		return xlSpanText(sp, setPart(a, fix(pa)), setPart(b, fix(pb))), true
	})
	// 2. move cells of s.
	mv := func(ref CellRef) (CellRef, bool) {
		v := ref.Col
		if rows {
			v = ref.Row
		}
		switch {
		case !del:
			v = xlShiftIdx(v, at, cnt)
		case v >= at && v < at+cnt:
			return ref, false
		case v >= at+cnt:
			v -= cnt
		}
		if v >= limit {
			return ref, false
		}
		if rows {
			ref.Row = v
		} else {
			ref.Col = v
		}
		return ref, true
	}
	cells := map[CellRef]*xlCell{}
	for ref, c := range s.cells {
		if nr, ok := mv(ref); ok {
			cells[nr] = c
		}
	}
	styles := map[CellRef]Style{}
	for ref, st := range s.styles {
		if nr, ok := mv(ref); ok {
			styles[nr] = st
		}
	}
	comments := map[CellRef]string{}
	for ref, t := range s.comments {
		if nr, ok := mv(ref); ok {
			comments[nr] = t
		}
	}
	dims := s.colW
	if rows {
		dims = s.rowH
	}
	nd := map[int]int{}
	for i, v := range dims {
		r := CellRef{Col: i}
		if rows {
			r = CellRef{Row: i}
		}
		if nr, ok := mv(r); ok {
			if rows {
				nd[nr.Row] = v
			} else {
				nd[nr.Col] = v
			}
		}
	}
	s.cells, s.styles, s.comments = cells, styles, comments
	if rows {
		s.rowH = nd
	} else {
		s.colW = nd
	}
	var merges []Range
	for _, m := range s.merges {
		a, ok1 := mv(m.Min)
		b, ok2 := mv(m.Max)
		if ok1 && ok2 && a != b {
			merges = append(merges, NormRange(a, b))
		}
	}
	s.merges = merges
	var conds []*xlCondEntry
	for _, e := range s.conds {
		if r, ok := xlShiftRange(e.rng, at, cnt, del, rows, limit); ok {
			e.rng, e.ver = r, -1
			conds = append(conds, e)
		}
	}
	s.conds = conds
	var valid []xlValid
	for _, v := range s.valid {
		if r, ok := xlShiftRange(v.rng, at, cnt, del, rows, limit); ok {
			v.rng = r
			valid = append(valid, v)
		}
	}
	s.valid = valid
	hidden := s.hiddenCols
	if rows {
		hidden = s.hiddenRows
	}
	moved := map[int]bool{}
	for i := range hidden {
		r := CellRef{Col: i}
		if rows {
			r = CellRef{Row: i}
		}
		if nr, ok := mv(r); ok {
			if rows {
				moved[nr.Row] = true
			} else {
				moved[nr.Col] = true
			}
		}
	}
	if rows {
		s.hiddenRows = moved
	} else {
		s.hiddenCols = moved
	}
	w.reloadAll()
	after := w.captureState()
	w.push(xlCmd{
		undo: func() { w.restoreState(before) },
		redo: func() { w.restoreState(after) },
	})
}

// ---- starter workbook ----

// NewWorkbook builds the starter workbook: Sheet1 holds a regional sales table
// with formulas, Sheet2 is empty, and Summary rolls Sheet1 up with
// cross-sheet formulas.
func NewWorkbook() *Workbook {
	w := &Workbook{}
	s1 := newSheet(w, "Sheet1")
	sum := newSheet(w, "Summary")
	w.Sheets = []*Sheet{s1, newSheet(w, "Sheet2"), sum}
	hdr := []string{"Region", "Product", "Q1", "Q2", "Q3", "Q4", "Total", "Growth"}
	for c, h := range hdr {
		s1.setRawQuiet(CellRef{0, c}, h)
	}
	data := []struct {
		region, product string
		q               [4]float64
	}{
		{"North", "Widget", [4]float64{12500, 13200, 14100, 15800}},
		{"North", "Gadget", [4]float64{8300, 8900, 9400, 9100}},
		{"South", "Widget", [4]float64{10200, 9800, 11400, 12900}},
		{"South", "Gadget", [4]float64{7600, 8100, 7900, 8800}},
		{"East", "Widget", [4]float64{14800, 15200, 16700, 17300}},
		{"East", "Gadget", [4]float64{9100, 9700, 10300, 10100}},
		{"West", "Widget", [4]float64{11700, 12400, 12100, 13600}},
		{"West", "Gadget", [4]float64{6900, 7400, 8200, 8700}},
	}
	for i, d := range data {
		row := i + 1
		s1.setRawQuiet(CellRef{row, 0}, d.region)
		s1.setRawQuiet(CellRef{row, 1}, d.product)
		for q := 0; q < 4; q++ {
			s1.setRawQuiet(CellRef{row, 2 + q}, fmt.Sprintf("%.0f", d.q[q]))
		}
		s1.setRawQuiet(CellRef{row, 6}, fmt.Sprintf("=SUM(C%d:F%d)", row+1, row+1))
		s1.setRawQuiet(CellRef{row, 7}, fmt.Sprintf("=IFERROR((F%d-C%d)/C%d,0)", row+1, row+1, row+1))
	}
	last := len(data) + 1
	s1.setRawQuiet(CellRef{last, 0}, "Total")
	for c := 2; c <= 6; c++ {
		col := ColName(c)
		s1.setRawQuiet(CellRef{last, c}, fmt.Sprintf("=SUM(%s2:%s%d)", col, col, last))
	}
	s1.setRawQuiet(CellRef{last, 7}, fmt.Sprintf("=IFERROR((F%d-C%d)/C%d,0)", last+1, last+1, last+1))
	s1.SetStyle(Range{CellRef{0, 0}, CellRef{0, 7}}, func(st *Style) {
		st.Bold, st.BG, st.Align = true, "#DDEBF7", "center"
		st.Edges.Bottom = EdgeThin
	})
	s1.SetStyle(Range{CellRef{1, 2}, CellRef{last, 6}}, func(st *Style) { st.NumFmt, st.Decimals = "Currency", -1 })
	s1.SetStyle(Range{CellRef{1, 7}, CellRef{last, 7}}, func(st *Style) { st.NumFmt, st.Decimals = "Percent", 1 })
	s1.SetStyle(Range{CellRef{last, 0}, CellRef{last, 7}}, func(st *Style) { st.Bold, st.Edges.Top = true, EdgeThin })
	s1.SetColWidth(0, sampleWideCol)
	s1.SetColWidth(1, sampleWideCol)
	for c := 2; c <= 6; c++ {
		s1.SetColWidth(c, sampleColWidth)
	}
	s1.SetFreeze(1, 0)
	seedSummary(sum)
	w.reloadAll()
	return w
}

// seedSummary fills the Summary sheet: headline numbers computed from Sheet1
// with SUM, AVERAGE, MAX and IF, and a quarter table a chart can bind to.
func seedSummary(s *Sheet) {
	set := func(ref, raw string) {
		c, _ := ParseA1(ref)
		s.setRawQuiet(c, raw)
	}
	set("A1", "Executive summary")
	set("A3", "Total revenue")
	set("B3", "=SUM(Sheet1!C2:F9)")
	set("A4", "Average line total")
	set("B4", "=AVERAGE(Sheet1!G2:G9)")
	set("A5", "Best Q4 line")
	set("B5", "=MAX(Sheet1!F2:F9)")
	set("A6", "Q4 vs Q1")
	set("B6", "=(SUM(Sheet1!F2:F9)-SUM(Sheet1!C2:C9))/SUM(Sheet1!C2:C9)")
	set("A7", "Status")
	set("B7", `=IF(B6>0.1,"On track","Review")`)
	set("A9", "Quarter")
	set("B9", "Revenue")
	for i, col := range []string{"C", "D", "E", "F"} {
		row := strconv.Itoa(sampleQuarterRow + i)
		set("A"+row, "Q"+strconv.Itoa(i+1))
		set("B"+row, "=SUM(Sheet1!"+col+"2:"+col+"9)")
	}
	s.SetStyle(Range{CellRef{0, 0}, CellRef{0, 1}}, func(st *Style) { st.Bold, st.FontSize = true, 18 })
	s.SetStyle(Range{CellRef{8, 0}, CellRef{8, 1}}, func(st *Style) {
		st.Bold, st.BG, st.Align = true, "#DDEBF7", "center"
		st.Edges.Bottom = EdgeThin
	})
	s.SetStyle(Range{CellRef{2, 1}, CellRef{4, 1}}, func(st *Style) { st.NumFmt, st.Decimals = "Currency", -1 })
	s.SetStyle(Range{CellRef{5, 1}, CellRef{5, 1}}, func(st *Style) { st.NumFmt, st.Decimals = "Percent", 1 })
	s.SetStyle(Range{CellRef{9, 1}, CellRef{12, 1}}, func(st *Style) { st.NumFmt, st.Decimals = "Currency", -1 })
	s.SetStyle(Range{CellRef{6, 1}, CellRef{6, 1}}, func(st *Style) { st.Bold = true })
	s.SetColWidth(0, sampleLabelCol)
	s.SetColWidth(1, sampleValueCol)
	s.AddCondFormat(Range{CellRef{9, 1}, CellRef{12, 1}}, CondRule{Kind: "databar", Preset: "green"})
}

// ---- state snapshots (sort, structural edits) ----

type xlSheetState struct {
	sh         *Sheet
	name       string
	raws       map[CellRef]string
	styles     map[CellRef]Style
	comments   map[CellRef]string
	colW, rowH map[int]int
	merges     []Range
	conds      []xlCondEntry
	valid      []xlValid
	hiddenR    map[int]bool
	hiddenC    map[int]bool
}

type xlState struct {
	sheets []xlSheetState
	active int
}

func (w *Workbook) captureState() xlState {
	st := xlState{active: w.Active}
	for _, s := range w.Sheets {
		ss := xlSheetState{sh: s, name: s.Name, raws: map[CellRef]string{}, styles: map[CellRef]Style{},
			comments: map[CellRef]string{}, colW: map[int]int{}, rowH: map[int]int{}, merges: append([]Range(nil), s.merges...),
			valid: append([]xlValid(nil), s.valid...), hiddenR: map[int]bool{}, hiddenC: map[int]bool{}}
		for _, e := range s.conds {
			ss.conds = append(ss.conds, *e)
		}
		for k := range s.hiddenRows {
			ss.hiddenR[k] = true
		}
		for k := range s.hiddenCols {
			ss.hiddenC[k] = true
		}
		for ref, c := range s.cells {
			ss.raws[ref] = c.raw
		}
		for k, v := range s.styles {
			ss.styles[k] = v
		}
		for k, v := range s.comments {
			ss.comments[k] = v
		}
		for k, v := range s.colW {
			ss.colW[k] = v
		}
		for k, v := range s.rowH {
			ss.rowH[k] = v
		}
		st.sheets = append(st.sheets, ss)
	}
	return st
}

func (w *Workbook) restoreState(st xlState) {
	w.Sheets = w.Sheets[:0]
	for _, ss := range st.sheets {
		s := ss.sh
		s.Name = ss.name
		s.cells = map[CellRef]*xlCell{}
		for ref, raw := range ss.raws {
			s.cells[ref] = &xlCell{raw: raw}
			if len(raw) <= 1 || raw[0] != '=' {
				s.cells[ref].val = xlInfer(raw).val
			}
		}
		s.styles, s.comments, s.colW, s.rowH = map[CellRef]Style{}, map[CellRef]string{}, map[int]int{}, map[int]int{}
		for k, v := range ss.styles {
			s.styles[k] = v
		}
		for k, v := range ss.comments {
			s.comments[k] = v
		}
		for k, v := range ss.colW {
			s.colW[k] = v
		}
		for k, v := range ss.rowH {
			s.rowH[k] = v
		}
		s.merges = append([]Range(nil), ss.merges...)
		s.valid = append([]xlValid(nil), ss.valid...)
		s.conds = nil
		for _, e := range ss.conds {
			c := e
			c.ver = -1
			s.conds = append(s.conds, &c)
		}
		s.hiddenRows, s.hiddenCols = map[int]bool{}, map[int]bool{}
		for k := range ss.hiddenR {
			s.hiddenRows[k] = true
		}
		for k := range ss.hiddenC {
			s.hiddenCols[k] = true
		}
		w.Sheets = append(w.Sheets, s)
	}
	w.Active = min(st.active, len(w.Sheets)-1)
	w.reloadAll()
}

package main

import (
	"strconv"
	"strings"
)

// xlTSVField quotes a field when it holds tabs, newlines or quotes.
func xlTSVField(s string) string {
	if strings.ContainsAny(s, "\t\n\r\"") {
		return "\"" + strings.ReplaceAll(s, "\"", "\"\"") + "\""
	}
	return s
}

// CopyTSV serializes r as tab-separated text. Cells are written as typed (formulas
// keep their "=" text) so an internal round trip is lossless; rows end with "\n".
func (s *Sheet) CopyTSV(r Range) string {
	r = r.clip()
	var sb strings.Builder
	for row := r.Min.Row; row <= r.Max.Row; row++ {
		for col := r.Min.Col; col <= r.Max.Col; col++ {
			if col > r.Min.Col {
				sb.WriteByte('\t')
			}
			sb.WriteString(xlTSVField(s.Raw(row, col)))
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

// CopyDisplayTSV serializes the formatted display text instead (what other
// applications usually want).
func (s *Sheet) CopyDisplayTSV(r Range) string {
	r = r.clip()
	var sb strings.Builder
	for row := r.Min.Row; row <= r.Max.Row; row++ {
		for col := r.Min.Col; col <= r.Max.Col; col++ {
			if col > r.Min.Col {
				sb.WriteByte('\t')
			}
			sb.WriteString(xlTSVField(s.Display(row, col)))
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

// ParseTSV splits tab-separated text into a grid, honoring double-quoted fields.
func ParseTSV(tsv string) [][]string {
	tsv = strings.ReplaceAll(tsv, "\r\n", "\n")
	var rows [][]string
	var row []string
	var field strings.Builder
	inQ, started := false, false
	flushField := func() {
		row = append(row, field.String())
		field.Reset()
		started = false
	}
	flushRow := func() {
		flushField()
		rows = append(rows, row)
		row = nil
	}
	for i := 0; i < len(tsv); i++ {
		c := tsv[i]
		switch {
		case inQ:
			if c == '"' {
				if i+1 < len(tsv) && tsv[i+1] == '"' {
					field.WriteByte('"')
					i++
				} else {
					inQ = false
				}
			} else {
				field.WriteByte(c)
			}
		case c == '"' && !started:
			inQ, started = true, true
		case c == '\t':
			flushField()
		case c == '\n':
			flushRow()
		default:
			started = true
			field.WriteByte(c)
		}
	}
	if started || field.Len() > 0 || len(row) > 0 {
		flushRow()
	}
	return rows
}

// PasteTSV writes tab-separated text with its top-left at `at` as one undo step
// and returns the pasted range.
func (w *Workbook) PasteTSV(s *Sheet, at CellRef, tsv string) Range {
	grid := ParseTSV(tsv)
	if len(grid) == 0 {
		return Range{at, at}
	}
	wd := 0
	for _, r := range grid {
		wd = max(wd, len(r))
	}
	rng := Range{at, CellRef{at.Row + len(grid) - 1, at.Col + max(wd, 1) - 1}}.clip()
	w.EditRange(s, rng, xlPad(grid, rng))
	return rng
}

func xlPad(grid [][]string, rng Range) [][]string {
	out := make([][]string, rng.Rows())
	for i := range out {
		out[i] = make([]string, rng.Cols())
		if i < len(grid) {
			copy(out[i], grid[i])
		}
	}
	return out
}

// ClearRange empties the contents of r (styles stay) as one undo step.
func (w *Workbook) ClearRange(s *Sheet, r Range) {
	refs := xlRangeRefs(r)
	w.record(s, refs, func() {
		var seeds []xlKey
		for _, ref := range refs {
			if s.setRawQuiet(ref, "") {
				seeds = append(seeds, xlKey{s, ref})
			}
		}
		w.recalc(seeds)
	})
}

// ---- internal rich clipboard ----

// ClipCell is one copied cell.
type ClipCell struct {
	Raw      string
	Val      Value
	Style    Style
	HasStyle bool
}

// ClipBuf is an internal clipboard buffer that keeps formulas and styles.
type ClipBuf struct {
	Src   Range
	Sheet *Sheet
	Cells [][]ClipCell
	Cut   bool
}

// CopyRich captures r with formulas, values and styles.
func (s *Sheet) CopyRich(r Range) *ClipBuf {
	r = r.clip()
	b := &ClipBuf{Src: r, Sheet: s}
	for row := r.Min.Row; row <= r.Max.Row; row++ {
		line := make([]ClipCell, 0, r.Cols())
		for col := r.Min.Col; col <= r.Max.Col; col++ {
			st, ok := s.styles[CellRef{row, col}]
			line = append(line, ClipCell{Raw: s.Raw(row, col), Val: s.Value(row, col), Style: st, HasStyle: ok})
		}
		b.Cells = append(b.Cells, line)
	}
	return b
}

// PasteMode selects what Paste Special writes.
type PasteMode int

const (
	PasteAll PasteMode = iota
	PasteValues
	PasteFormulas
	PasteFormats
)

// xlValueRaw renders a value as typed text that reads back as the same value.
func xlValueRaw(v Value) string {
	switch v.Kind {
	case KindNumber:
		return strconv.FormatFloat(v.Num, 'g', -1, 64)
	case KindString:
		if xlInfer(v.Str).val.Kind != KindString || strings.HasPrefix(v.Str, "'") || strings.HasPrefix(v.Str, "=") {
			return "'" + v.Str
		}
		return v.Str
	case KindBool:
		return v.String()
	case KindError:
		return v.Err
	}
	return ""
}

// PasteSpecial pastes buf with its top-left at `at`. Relative references in
// formulas shift by the move distance. A cut buffer clears its source first.
// Returns the destination range.
func (w *Workbook) PasteSpecial(s *Sheet, at CellRef, buf *ClipBuf, mode PasteMode, skipBlanks, transpose bool) Range {
	if buf == nil || len(buf.Cells) == 0 {
		return Range{at, at}
	}
	h, wd := len(buf.Cells), len(buf.Cells[0])
	dh, dw := h, wd
	if transpose {
		dh, dw = wd, h
	}
	if buf.Cut && mode == PasteAll && !transpose && !skipBlanks {
		return w.moveCut(s, at, buf)
	}
	dst := Range{at, CellRef{at.Row + dh - 1, at.Col + dw - 1}}.clip()
	refs := xlRangeRefs(dst)
	if buf.Cut && buf.Sheet == s {
		refs = append(refs, xlRangeRefs(buf.Src)...)
	}
	if buf.Cut && buf.Sheet != s {
		w.ClearRange(buf.Sheet, buf.Src)
	}
	w.record(s, refs, func() {
		if buf.Cut && buf.Sheet == s {
			for _, ref := range xlRangeRefs(buf.Src) {
				s.setRawQuiet(ref, "")
				delete(s.styles, ref)
			}
		}
		var seeds []xlKey
		for i := 0; i < h; i++ {
			for j := 0; j < wd; j++ {
				cell := buf.Cells[i][j]
				ref := CellRef{at.Row + i, at.Col + j}
				if transpose {
					ref = CellRef{at.Row + j, at.Col + i}
				}
				if !dst.Contains(ref) {
					continue
				}
				if skipBlanks && cell.Raw == "" {
					continue
				}
				if mode != PasteFormats {
					raw := cell.Raw
					isF := len(raw) > 1 && raw[0] == '='
					switch {
					case isF && mode == PasteValues:
						raw = xlValueRaw(cell.Val)
					case isF:
						src := CellRef{buf.Src.Min.Row + i, buf.Src.Min.Col + j}
						raw = ShiftFormula(raw, ref.Col-src.Col, ref.Row-src.Row)
					}
					if s.setRawQuiet(ref, raw) {
						seeds = append(seeds, xlKey{s, ref})
					}
				}
				if mode == PasteAll || mode == PasteFormats {
					if cell.HasStyle {
						s.styles[ref] = cell.Style
					} else {
						delete(s.styles, ref)
					}
				}
			}
		}
		s.ver++
		w.recalc(seeds)
	})
	return dst
}

// moveCut moves a cut block to at. Unlike a copy, the moved formulas keep
// their references, and every formula in the workbook that pointed into the
// block follows it to the new place. The whole move is one undo step.
func (w *Workbook) moveCut(dest *Sheet, at CellRef, buf *ClipBuf) Range {
	src, from := buf.Src, buf.Sheet
	dst := Range{at, CellRef{at.Row + src.Rows() - 1, at.Col + src.Cols() - 1}}.clip()
	dr, dc := at.Row-src.Min.Row, at.Col-src.Min.Col
	if dr == 0 && dc == 0 && from == dest {
		return dst
	}
	w.undoable(func() bool {
		inside := func(p xlRefPart) bool { return src.Contains(CellRef{p.Row, p.Col}) }
		for _, sh := range w.Sheets {
			for _, c := range sh.cells {
				if len(c.raw) < 2 || c.raw[0] != '=' {
					continue
				}
				body := c.raw[1:]
				var sb strings.Builder
				last := 0
				for _, sp := range xlScanRefs(body) {
					target := sh
					if sp.hasSheet {
						target = w.sheetByName(sp.sheet)
					}
					if target != from || !inside(sp.a) || !inside(sp.b) {
						continue
					}
					a, b := sp.a, sp.b
					a.Row, a.Col, b.Row, b.Col = a.Row+dr, a.Col+dc, b.Row+dr, b.Col+dc
					if a.Row >= MaxRows || a.Col >= MaxCols || b.Row >= MaxRows || b.Col >= MaxCols {
						continue
					}
					sb.WriteString(body[last:sp.from])
					if sh != dest {
						sb.WriteString(xlQuoteSheet(dest.Name) + "!")
					}
					sb.WriteString(xlSpanText(sp, a, b))
					last = sp.to
				}
				if last > 0 || sb.Len() > 0 {
					sb.WriteString(body[last:])
					c.raw = "=" + sb.String()
				}
			}
		}
		type moved struct {
			raw   string
			st    Style
			hasSt bool
			note  string
		}
		var cells []moved
		for _, ref := range xlRangeRefs(src) {
			m := moved{raw: from.Raw(ref.Row, ref.Col), note: from.comments[ref]}
			m.st, m.hasSt = from.styles[ref]
			cells = append(cells, m)
			from.setRawQuiet(ref, "")
			delete(from.styles, ref)
			delete(from.comments, ref)
		}
		for i, ref := range xlRangeRefs(src) {
			to := CellRef{ref.Row + dr, ref.Col + dc}
			if !dst.Contains(to) {
				continue
			}
			m := cells[i]
			dest.setRawQuiet(to, m.raw)
			delete(dest.styles, to)
			if m.hasSt {
				dest.styles[to] = m.st
			}
			if m.note != "" {
				dest.comments[to] = m.note
			}
		}
		w.reloadAll()
		return true
	})
	return dst
}

// Paste pastes buf in full (formulas shifted, styles kept).
func (w *Workbook) Paste(s *Sheet, at CellRef, buf *ClipBuf) Range {
	return w.PasteSpecial(s, at, buf, PasteAll, false, false)
}

// ---- AutoFill ----

// xlFillMinSeeds is how many numeric seeds define a linear fill step.
const xlFillMinSeeds = 2

// xlNameLists returns the month and weekday name series AutoFill continues.
func xlNameLists() [][]string {
	return [][]string{
		{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"},
		{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"},
		{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"},
		{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
	}
}

func xlListIndex(list []string, s string) int {
	for i, v := range list {
		if strings.EqualFold(v, s) {
			return i
		}
	}
	return -1
}

// xlCaseLike matches the capitalization of tpl.
func xlCaseLike(tpl, s string) string {
	switch {
	case tpl != "" && tpl == strings.ToUpper(tpl):
		return strings.ToUpper(s)
	case tpl != "" && tpl == strings.ToLower(tpl):
		return strings.ToLower(s)
	}
	return s
}

// xlSeriesGen returns a generator for the k-th (1-based, outward) value after a
// seed vector, or nil when the seeds do not form a series.
func xlSeriesGen(seeds []xlCell, styles []Style) func(k int) (string, bool) {
	n := len(seeds)
	for _, c := range seeds {
		if len(c.raw) > 0 && c.raw[0] == '=' {
			return nil
		}
	}
	allNum := true
	for _, c := range seeds {
		if c.val.Kind != KindNumber {
			allNum = false
		}
	}
	if allNum {
		isDate := styles[n-1].NumFmt == "Date"
		step := 0.0
		switch {
		case n >= xlFillMinSeeds:
			step = (seeds[n-1].val.Num - seeds[0].val.Num) / float64(n-1)
		case isDate:
			step = 1
		default:
			return nil
		}
		last := seeds[n-1].val.Num
		return func(k int) (string, bool) {
			return strconv.FormatFloat(last+step*float64(k), 'f', -1, 64), true
		}
	}
	for _, list := range xlNameLists() {
		idx := make([]int, n)
		ok := true
		for i, c := range seeds {
			idx[i] = xlListIndex(list, c.raw)
			if idx[i] < 0 {
				ok = false
			}
		}
		if ok {
			step := 1
			if n >= xlFillMinSeeds {
				step = (idx[n-1] - idx[0] + len(list)) % len(list) / (n - 1)
				if step == 0 {
					step = 1
				}
			}
			tpl := seeds[n-1].raw
			return func(k int) (string, bool) {
				i := ((idx[n-1]+step*k)%len(list) + len(list)) % len(list)
				return xlCaseLike(tpl, list[i]), true
			}
		}
	}
	// text with a trailing number: "Item 1", "Q3"
	type tn struct {
		prefix string
		num    int
		width  int
	}
	parts := make([]tn, n)
	for i, c := range seeds {
		raw := c.raw
		j := len(raw)
		for j > 0 && raw[j-1] >= '0' && raw[j-1] <= '9' {
			j--
		}
		if j == len(raw) || j == 0 {
			return nil
		}
		v, err := strconv.Atoi(raw[j:])
		if err != nil {
			return nil
		}
		parts[i] = tn{raw[:j], v, len(raw) - j}
		if i > 0 && parts[i].prefix != parts[0].prefix {
			return nil
		}
	}
	step := 1
	if n >= xlFillMinSeeds {
		step = (parts[n-1].num - parts[0].num) / (n - 1)
	}
	return func(k int) (string, bool) {
		v := parts[n-1].num + step*k
		if v < 0 {
			return "", false
		}
		s := strconv.Itoa(v)
		for len(s) < parts[n-1].width {
			s = "0" + s
		}
		return parts[0].prefix + s, true
	}
}

// AutoFill extends the seed range src to cover `to` (a cell outside src in one
// straight direction) and returns the newly filled range. Numbers, dates, month
// and weekday names and "Item 1" style text continue as series; formulas shift;
// everything else repeats. Recorded as one undo step.
func (w *Workbook) AutoFill(s *Sheet, src Range, to CellRef) Range {
	vertical := to.Row > src.Max.Row || to.Row < src.Min.Row
	horizontal := to.Col > src.Max.Col || to.Col < src.Min.Col
	if vertical == horizontal { // neither or diagonal
		return Range{}
	}
	var dst Range
	switch {
	case to.Row > src.Max.Row:
		dst = Range{CellRef{src.Max.Row + 1, src.Min.Col}, CellRef{to.Row, src.Max.Col}}
	case to.Row < src.Min.Row:
		dst = Range{CellRef{to.Row, src.Min.Col}, CellRef{src.Min.Row - 1, src.Max.Col}}
	case to.Col > src.Max.Col:
		dst = Range{CellRef{src.Min.Row, src.Max.Col + 1}, CellRef{src.Max.Row, to.Col}}
	default:
		dst = Range{CellRef{src.Min.Row, to.Col}, CellRef{src.Max.Row, src.Min.Col - 1}}
	}
	dst = dst.clip()
	if dst.Min.Row > dst.Max.Row || dst.Min.Col > dst.Max.Col {
		return Range{}
	}
	forward := to.Row > src.Max.Row || to.Col > src.Max.Col
	w.record(s, xlRangeRefs(dst), func() {
		var seeds []xlKey
		lines := src.Cols()
		if !vertical {
			lines = src.Rows()
		}
		for line := 0; line < lines; line++ {
			// seed cells along the fill axis, ordered outward from the source edge
			var refs []CellRef
			n := src.Rows()
			if !vertical {
				n = src.Cols()
			}
			for i := 0; i < n; i++ {
				idx := i
				if !forward {
					idx = n - 1 - i
				}
				if vertical {
					refs = append(refs, CellRef{src.Min.Row + idx, src.Min.Col + line})
				} else {
					refs = append(refs, CellRef{src.Min.Row + line, src.Min.Col + idx})
				}
			}
			// refs is in natural order for forward; for backward it runs from the far edge inward,
			// so reverse it back to read as a forward series moving outward.
			ordered := make([]xlCell, 0, n)
			sts := make([]Style, 0, n)
			for _, ref := range refs {
				c := s.cells[ref]
				if c == nil {
					c = &xlCell{}
				}
				ordered = append(ordered, *c)
				sts = append(sts, s.styles[ref])
			}
			gen := xlSeriesGen(ordered, sts)
			count := dst.Rows()
			if !vertical {
				count = dst.Cols()
			}
			for k := 1; k <= count; k++ {
				var ref CellRef
				switch {
				case vertical && forward:
					ref = CellRef{src.Max.Row + k, src.Min.Col + line}
				case vertical:
					ref = CellRef{src.Min.Row - k, src.Min.Col + line}
				case forward:
					ref = CellRef{src.Min.Row + line, src.Max.Col + k}
				default:
					ref = CellRef{src.Min.Row + line, src.Min.Col - k}
				}
				if !dst.Contains(ref) {
					continue
				}
				// source cell by cycling the seeds
				si := (k - 1) % n
				var from CellRef
				if forward {
					if vertical {
						from = CellRef{src.Min.Row + si, ref.Col}
					} else {
						from = CellRef{ref.Row, src.Min.Col + si}
					}
				} else {
					if vertical {
						from = CellRef{src.Max.Row - si, ref.Col}
					} else {
						from = CellRef{ref.Row, src.Max.Col - si}
					}
				}
				raw := s.Raw(from.Row, from.Col)
				if gen != nil {
					if g, ok := gen(k); ok {
						raw = g
					}
				} else if len(raw) > 1 && raw[0] == '=' {
					raw = ShiftFormula(raw, ref.Col-from.Col, ref.Row-from.Row)
				}
				if s.setRawQuiet(ref, raw) {
					seeds = append(seeds, xlKey{s, ref})
				}
				if st, ok := s.styles[from]; ok {
					s.styles[ref] = st
				} else {
					delete(s.styles, ref)
				}
			}
		}
		s.ver++
		w.recalc(seeds)
	})
	return dst
}

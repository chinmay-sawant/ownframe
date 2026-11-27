package main

const historyLimit = 200

// xlCmd is one reversible step.
type xlCmd struct{ undo, redo func() }

type xlHistory struct {
	undo, redo []xlCmd
}

func (w *Workbook) push(c xlCmd) {
	w.hist.undo = append(w.hist.undo, c)
	if len(w.hist.undo) > historyLimit {
		w.hist.undo = append([]xlCmd(nil), w.hist.undo[len(w.hist.undo)-historyLimit:]...)
	}
	w.hist.redo = nil
}

func (w *Workbook) CanUndo() bool { return len(w.hist.undo) > 0 }
func (w *Workbook) CanRedo() bool { return len(w.hist.redo) > 0 }

// ClearHistory drops all undo and redo steps (after New/Open).
func (w *Workbook) ClearHistory() { w.hist = xlHistory{} }

func (w *Workbook) Undo() bool {
	n := len(w.hist.undo)
	if n == 0 {
		return false
	}
	c := w.hist.undo[n-1]
	w.hist.undo = w.hist.undo[:n-1]
	c.undo()
	w.hist.redo = append(w.hist.redo, c)
	return true
}

func (w *Workbook) Redo() bool {
	n := len(w.hist.redo)
	if n == 0 {
		return false
	}
	c := w.hist.redo[n-1]
	w.hist.redo = w.hist.redo[:n-1]
	c.redo()
	w.hist.undo = append(w.hist.undo, c)
	return true
}

// xlSnap captures one cell's text and style.
type xlSnap struct {
	raw   string
	st    Style
	hasSt bool
}

func (s *Sheet) snap(ref CellRef) xlSnap {
	st, ok := s.styles[ref]
	return xlSnap{raw: s.Raw(ref.Row, ref.Col), st: st, hasSt: ok}
}

// restoreQuiet puts a snapshot back; it reports whether the text changed.
func (s *Sheet) restoreQuiet(ref CellRef, n xlSnap) bool {
	changed := s.setRawQuiet(ref, n.raw)
	if n.hasSt {
		s.styles[ref] = n.st
	} else {
		delete(s.styles, ref)
	}
	s.ver++
	return changed
}

func (w *Workbook) applySnaps(s *Sheet, refs []CellRef, snaps []xlSnap) {
	var seeds []xlKey
	for i, ref := range refs {
		if s.restoreQuiet(ref, snaps[i]) {
			seeds = append(seeds, xlKey{s, ref})
		}
	}
	w.recalc(seeds)
}

// record runs mutate (which edits refs of s) and stores an undo step.
func (w *Workbook) record(s *Sheet, refs []CellRef, mutate func()) {
	before := make([]xlSnap, len(refs))
	for i, ref := range refs {
		before[i] = s.snap(ref)
	}
	mutate()
	after := make([]xlSnap, len(refs))
	changed := false
	for i, ref := range refs {
		after[i] = s.snap(ref)
		if after[i] != before[i] {
			changed = true
		}
	}
	if !changed {
		return
	}
	w.push(xlCmd{
		undo: func() { w.applySnaps(s, refs, before) },
		redo: func() { w.applySnaps(s, refs, after) },
	})
}

func xlRangeRefs(r Range) []CellRef {
	r = r.clip()
	var refs []CellRef
	for row := r.Min.Row; row <= r.Max.Row; row++ {
		for col := r.Min.Col; col <= r.Max.Col; col++ {
			refs = append(refs, CellRef{row, col})
		}
	}
	return refs
}

// Edit sets one cell and records it for undo.
func (w *Workbook) Edit(sheet *Sheet, row, col int, raw string) {
	ref := CellRef{row, col}
	w.record(sheet, []CellRef{ref}, func() { sheet.SetRaw(row, col, raw) })
}

// EditRange writes a block of typed text starting at r.Min as one undo step. A
// 1x1 block fills the whole range.
func (w *Workbook) EditRange(sheet *Sheet, r Range, raw [][]string) {
	if len(raw) == 0 {
		return
	}
	target := r
	if len(raw) != 1 || len(raw[0]) != 1 {
		h, wd := len(raw), 0
		for _, row := range raw {
			wd = max(wd, len(row))
		}
		target = Range{r.Min, CellRef{r.Min.Row + h - 1, r.Min.Col + wd - 1}}
	}
	refs := xlRangeRefs(target)
	w.record(sheet, refs, func() {
		var seeds []xlKey
		for _, ref := range refs {
			i, j := ref.Row-target.Min.Row, ref.Col-target.Min.Col
			if len(raw) == 1 && len(raw[0]) == 1 {
				i, j = 0, 0
			}
			if i >= len(raw) || j >= len(raw[i]) {
				continue
			}
			if sheet.setRawQuiet(ref, raw[i][j]) {
				seeds = append(seeds, xlKey{sheet, ref})
			}
		}
		w.recalc(seeds)
	})
}

// Format applies a style change to r and records it for undo.
func (w *Workbook) Format(sheet *Sheet, r Range, f func(*Style)) {
	w.record(sheet, xlRangeRefs(r), func() { sheet.SetStyle(r, f) })
}

// SortRange sorts r by keyCol as one undo step.
func (w *Workbook) SortRange(s *Sheet, r Range, keyCol int, asc, hasHeader bool) {
	before := w.captureState()
	s.Sort(r, keyCol, asc, hasHeader)
	after := w.captureState()
	w.push(xlCmd{undo: func() { w.restoreState(before) }, redo: func() { w.restoreState(after) }})
}

package main

import "strings"

// xlKey identifies one cell in the workbook.
type xlKey struct {
	sh  *Sheet
	ref CellRef
}

// xlCell is one stored cell: typed text, its value, and (for formulas) the AST
// and the cells it reads.
type xlCell struct {
	raw  string
	val  Value
	ast  xlNode
	deps []xlKey
}

func (w *Workbook) sheetByName(name string) *Sheet {
	for _, s := range w.Sheets {
		if strings.EqualFold(s.Name, name) {
			return s
		}
	}
	return nil
}

// collectDeps lists the distinct cells a formula reads.
func (w *Workbook) collectDeps(cur *Sheet, n xlNode) []xlKey {
	seen := map[xlKey]bool{}
	var out []xlKey
	add := func(sh *Sheet, r Range) {
		if sh == nil {
			return
		}
		r = r.clip()
		for row := r.Min.Row; row <= r.Max.Row; row++ {
			for col := r.Min.Col; col <= r.Max.Col; col++ {
				k := xlKey{sh, CellRef{row, col}}
				if !seen[k] {
					seen[k] = true
					out = append(out, k)
				}
			}
		}
	}
	pick := func(name string, has bool) *Sheet {
		if !has {
			return cur
		}
		return w.sheetByName(name)
	}
	var walk func(n xlNode)
	walk = func(n xlNode) {
		switch t := n.(type) {
		case nRef:
			add(pick(t.sheet, t.hasSheet), Range{t.ref, t.ref})
		case nRange:
			add(pick(t.sheet, t.hasSheet), t.rng)
		case nUnary:
			walk(t.x)
		case nPercent:
			walk(t.x)
		case nBinary:
			walk(t.l)
			walk(t.r)
		case nCall:
			for _, a := range t.args {
				walk(a)
			}
		}
	}
	walk(n)
	return out
}

func (w *Workbook) addEdges(k xlKey, deps []xlKey) {
	if w.rev == nil {
		w.rev = map[xlKey]map[xlKey]struct{}{}
	}
	for _, d := range deps {
		m := w.rev[d]
		if m == nil {
			m = map[xlKey]struct{}{}
			w.rev[d] = m
		}
		m[k] = struct{}{}
	}
}

func (w *Workbook) dropEdges(k xlKey, deps []xlKey) {
	for _, d := range deps {
		if m := w.rev[d]; m != nil {
			delete(m, k)
			if len(m) == 0 {
				delete(w.rev, d)
			}
		}
	}
}

// recalc re-evaluates the given cells and everything that depends on them, in
// topological order. Cells caught in a cycle (or fed by one) become #CIRC!.
func (w *Workbook) recalc(seeds []xlKey) {
	dirty := map[xlKey]*xlCell{}
	seen := map[xlKey]bool{}
	queue := append([]xlKey(nil), seeds...)
	for _, s := range seeds {
		seen[s] = true
	}
	for len(queue) > 0 {
		k := queue[0]
		queue = queue[1:]
		if c := k.sh.cells[k.ref]; c != nil && c.ast != nil {
			dirty[k] = c
		}
		for d := range w.rev[k] {
			if !seen[d] {
				seen[d] = true
				queue = append(queue, d)
			}
		}
	}
	indeg := make(map[xlKey]int, len(dirty))
	var ready []xlKey
	for k, c := range dirty {
		n := 0
		for _, d := range c.deps {
			if _, ok := dirty[d]; ok {
				n++
			}
		}
		indeg[k] = n
		if n == 0 {
			ready = append(ready, k)
		}
	}
	done := 0
	for len(ready) > 0 {
		k := ready[len(ready)-1]
		ready = ready[:len(ready)-1]
		done++
		w.evalCell(k, dirty[k])
		for d := range w.rev[k] {
			if _, ok := dirty[d]; ok {
				indeg[d]--
				if indeg[d] == 0 {
					ready = append(ready, d)
				}
			}
		}
	}
	if done < len(dirty) {
		for k, c := range dirty {
			if indeg[k] > 0 {
				c.val = xlErrV(ErrCirc)
			}
		}
	}
	for k := range dirty {
		k.sh.ver++
	}
}

func (w *Workbook) evalCell(k xlKey, c *xlCell) {
	ctx := &xlCtx{wb: w, sh: k.sh, at: k.ref}
	v := ctx.scalar(ctx.eval(c.ast))
	if v.Kind == KindEmpty {
		v = xlNumV(0)
	}
	c.val = v
}

// reloadAll re-parses every formula from its raw text, rebuilds dependency
// edges, and recalculates the whole workbook.
func (w *Workbook) reloadAll() {
	w.rev = map[xlKey]map[xlKey]struct{}{}
	var seeds []xlKey
	for _, s := range w.Sheets {
		for ref, c := range s.cells {
			if len(c.raw) > 1 && c.raw[0] == '=' {
				ast, err := xlParse(c.raw[1:])
				if err != nil {
					c.ast, c.deps, c.val = nil, nil, xlErrV(ErrSyn)
					continue
				}
				k := xlKey{s, ref}
				c.ast = ast
				c.deps = w.collectDeps(s, ast)
				w.addEdges(k, c.deps)
				seeds = append(seeds, k)
			}
		}
		s.ver++
	}
	w.recalc(seeds)
}

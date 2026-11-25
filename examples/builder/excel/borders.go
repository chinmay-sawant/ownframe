package main

// BorderPreset names one entry of the Borders menu.
const (
	BorderBottom  = "bottom"
	BorderTop     = "top"
	BorderLeft    = "left"
	BorderRight   = "right"
	BorderNone    = "none"
	BorderAll     = "all"
	BorderOutside = "outside"
	BorderThick   = "thick"
	BorderTotal   = "total" // thin top and double bottom, the accounting total look
)

// edgesFor returns the edge settings a preset puts on the cell at ref when the
// preset is applied to rng; keep reports which edges the preset sets.
func edgesFor(preset string, rng Range, ref CellRef) (set Edges, touched Edges) {
	top, bot := ref.Row == rng.Min.Row, ref.Row == rng.Max.Row
	left, right := ref.Col == rng.Min.Col, ref.Col == rng.Max.Col
	mark := func(on bool, dst, mk *string, v string) {
		if on {
			*dst, *mk = v, "x"
		}
	}
	switch preset {
	case BorderBottom:
		mark(bot, &set.Bottom, &touched.Bottom, EdgeThin)
	case BorderTop:
		mark(top, &set.Top, &touched.Top, EdgeThin)
	case BorderLeft:
		mark(left, &set.Left, &touched.Left, EdgeThin)
	case BorderRight:
		mark(right, &set.Right, &touched.Right, EdgeThin)
	case BorderAll:
		set = Edges{EdgeThin, EdgeThin, EdgeThin, EdgeThin}
		touched = Edges{"x", "x", "x", "x"}
	case BorderOutside, BorderThick:
		v := EdgeThin
		if preset == BorderThick {
			v = EdgeThick
		}
		mark(left, &set.Left, &touched.Left, v)
		mark(right, &set.Right, &touched.Right, v)
		mark(top, &set.Top, &touched.Top, v)
		mark(bot, &set.Bottom, &touched.Bottom, v)
	case BorderTotal:
		mark(top, &set.Top, &touched.Top, EdgeThin)
		mark(bot, &set.Bottom, &touched.Bottom, EdgeDouble)
	}
	return set, touched
}

// ApplyBorder applies a Borders-menu preset to r as one undo step. "none"
// clears every edge. Other presets set only the edges they name.
func (w *Workbook) ApplyBorder(s *Sheet, r Range, preset string) {
	r = r.clip()
	w.record(s, xlRangeRefs(r), func() {
		for row := r.Min.Row; row <= r.Max.Row; row++ {
			for col := r.Min.Col; col <= r.Max.Col; col++ {
				ref := CellRef{row, col}
				st := s.styles[ref].migrated()
				if preset == BorderNone {
					st.Edges = Edges{}
				} else {
					set, touched := edgesFor(preset, r, ref)
					pick := func(dst *string, mk, v string) {
						if mk != "" {
							*dst = v
						}
					}
					pick(&st.Edges.Left, touched.Left, set.Left)
					pick(&st.Edges.Right, touched.Right, set.Right)
					pick(&st.Edges.Top, touched.Top, set.Top)
					pick(&st.Edges.Bottom, touched.Bottom, set.Bottom)
				}
				if st == (Style{}) {
					delete(s.styles, ref)
				} else {
					s.styles[ref] = st
				}
			}
		}
		s.ver++
	})
}

package main

// walkTrace is the click state snapshot taken on the game goroutine. The HTTP
// endpoint reads this instead of the live view and page stats, so a poll can
// never race the click handler.
type walkTrace struct {
	started   int
	selected  string
	last      int64
	redraws   uint64
	parses    uint64
	cascades  uint64
	layouts   uint64
	repaints  uint64
	ops       int64
	boxes     int64
	tmpl      int64
	style     int64
	display   int64
	paint     int64
	changed   int64
	dirty     int64
	alloc     uint64
	startedNS int64
	seq       uint64
}

// snapshot records the counters the click script polls. The caller holds w.mu.
func (w *walker) snapshot() {
	started := 0
	for _, row := range w.v.Catalog {
		if w.v.Started[row.Button] {
			started++
		}
	}
	selected := w.v.Selected
	if selected == "" {
		selected = "-"
	}
	st := w.p.Testing().Stats()
	w.trace = walkTrace{
		started: started, selected: selected,
		last: st.LastRedraw.Microseconds(), redraws: st.Redraws,
		parses: st.Parses, cascades: st.Cascades,
		layouts: st.Layouts, repaints: st.Repaints,
		ops: int64(st.Ops), boxes: int64(st.Boxes),
		tmpl: st.LastTemplate.Microseconds(), style: st.LayoutTime.Microseconds(),
		display: st.DisplayListTime.Microseconds(), paint: st.PaintTime.Microseconds(),
		changed: int64(st.ChangedOps), dirty: int64(st.DirtyOps), alloc: st.AllocFrame,
		startedNS: w.d.startedAt.UnixNano(), seq: w.d.startedSeq,
	}
}

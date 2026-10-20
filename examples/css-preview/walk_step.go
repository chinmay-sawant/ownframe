package main

func (w *walker) step() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.resetReq {
		w.resetReq = false
		w.reset()
		if w.resetDone != nil {
			close(w.resetDone)
			w.resetDone = nil
		}
	}
	if w.stopReq {
		w.stopReq = false
		w.d.samples = map[string]*sample{}
		w.d.pendingKey = ""
	}
	w.snapshot()
	if w.next < 0 || w.next >= len(w.v.Catalog) {
		return
	}
	row := w.v.Catalog[w.next]
	if w.m != nil && w.scrollFor == w.next {
		_, viewH := w.p.Size()
		y := w.m.rowTop[w.next/2] - 200
		if maxY := w.m.contentH - viewH; y > maxY {
			y = maxY
		}
		if y < 0 {
			y = 0
		}
		w.p.ScrollTo(0, y)
		w.scrollFor = -1

		return
	}

	b, ok := findBox(w.p.Boxes(), row.Button)
	if !ok {
		return
	}
	if w.m == nil && w.scrollFor == w.next {
		w.p.ScrollTo(0, int(b.Y)-200)
		w.scrollFor = -1

		return
	}
	_, scrollY := w.p.ScrollOffset()
	clientY := int(b.Y+b.H/2) - scrollY
	_, viewH := w.p.Size()
	if clientY < 0 || clientY > viewH {
		return
	}
	w.aim = [3]int{w.next, int(b.X + b.W/2), clientY}
}

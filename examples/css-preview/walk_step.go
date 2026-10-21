package main

// publishBoxes copies the page boxes on the game goroutine and keeps them
// for the HTTP /box route. The route reads this snapshot under the walker
// mutex, so it never reads the page while the loop swaps the boxes. The
// page bumps its generation on every redraw and on every element patch,
// which are the two paths that replace the boxes.
func (w *walker) publishBoxes() {
	gen := w.p.Generation()
	if gen == w.boxGen {
		return
	}
	w.boxGen = gen
	w.boxes = w.p.Boxes()
}

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
		w.d.pending = nil
	}
	w.snapshot()
	w.publishBoxes()
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

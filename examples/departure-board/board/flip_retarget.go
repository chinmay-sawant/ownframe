package board

// retarget starts the row bands on the tiles whose target glyph changed. It
// shares the band math with rebind and never asks the page for a new layout:
// each flap keeps the op it already holds, and the generation stays put so a
// later real Redraw still rebinds.
func (f *flipper) retarget(rows []Row) {
	for r, row := range rows {
		for k, c := range row.Cells {
			fl := f.cells[c.ID]
			if fl == nil || fl.want == c.Ch {
				continue
			}

			fl.want = c.Ch
			fl.delay = r * rowDelay
			fl.spin = flapSpin(r, k)
			fl.done = false
		}
	}
}

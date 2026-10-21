package main

import (
	"context"
	"time"
)

// resizeSettle is how long the window size has to hold before the catalog is
// measured again. A drag resizes many times, and one measurement per size
// would cost more than it is worth.
const resizeSettle = 300 * time.Millisecond

// remeasureWait is the delay between the checks for a settled window size.
const remeasureWait = 100 * time.Millisecond

// checkSize notices a resize. The catalog geometry was measured at the old
// size, so the rows and the scroll range have to be measured again.
func (d *demo) checkSize() {
	w, h := d.p.Size()
	if w == d.sizeW && h == d.sizeH {
		return
	}
	d.sizeW, d.sizeH = w, h
	d.remeasure = true
	d.resizeAt = time.Now()
	d.wake()
}

// remeasureCatalog lays the whole catalog out again and measures its rows,
// so a resize gets fresh geometry. The windowed rows come back after it.
func (d *demo) remeasureCatalog(ctx context.Context) error {
	w := d.window
	if w == nil {
		return nil
	}
	p, v := d.p, d.v
	rows, top, bot := v.Rows, v.TopPad, v.BotPad
	p.SetScrollWindow(nil)
	v.Rows, v.TopPad, v.BotPad = v.Catalog, 0, 0
	p.SetData(v)
	if err := p.Redraw(ctx); err != nil {
		return err
	}
	m, ok := measureCatalog(p)
	p.SetScrollWindow(w.slice)
	v.Rows, v.TopPad, v.BotPad = rows, top, bot
	p.SetData(v)
	if ok {
		w.metrics = m
		w.prevStart, w.prevEnd = -1, -1
	}
	return p.Redraw(ctx)
}

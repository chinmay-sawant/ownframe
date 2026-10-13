package page

import "sync"

// DisplayPatch reports how one display update changed retained ops.
type DisplayPatch struct {
	Changed int
	Added   int
	Removed int
	Regions int
	Full    bool
}

var displayOpPool = sync.Pool{
	New: func() any {
		return make([]DisplayOp, 0, 64)
	},
}

func (p *Page) rentOps(n int) []DisplayOp {
	buf := displayOpPool.Get().([]DisplayOp)
	if cap(buf) < n {
		buf = make([]DisplayOp, 0, n)
	}

	return buf[:0]
}

func (p *Page) freeOps(buf []DisplayOp) {
	displayOpPool.Put(buf[:0])
}

func (p *Page) PatchReport() DisplayPatch {
	return p.lastPatch
}

func (p *Page) LastPatch() DisplayPatch {
	return p.lastPatch
}

func (p *Page) DisplayPatch() DisplayPatch {
	return p.lastPatch
}

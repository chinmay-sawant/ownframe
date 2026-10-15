package page

import "sync"

// DisplayPatch reports how one display update changed retained ops.
// ChangedProperties is the canonical property set; use it in new code.
// Properties, Props, and ChangedProps mirror it for compat; do not mutate.
type DisplayPatch struct {
	Changed int
	Added   int
	Removed int
	Regions int
	Full    bool

	// Deprecated: use ChangedProperties.
	Properties []string
	// Deprecated: use ChangedProperties.
	Props             []string
	ChangedProperties []string
	// Deprecated: use ChangedProperties.
	ChangedProps []string
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

// LastPatch returns the last patch report; use it.
func (p *Page) LastPatch() DisplayPatch {
	return p.lastPatch
}

// Deprecated: use LastPatch.
func (p *Page) PatchReport() DisplayPatch {
	return p.LastPatch()
}

// Deprecated: use LastPatch.
func (p *Page) DisplayPatch() DisplayPatch {
	return p.LastPatch()
}

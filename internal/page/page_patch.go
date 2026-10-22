package page

import "sync"

// displayOpCap is the pooled buffer capacity for one display update.
const displayOpCap = 64

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
		b := make([]DisplayOp, 0, displayOpCap)
		return &b
	},
}

func (p *Page) rentOps(n int) []DisplayOp {
	bp, ok := displayOpPool.Get().(*[]DisplayOp)
	if !ok {
		return make([]DisplayOp, 0, n)
	}
	buf := *bp
	if cap(buf) < n {
		buf = make([]DisplayOp, 0, n)
	}

	return buf[:0]
}

func (p *Page) freeOps(buf []DisplayOp) {
	b := buf[:0]
	displayOpPool.Put(&b)
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

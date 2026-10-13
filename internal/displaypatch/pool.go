package displaypatch

// Pool recycles freed operation slots.
type Pool struct {
	free []int
}

// NewPool returns an empty pool.
func NewPool() *Pool { return &Pool{} }

// Release records freed operation IDs for reuse.
func (p *Pool) Release(ids ...int) {
	if p == nil {
		return
	}
	p.free = append(p.free, ids...)
}

// Acquire returns a recycled ID or -1 if none.
func (p *Pool) Acquire() int {
	if p == nil || len(p.free) == 0 {
		return -1
	}
	id := p.free[len(p.free)-1]
	p.free = p.free[:len(p.free)-1]
	return id
}

// Size reports recycled slots.
func (p *Pool) Size() int {
	if p == nil {
		return 0
	}
	return len(p.free)
}

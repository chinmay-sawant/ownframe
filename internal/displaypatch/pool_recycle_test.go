package displaypatch

import "testing"

func TestPoolRecycleMultiple(t *testing.T) {
	p := NewPool()
	p.Release(4, 5)
	if got := p.Acquire(); got != 5 {
		t.Fatalf("got %d", got)
	}
	if got := p.Acquire(); got != 4 {
		t.Fatalf("got %d", got)
	}
	if got := p.Acquire(); got != -1 {
		t.Fatalf("got %d", got)
	}
	if p.Size() != 0 {
		t.Fatalf("size %d", p.Size())
	}
}

func TestPoolNilSafety(t *testing.T) {
	var p *Pool
	p.Release(1)
	if got := p.Acquire(); got != -1 {
		t.Fatalf("got %d", got)
	}
	if p.Size() != 0 {
		t.Fatalf("size %d", p.Size())
	}
}

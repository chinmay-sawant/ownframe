package replay

import "testing"

func TestLRUOrderTouchAndEvict(t *testing.T) {
	t.Parallel()

	var order lruOrder[int]
	for _, k := range []int{1, 2, 3} {
		order.add(k)
	}
	order.touch(1)

	sizes := map[int]int{1: 100, 2: 200, 3: 300}
	pixels := 600
	var evicted []int
	pixels = order.evict(order.length(), 2, pixels, 450, 1, func(k int) int {
		evicted = append(evicted, k)
		px := sizes[k]
		delete(sizes, k)
		return px
	})
	if len(evicted) != 1 || evicted[0] != 2 {
		t.Fatalf("evicted = %v, want [2]", evicted)
	}
	if order.length() != 2 || pixels != 400 {
		t.Fatalf("after evict: len=%d pixels=%d, want 2 400", order.length(), pixels)
	}

	// Self-eviction guard: a single protected entry exceeding maxPixels stays.
	var single lruOrder[int]
	single.add(99)
	dropped := 0
	singlePx := single.evict(1, 1, 10000, 100, 99, func(int) int {
		dropped++
		return 10000
	})
	if dropped != 0 || single.length() != 1 || singlePx != 10000 {
		t.Fatalf("protected entry was evicted: dropped=%d len=%d px=%d", dropped, single.length(), singlePx)
	}
}

// TestLRUOrderTouchAllocsZero pins PERF-LRU-01: a hit on a full 256-entry
// text raster cache stamps one tick and allocates nothing.
func TestLRUOrderTouchAllocsZero(t *testing.T) {
	var order lruOrder[textRasterKey]
	keys := make([]textRasterKey, 256)
	for i := range keys {
		keys[i] = textRasterKey{content: "run", size: float64(i), font: "face"}
		order.add(keys[i])
	}
	i := 0
	allocs := testing.AllocsPerRun(1000, func() {
		order.touch(keys[i%len(keys)])
		i++
	})
	if allocs != 0 {
		t.Fatalf("touch allocs/op = %v, want 0", allocs)
	}
}

func BenchmarkLRUOrderTouch256(b *testing.B) {
	var order lruOrder[textRasterKey]
	keys := make([]textRasterKey, 256)
	for i := range keys {
		keys[i] = textRasterKey{content: "run", size: float64(i), font: "face"}
		order.add(keys[i])
	}
	b.ReportAllocs()
	for i := 0; b.Loop(); i++ {
		order.touch(keys[i%len(keys)])
	}
}

package replay

import (
	"slices"
	"testing"
)

func TestLRUTouchAndEvict(t *testing.T) {
	t.Parallel()

	order := []int{1, 2, 3}
	order = lruTouch(order, 1)
	if !slices.Equal(order, []int{2, 3, 1}) {
		t.Fatalf("order after touch(1) = %v, want [2 3 1]", order)
	}

	sizes := map[int]int{1: 100, 2: 200, 3: 300}
	pixels := 600
	var evicted []int
	order, pixels = lruEvict(order, len(order), 2, pixels, 450, 1, func(k int) int {
		evicted = append(evicted, k)
		px := sizes[k]
		delete(sizes, k)
		return px
	})
	if !slices.Equal(evicted, []int{2}) {
		t.Fatalf("evicted = %v, want [2]", evicted)
	}
	if !slices.Equal(order, []int{3, 1}) || pixels != 400 {
		t.Fatalf("after evict: order=%v pixels=%d, want [3 1] 400", order, pixels)
	}

	// Self-eviction guard: a single protected entry exceeding maxPixels stays.
	single := []int{99}
	singlePx := 10000
	dropped := 0
	single, singlePx = lruEvict(single, 1, 1, singlePx, 100, 99, func(int) int {
		dropped++
		return 10000
	})
	if dropped != 0 || !slices.Equal(single, []int{99}) || singlePx != 10000 {
		t.Fatalf("protected entry was evicted: dropped=%d order=%v px=%d", dropped, single, singlePx)
	}
}

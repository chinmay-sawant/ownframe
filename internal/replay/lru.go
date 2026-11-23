package replay

import "sync/atomic"

// lruOrder tracks recency for one cache with a monotonic access tick per key.
// A hit stores one atomic and moves no slice elements, so touching a key
// takes no allocation and does not copy pointer-bearing keys through the GC
// write barrier. Only eviction scans for the smallest tick. add and remove
// need the owning cache's write lock; touch only reads the map, so a cache
// may call it under a read lock.
type lruOrder[K comparable] struct {
	ticks map[K]*atomic.Uint64
	clock atomic.Uint64
}

// length reports how many keys are tracked.
func (o *lruOrder[K]) length() int { return len(o.ticks) }

// reset forgets every key.
func (o *lruOrder[K]) reset() { o.ticks = nil }

// reserve allocates the index for n keys when a park emptied it.
func (o *lruOrder[K]) reserve(n int) {
	if o.ticks == nil {
		o.ticks = make(map[K]*atomic.Uint64, n)
	}
}

// add tracks key as most recently used.
func (o *lruOrder[K]) add(key K) {
	if o.ticks == nil {
		o.ticks = make(map[K]*atomic.Uint64)
	}
	t := o.ticks[key]
	if t == nil {
		t = new(atomic.Uint64)
		o.ticks[key] = t
	}
	t.Store(o.clock.Add(1))
}

// touch marks key most recently used. Unknown keys are ignored.
func (o *lruOrder[K]) touch(key K) {
	if t := o.ticks[key]; t != nil {
		t.Store(o.clock.Add(1))
	}
}

// oldest returns the least recently used key other than protected.
func (o *lruOrder[K]) oldest(protected K) (K, bool) {
	var (
		victim K
		best   uint64
		found  bool
	)
	for k, t := range o.ticks {
		if k == protected {
			continue
		}
		if v := t.Load(); !found || v < best {
			victim, best, found = k, v, true
		}
	}
	return victim, found
}

// evict drops the oldest non-protected keys while count exceeds maxEntries or
// pixels exceeds maxPixels. drop removes the victim from the caller's map,
// deallocates its GPU resource, and returns the pixels freed. protected is
// never evicted so a newly inserted entry larger than maxPixels is never
// disposed before being returned to its caller. It returns the updated pixel
// total.
func (o *lruOrder[K]) evict(count, maxEntries, pixels, maxPixels int, protected K, drop func(K) int) int {
	for count > maxEntries || pixels > maxPixels {
		victim, ok := o.oldest(protected)
		if !ok {
			break
		}
		delete(o.ticks, victim)
		pixels -= drop(victim)
		count--
	}
	return pixels
}

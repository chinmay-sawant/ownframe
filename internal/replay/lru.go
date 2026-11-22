package replay

// lruTouch moves key to the back of order so the least-recently-used key stays
// at index 0.
func lruTouch[K comparable](order []K, key K) []K {
	for i, candidate := range order {
		if candidate != key {
			continue
		}
		copy(order[i:], order[i+1:])
		order[len(order)-1] = key
		return order
	}
	return order
}

// lruEvict drops the oldest non-protected keys from order while count exceeds
// maxEntries or pixels exceeds maxPixels. drop removes the victim from the
// caller's map, deallocates its GPU resource, and returns the pixels freed.
// protected is never evicted so a newly inserted entry larger than maxPixels
// is never disposed before being returned to its caller.
func lruEvict[K comparable](order []K, count, maxEntries, pixels, maxPixels int, protected K, drop func(K) int) ([]K, int) {
	for count > maxEntries || pixels > maxPixels {
		idx := -1
		for i, candidate := range order {
			if candidate == protected {
				continue
			}
			idx = i
			break
		}
		if idx < 0 {
			break
		}
		victim := order[idx]
		order = append(order[:idx], order[idx+1:]...)
		pixels -= drop(victim)
		count--
	}
	return order, pixels
}

package depthalert

func (d *displayFilter) view(scanning bool) view {
	return view{Scanning: scanning, State: d.state.String(),
		Distance: distance(d.nearView), Coverage: coverage((d.coverView + coverBucket/2) / coverBucket * coverBucket)}
}

func settle(old, next, deadband int32) int32 {
	if old == 0 || abs(old-next) >= deadband {
		return next
	}
	return old
}

func abs(n int32) int32 {
	if n < 0 {
		return -n
	}
	return n
}

func median(samples []int32, n int) int32 {
	var sorted [displaySamples]int32
	copy(sorted[:], samples[:n])
	for i := 1; i < n; i++ {
		for j := i; j > 0 && sorted[j] < sorted[j-1]; j-- {
			sorted[j], sorted[j-1] = sorted[j-1], sorted[j]
		}
	}
	return sorted[n/2]
}

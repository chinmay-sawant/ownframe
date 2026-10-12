package board

// destWidth and statWidth size the fixed grid: the widest destination in the
// seed and the longest string the status cycle can produce.
var destWidth, statWidth = widestDest(), widestStat()

// widestDest returns the longest seeded destination.
func widestDest() int {
	w := 0
	for _, d := range seed {
		if len(d.dest) > w {
			w = len(d.dest)
		}
	}

	return w
}

// widestStat returns the longest status in the cycle.
func widestStat() int {
	w := 0
	for _, s := range statusCycle {
		if len(s) > w {
			w = len(s)
		}
	}

	return w
}

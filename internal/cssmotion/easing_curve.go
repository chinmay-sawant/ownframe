package cssmotion

// halfDivisor splits a bisection range into two halves.
const halfDivisor = 2

func cubic(t, x1, y1, x2, y2 float64) float64 {
	lo, hi, u := 0.0, 1.0, t
	for i := 0; i < 24; i++ {
		x := bezier(u, x1, x2)
		if x < t {
			lo = u
		} else {
			hi = u
		}
		u = (lo + hi) / halfDivisor
	}
	return bezier(u, y1, y2)
}

func bezier(t, a, b float64) float64 {
	u := 1 - t
	return 3*u*u*t*a + 3*u*t*t*b + t*t*t
}

func steps(t float64, count int, position string) float64 {
	if count <= 1 && position == "jump-none" {
		return t
	}
	switch position {
	case "start", "jump-start":
		return float64(int(t*float64(count))+1) / float64(count)
	case "jump-both":
		return float64(int(t*float64(count))+1) / float64(count+1)
	case "jump-none":
		return float64(int(t*float64(count))) / float64(count-1)
	default:
		return float64(int(t*float64(count))) / float64(count)
	}
}

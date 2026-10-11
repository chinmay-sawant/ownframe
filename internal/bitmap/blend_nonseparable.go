package bitmap

func lum(c [3]float64) float64        { return .3*c[0] + .59*c[1] + .11*c[2] }
func saturation(c [3]float64) float64 { return max(c[0], c[1], c[2]) - min(c[0], c[1], c[2]) }

func setLum(c [3]float64, l float64) [3]float64 {
	d := l - lum(c)
	for i := range c {
		c[i] += d
	}
	n, x := min(c[0], c[1], c[2]), max(c[0], c[1], c[2])
	if n < 0 {
		for i := range c {
			c[i] = l + (c[i]-l)*l/(l-n)
		}
	}
	if x > 1 {
		for i := range c {
			c[i] = l + (c[i]-l)*(1-l)/(x-l)
		}
	}
	return c
}

func setSat(c [3]float64, s float64) [3]float64 {
	lo, hi := min(c[0], c[1], c[2]), max(c[0], c[1], c[2])
	for i := range c {
		if hi == lo {
			c[i] = 0
		} else {
			c[i] = (c[i] - lo) * s / (hi - lo)
		}
	}
	return c
}

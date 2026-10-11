package bitmap

import "math"

func blendRGB(b, s [3]float64, mode string) [3]float64 {
	switch mode {
	case "hue":
		return setLum(setSat(s, saturation(b)), lum(b))
	case "saturation":
		return setLum(setSat(b, saturation(s)), lum(b))
	case "color":
		return setLum(s, lum(b))
	case "luminosity":
		return setLum(b, lum(s))
	}
	for i := range b {
		b[i] = blendChannel(b[i], s[i], mode)
	}
	return b
}

func blendChannel(b, s float64, mode string) float64 {
	switch mode {
	case "multiply":
		return b * s
	case "screen":
		return b + s - b*s
	case "overlay":
		return blendChannel(s, b, "hard-light")
	case "darken":
		return min(b, s)
	case "lighten":
		return max(b, s)
	case "color-dodge":
		if b == 0 {
			return 0
		}
		if s == 1 {
			return 1
		}
		return min(1, b/(1-s))
	case "color-burn":
		if b == 1 {
			return 1
		}
		if s == 0 {
			return 0
		}
		return 1 - min(1, (1-b)/s)
	case "hard-light":
		if s <= .5 {
			return 2 * b * s
		}
		return 1 - 2*(1-b)*(1-s)
	case "soft-light":
		if s <= .5 {
			return b - (1-2*s)*b*(1-b)
		}
		d := math.Sqrt(b)
		if b <= .25 {
			d = ((16*b-12)*b + 4) * b
		}
		return b + (2*s-1)*(d-b)
	case "difference":
		return math.Abs(b - s)
	case "exclusion":
		return b + s - 2*b*s
	}
	return s
}

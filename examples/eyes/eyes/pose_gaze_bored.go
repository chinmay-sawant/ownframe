package eyes

import "math"

func boredGaze(age float64) (float64, float64) {
	phase := math.Mod(age, 6)
	switch {
	case phase < 1.5:
		return -8, 0
	case phase < 2.5:
		k := smooth((phase - 1.5) / 1)
		return -8 + 16*k, 0
	case phase < 3.7:
		return 8, -1
	case phase < 4.7:
		k := smooth((phase - 3.7) / 1)
		return 8 - 16*k, -1
	default:
		return -8, 0
	}
}

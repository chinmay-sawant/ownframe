package eyes

import "math"

func poseGaze(m poseMotion, age float64) (float64, float64) {
	switch m {
	case motionOpen:
		return 6 * math.Sin(age*1.05), 1.8 * math.Sin(age*.62)
	case motionSquint:
		return 7 * math.Sin(age*.9), 1.5 + 1.5*math.Sin(age*.65)
	case motionAngry:
		return 5 * math.Sin(age*1.8), 1 + 1.5*math.Abs(math.Sin(age*1.4))
	case motionSleepy:
		return 3 * math.Sin(age*.45), 2 + 1.5*math.Sin(age*.35)
	case motionSurprised:
		return 10 * math.Sin(age*1.7), 2 * math.Sin(age*2.1)
	case motionCurious:
		return 7 * math.Sin(age*.9), -2 + 2*math.Sin(age*.6)
	case motionWink:
		return 2.5 * math.Sin(age*1.3), 1.5 * math.Sin(age*.8)
	case motionBored:
		return boredGaze(age)
	case motionHappy:
		return 3 * math.Sin(age*.8), math.Sin(age * .9)
	case motionWorried:
		return 2.5 * math.Sin(age*.6), -2 - 1.5*math.Max(0, math.Sin(age*.4))
	case motionExcited:
		return 5 * math.Sin(age*1.4), 0
	case motionShy:
		return shyGaze(age)
	case motionRelieved:
		return 2 * math.Sin(age*.4), math.Sin(age * .5)
	default:
		return 0, 0
	}
}

func shyGaze(age float64) (float64, float64) {
	phase := math.Mod(age, 4.8)
	switch {
	case phase < 1.2:
		return -7, -7
	case phase < 2:
		k := smooth((phase - 1.2) / .8)
		return -7 + 12*k, -7
	case phase < 2.8:
		k := smooth((phase - 2) / .8)
		return 5 - 12*k, -7 + 13*k
	case phase < 3.8:
		return -7, 6
	default:
		k := smooth((phase - 3.8) / 1)
		return -7, 6 - 13*k
	}
}

func smooth(v float64) float64 {
	v = clamp(v, 0, 1)
	return v * v * (3 - 2*v)
}

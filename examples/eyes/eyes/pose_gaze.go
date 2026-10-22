package eyes

import "math"

func poseGaze(m poseMotion, age float64) (float64, float64) {
	switch m {
	case motionOpen:
		return gazeOpenX * math.Sin(age*gazeOpenFreqX), gazeOpenY * math.Sin(age*gazeOpenFreqY)
	case motionSquint:
		return gazeSquintX * math.Sin(age*gazeSquintFreqX), gazeSquintBaseY + gazeSquintAmpY*math.Sin(age*gazeSquintFreqY)
	case motionAngry:
		return gazeAngryX * math.Sin(age*gazeAngryFreqX), gazeAngryBaseY + gazeAngryAmpY*math.Abs(math.Sin(age*gazeAngryFreqY))
	case motionSleepy:
		return gazeSleepyX * math.Sin(age*gazeSleepyFreqX), gazeSleepyBaseY + gazeSleepyAmpY*math.Sin(age*gazeSleepyFreqY)
	case motionSurprised:
		return gazeSurprisedX * math.Sin(age*gazeSurprisedFreqX), gazeSurprisedAmpY * math.Sin(age*gazeSurprisedFreqY)
	case motionCurious:
		return gazeCuriousX * math.Sin(age*gazeCuriousFreqX), gazeCuriousBaseY + gazeCuriousAmpY*math.Sin(age*gazeCuriousFreqY)
	case motionWink:
		return gazeWinkX * math.Sin(age*gazeWinkFreqX), gazeWinkBaseY * math.Sin(age*gazeWinkFreqY)
	case motionBored:
		return boredGaze(age)
	case motionHappy:
		return gazeHappyX * math.Sin(age*gazeHappyFreqX), math.Sin(age * gazeHappyFreqY)
	case motionWorried:
		return gazeWorriedX * math.Sin(age*gazeWorriedFreqX), gazeWorriedBaseY - gazeWorriedAmpY*math.Max(0, math.Sin(age*gazeWorriedFreqY))
	case motionExcited:
		return gazeExcitedX * math.Sin(age*gazeExcitedFreqX), 0
	case motionShy:
		return shyGaze(age)
	case motionRelieved:
		return gazeRelievedX * math.Sin(age*gazeRelievedFreqX), math.Sin(age * gazeRelievedFreqY)
	default:
		return 0, 0
	}
}

func smooth(v float64) float64 {
	v = clamp(v, 0, 1)
	return v * v * (smoothHigh - smoothLow*v)
}

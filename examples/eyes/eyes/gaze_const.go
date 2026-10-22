package eyes

// Gaze sway per motion: horizontal and vertical amplitude with the sine
// frequency that drives each axis.
const (
	gazeOpenX     = 6
	gazeOpenFreqX = 1.05
	gazeOpenY     = 1.8
	gazeOpenFreqY = 0.62

	gazeSquintX     = 7
	gazeSquintFreqX = 0.9
	gazeSquintBaseY = 1.5
	gazeSquintAmpY  = 1.5
	gazeSquintFreqY = 0.65

	gazeAngryX     = 5
	gazeAngryFreqX = 1.8
	gazeAngryBaseY = 1
	gazeAngryAmpY  = 1.5
	gazeAngryFreqY = 1.4

	gazeSleepyX     = 3
	gazeSleepyFreqX = 0.45
	gazeSleepyBaseY = 2
	gazeSleepyAmpY  = 1.5
	gazeSleepyFreqY = 0.35

	gazeSurprisedX     = 10
	gazeSurprisedFreqX = 1.7
	gazeSurprisedAmpY  = 2
	gazeSurprisedFreqY = 2.1

	gazeCuriousX     = 7
	gazeCuriousFreqX = 0.9
	gazeCuriousBaseY = -2
	gazeCuriousAmpY  = 2
	gazeCuriousFreqY = 0.6

	gazeWinkX     = 2.5
	gazeWinkFreqX = 1.3
	gazeWinkBaseY = 1.5
	gazeWinkFreqY = 0.8

	gazeHappyX     = 3
	gazeHappyFreqX = 0.8
	gazeHappyFreqY = 0.9

	gazeWorriedX     = 2.5
	gazeWorriedFreqX = 0.6
	gazeWorriedBaseY = -2
	gazeWorriedAmpY  = 1.5
	gazeWorriedFreqY = 0.4

	gazeExcitedX     = 5
	gazeExcitedFreqX = 1.4

	gazeRelievedX     = 2
	gazeRelievedFreqX = 0.4
	gazeRelievedFreqY = 0.5
)

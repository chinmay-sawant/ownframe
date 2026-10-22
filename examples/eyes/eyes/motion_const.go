package eyes

import "math"

// Relaxed drift is one slow sine around the pose.
const (
	relaxedOpenEase = 0.07
	relaxedWave     = 2 * math.Pi / 3
)

// Happy brightens then grins on a shared beat.
const (
	happyBrightFreq = 2.1
	happyGrinLag    = 1.3
	happyBrightOpen = 0.15
	happyGrinClose  = 0.2
	happyBrightTop  = 0.08
	happyGrinBottom = 0.16
	happyGazeBright = 2
)

// Worried wobble, focused glance, and the excited bounce and shake.
const (
	worriedWobbleAmp  = 1.4
	worriedWobbleFreq = 2.1
	focusedGlanceAmp  = 3
	focusedGlanceFreq = 0.7
	excitedBounce     = 0.5
	excitedBounceFreq = 4.8
	excitedShakeAmp   = 2.2
	excitedShakeFreq  = 24
	excitedShakeAmp2  = 0.7
	excitedShakeFreq2 = 31
	excitedOpenEase   = 0.04
	excitedHeightEase = 0.05
	excitedGazeDip    = 1.5
	excitedTremorAmp  = 0.8
	excitedTremorFreq = 29
	nervousDartAmp    = 7
	nervousDartFreq   = 5.5
	nervousDartAmp2   = 2
	nervousDartFreq2  = 2.4
)

// Relieved breathing eases the lids open, holds, and settles back.
const (
	relievedPeriod  = 5
	relievedPhaseA  = 0.35
	relievedClosed  = 0.2
	relievedPhaseB  = 1.2
	relievedEaseAmp = 0.5
	relievedEaseLen = 0.85
	relievedPhaseC  = 3.8
	relievedLidOpen = 0.7
	relievedFallLen = 1.2
)

// Expressive darts, widening, and drift for sad, scared, frustrated.
const (
	sadDriftAmp         = 1.2
	sadDriftFreq        = 1.4
	scaredDartAmp       = 25
	scaredDartFreq      = 4.2
	scaredDartAmp2      = 8
	scaredDartFreq2     = 11.5
	scaredWideAmp       = 0.035
	scaredWideFreq      = 5.5
	frustratedDriftAmp  = 15
	frustratedDriftFreq = 2.3
	frustratedSqueeze   = 0.07
	frustratedSqueezeAt = 3.8
)

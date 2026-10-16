package eyes

import "time"

type eyeStyle struct {
	open, height float64
	top, bottom  float64
	slant        float64
	gazeX, gazeY float64
}

type poseMotion uint8

const (
	motionNone poseMotion = iota
	motionRelaxed
	motionHappy
	motionWorried
	motionFocused
	motionExcited
	motionShy
	motionNervous
	motionSad
	motionScared
	motionFrustrated
	motionRelieved
)

type eyePose struct {
	name        string
	left, right eyeStyle
	wink        winkSide
	motion      poseMotion
}

var poses = allPoses()

func allPoses() []eyePose {
	poses := basePoses()
	poses = append(poses, extraPoses()...)
	return append(poses, expressivePoses()...)
}

func both(name string, style eyeStyle) eyePose {
	return eyePose{name: name, left: style, right: style}
}

func poseFor(mood int) eyePose {
	if mood < 0 || mood >= len(poses) {
		return poses[0]
	}
	return poses[mood]
}

func targetOpen(mood int) (float64, float64) {
	p := poseFor(mood)
	return p.left.open, p.right.open
}

func moodDuration(mood int) time.Duration {
	if mood == spiralMood {
		return 4 * time.Second
	}
	return 8 * time.Second
}

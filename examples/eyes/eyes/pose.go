package eyes

import "time"

type eyeStyle struct {
	open, height float64
	top, bottom  float64
	slant        float64
	gazeX, gazeY float64
}

type eyePose struct {
	name        string
	left, right eyeStyle
	wink        winkSide
}

var poses = []eyePose{
	both("open", eyeStyle{open: 1, height: 1}),
	both("squint", eyeStyle{open: .34, height: 1, top: .5, bottom: .5}),
	{
		name:  "angry",
		left:  eyeStyle{open: .64, height: 1, top: 1, bottom: .08, slant: 11},
		right: eyeStyle{open: .64, height: 1, top: 1, bottom: .08, slant: -11},
	},
	both("spiral", eyeStyle{open: 1, height: 1}),
	both("relaxed", eyeStyle{open: .82, height: 1, top: .55, bottom: .12}),
	both("sleepy", eyeStyle{open: .42, height: 1, top: .9, bottom: .1, gazeY: 7}),
	both("surprised", eyeStyle{open: 1, height: 1.16, gazeY: -4}),
	{
		name:  "curious",
		left:  eyeStyle{open: .65, height: 1, top: .6, slant: -9, gazeX: 12, gazeY: -5},
		right: eyeStyle{open: 1, height: 1, slant: 4, gazeX: 12, gazeY: -5},
	},
	{
		name:  "skeptical",
		left:  eyeStyle{open: .42, height: 1, top: .8, slant: -8, gazeX: -8},
		right: eyeStyle{open: .88, height: 1, top: .2, slant: 5, gazeX: -8},
	},
	{name: "wink-left", left: eyeStyle{open: 1, height: 1, top: 1},
		right: eyeStyle{open: 1, height: 1, top: 1, gazeX: -8}, wink: winkScreenRight},
	{name: "wink-right", left: eyeStyle{open: 1, height: 1, top: 1, gazeX: 8},
		right: eyeStyle{open: 1, height: 1, top: 1}, wink: winkScreenLeft},
	both("bored", eyeStyle{open: .4, height: 1, top: .7, bottom: .3, gazeY: 10}),
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

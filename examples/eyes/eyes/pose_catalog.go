package eyes

func basePoses() []eyePose {
	return []eyePose{
		bothMotion("open", eyeStyle{open: 1, height: 1}, motionOpen),
		bothMotion("squint", eyeStyle{open: .34, height: 1, top: .5, bottom: .5}, motionSquint),
		{name: "angry",
			left:   eyeStyle{open: .64, height: 1, top: 1, bottom: .08, slant: 11, gazeY: -4},
			right:  eyeStyle{open: .64, height: 1, top: 1, bottom: .08, slant: -11, gazeY: -4},
			motion: motionAngry},
		both("spiral", eyeStyle{open: 1, height: 1}),
		{name: "relaxed", left: eyeStyle{open: .68, height: 1, top: .82, bottom: .3, gazeY: 6},
			right: eyeStyle{open: .68, height: 1, top: .82, bottom: .3, gazeY: 6}, motion: motionRelaxed},
		bothMotion("sleepy", eyeStyle{open: .42, height: 1, top: .9, bottom: .1, gazeY: 7}, motionSleepy),
		bothMotion("surprised", eyeStyle{open: 1, height: 1.16, gazeY: -4}, motionSurprised),
		{name: "curious", left: eyeStyle{open: .65, height: 1, top: .6, slant: -9, gazeX: 12, gazeY: -5},
			right: eyeStyle{open: 1, height: 1, slant: 4, gazeX: 12, gazeY: -5}, motion: motionCurious},
		{name: "wink-left", left: eyeStyle{open: 1, height: 1, top: 1},
			right: eyeStyle{open: 1, height: 1, top: 1, gazeX: -8}, wink: winkScreenRight, motion: motionWink},
		{name: "wink-right", left: eyeStyle{open: 1, height: 1, top: 1, gazeX: 8},
			right: eyeStyle{open: 1, height: 1, top: 1}, wink: winkScreenLeft, motion: motionWink},
		bothMotion("bored", eyeStyle{open: .4, height: 1, top: .7, bottom: .3, gazeY: 10}, motionBored),
	}
}

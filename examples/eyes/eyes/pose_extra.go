package eyes

func extraPoses() []eyePose {
	return []eyePose{
		{name: "happy", left: eyeStyle{open: .65, height: 1, top: .3, bottom: .72},
			right: eyeStyle{open: .65, height: 1, top: .3, bottom: .72}, motion: motionHappy},
		{name: "worried", left: eyeStyle{open: .82, height: 1, top: .72, slant: -8},
			right: eyeStyle{open: .82, height: 1, top: .72, slant: 8}, motion: motionWorried},
		{name: "focused", left: eyeStyle{open: .48, height: .94, top: .72},
			right: eyeStyle{open: .48, height: .94, top: .72}, motion: motionFocused},
		{name: "confused", left: eyeStyle{open: .52, height: .96, top: .62, slant: -7, gazeX: 8},
			right: eyeStyle{open: .9, height: 1.04, top: .2, slant: 5, gazeX: 8}, motion: motionConfused},
		{name: "excited", left: eyeStyle{open: .94, height: 1.12, top: .08, gazeY: -3},
			right: eyeStyle{open: .94, height: 1.12, top: .08, gazeY: -3}, motion: motionExcited},
		{name: "shy", left: eyeStyle{open: .56, height: 1, top: .68, gazeX: 7, gazeY: 7},
			right: eyeStyle{open: .56, height: 1, top: .68, gazeX: 7, gazeY: 7}, motion: motionShy},
		{name: "nervous", left: eyeStyle{open: .86, height: 1.04, top: .18},
			right: eyeStyle{open: .86, height: 1.04, top: .18}, motion: motionNervous},
		{name: "anticipation", left: eyeStyle{open: .05, height: 1, top: 1},
			right: eyeStyle{open: .05, height: 1, top: 1}, motion: motionAnticipation},
	}
}

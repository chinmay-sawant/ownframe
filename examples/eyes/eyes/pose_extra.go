package eyes

func extraPoses() []eyePose {
	return []eyePose{
		{name: "happy", left: eyeStyle{open: .65, height: 1, top: .3, bottom: .72},
			right: eyeStyle{open: .65, height: 1, top: .3, bottom: .72}, motion: motionHappy},
		{name: "worried", left: eyeStyle{open: .82, height: 1, top: .72, slant: -8},
			right: eyeStyle{open: .82, height: 1, top: .72, slant: 8}, motion: motionWorried},
		{name: "focused", left: eyeStyle{open: .48, height: .94, top: .72},
			right: eyeStyle{open: .48, height: .94, top: .72}, motion: motionFocused},
		{name: "excited", left: eyeStyle{open: .94, height: 1.12, top: .08, gazeY: -3},
			right: eyeStyle{open: .94, height: 1.12, top: .08, gazeY: -3}, motion: motionExcited},
		{name: "shy", left: eyeStyle{open: .56, height: 1, top: .68, gazeX: 7, gazeY: 7},
			right: eyeStyle{open: .56, height: 1, top: .68, gazeX: 7, gazeY: 7}, motion: motionShy},
		{name: "nervous", left: eyeStyle{open: .76, height: 1.02, top: .5, bottom: .16, slant: -3, gazeY: 6},
			right: eyeStyle{open: .76, height: 1.02, top: .5, bottom: .16, slant: 3, gazeY: 6}, motion: motionNervous},
		{name: "relieved", left: eyeStyle{open: .7, height: 1, top: .82, bottom: .12, gazeY: 2},
			right: eyeStyle{open: .7, height: 1, top: .82, bottom: .12, gazeY: 2}, motion: motionRelieved},
	}
}

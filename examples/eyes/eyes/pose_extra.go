package eyes

func extraPoses() []eyePose {
	return []eyePose{
		{name: "happy", left: eyeStyle{open: .65, height: 1},
			right: eyeStyle{open: .65, height: 1}, motion: motionHappy},
		{name: "worried", left: eyeStyle{open: .82, height: 1, gazeY: 2},
			right: eyeStyle{open: .82, height: 1, gazeY: 2}, motion: motionWorried},
		{name: "focused", left: eyeStyle{open: .48, height: .94},
			right: eyeStyle{open: .48, height: .94}, motion: motionFocused},
		{name: "excited", left: eyeStyle{open: .94, height: 1.12, gazeY: -3},
			right: eyeStyle{open: .94, height: 1.12, gazeY: -3}, motion: motionExcited},
		{name: "shy", left: eyeStyle{open: .56, height: 1, gazeX: 7, gazeY: 7},
			right: eyeStyle{open: .56, height: 1, gazeX: 7, gazeY: 7}, motion: motionShy},
		{name: "nervous", left: eyeStyle{open: .86, height: 1.04},
			right: eyeStyle{open: .86, height: 1.04}, motion: motionNervous},
		{name: "relieved", left: eyeStyle{open: .7, height: 1, gazeY: 2},
			right: eyeStyle{open: .7, height: 1, gazeY: 2}, motion: motionRelieved},
	}
}

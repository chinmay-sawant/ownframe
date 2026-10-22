package eyes

func extraPoses() []eyePose {
	return []eyePose{
		{name: "happy", left: eyeStyle{open: happyOpen, height: 1, top: happyTop, bottom: happyBottom},
			right: eyeStyle{open: happyOpen, height: 1, top: happyTop, bottom: happyBottom}, motion: motionHappy},
		{name: "worried", left: eyeStyle{open: worriedOpen, height: 1, top: worriedTop, slant: -worriedSlant},
			right: eyeStyle{open: worriedOpen, height: 1, top: worriedTop, slant: worriedSlant}, motion: motionWorried},
		{name: "focused", left: eyeStyle{open: focusedOpen, height: focusedHeight, top: focusedTop},
			right: eyeStyle{open: focusedOpen, height: focusedHeight, top: focusedTop}, motion: motionFocused},
		{name: "excited", left: eyeStyle{open: excitedOpen, height: excitedHeight, top: excitedTop, gazeY: excitedGazeY},
			right: eyeStyle{open: excitedOpen, height: excitedHeight, top: excitedTop, gazeY: excitedGazeY}, motion: motionExcited},
		{name: "shy", left: eyeStyle{open: shyOpen, height: 1, top: shyTop, gazeX: shyGazeXY, gazeY: shyGazeXY},
			right: eyeStyle{open: shyOpen, height: 1, top: shyTop, gazeX: shyGazeXY, gazeY: shyGazeXY}, motion: motionShy},
		{name: "nervous", left: eyeStyle{open: nervousOpen, height: nervousHeight, top: nervousTop, bottom: nervousLow, slant: -nervousSlant, gazeY: nervousGazeY},
			right: eyeStyle{open: nervousOpen, height: nervousHeight, top: nervousTop, bottom: nervousLow, slant: nervousSlant, gazeY: nervousGazeY}, motion: motionNervous},
		{name: "relieved", left: eyeStyle{open: relievedLidOpen, height: 1, top: relievedTop, bottom: relievedBottom, gazeY: relievedGazeY},
			right: eyeStyle{open: relievedLidOpen, height: 1, top: relievedTop, bottom: relievedBottom, gazeY: relievedGazeY}, motion: motionRelieved},
	}
}

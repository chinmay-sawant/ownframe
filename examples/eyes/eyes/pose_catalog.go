package eyes

func basePoses() []eyePose {
	return []eyePose{
		bothMotion("open", eyeStyle{open: 1, height: 1}, motionOpen),
		bothMotion("squint", eyeStyle{open: squintOpen, height: 1, top: squintLid, bottom: squintLid}, motionSquint),
		{name: "angry",
			left:   eyeStyle{open: angryOpen, height: 1, top: 1, bottom: angryLid, slant: angrySlant, gazeY: angryGazeY},
			right:  eyeStyle{open: angryOpen, height: 1, top: 1, bottom: angryLid, slant: -angrySlant, gazeY: angryGazeY},
			motion: motionAngry},
		both("spiral", eyeStyle{open: 1, height: 1}),
		{name: "relaxed", left: eyeStyle{open: relaxedOpen, height: 1, top: relaxedTop, bottom: relaxedBottom, gazeY: relaxedGazeY},
			right: eyeStyle{open: relaxedOpen, height: 1, top: relaxedTop, bottom: relaxedBottom, gazeY: relaxedGazeY}, motion: motionRelaxed},
		bothMotion("sleepy", eyeStyle{open: sleepyOpen, height: 1, top: sleepyTop, bottom: sleepyBottom, gazeY: sleepyGazeY}, motionSleepy),
		bothMotion("surprised", eyeStyle{open: 1, height: surprisedHeight, gazeY: surprisedGazeY}, motionSurprised),
		{name: "curious", left: eyeStyle{open: curiousOpen, height: 1, top: curiousTop, slant: curiousSlantL, gazeX: curiousGazeX, gazeY: curiousGazeY},
			right: eyeStyle{open: 1, height: 1, slant: curiousSlantR, gazeX: curiousGazeX, gazeY: curiousGazeY}, motion: motionCurious},
		{name: "wink-left", left: eyeStyle{open: 1, height: 1, top: 1},
			right: eyeStyle{open: 1, height: 1, top: 1, gazeX: -winkGazeX}, wink: winkScreenRight, motion: motionWink},
		{name: "wink-right", left: eyeStyle{open: 1, height: 1, top: 1, gazeX: winkGazeX},
			right: eyeStyle{open: 1, height: 1, top: 1}, wink: winkScreenLeft, motion: motionWink},
		bothMotion("bored", eyeStyle{open: boredOpen, height: 1, top: boredTop, bottom: boredBottom, gazeY: boredGazeY}, motionBored),
	}
}

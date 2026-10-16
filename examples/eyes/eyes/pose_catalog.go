package eyes

func basePoses() []eyePose {
	return []eyePose{
		both("open", eyeStyle{open: 1, height: 1}),
		both("squint", eyeStyle{open: .34, height: 1}),
		both("angry", eyeStyle{open: .54, height: 1, gazeY: -4}),
		both("spiral", eyeStyle{open: 1, height: 1}),
		{name: "relaxed", left: eyeStyle{open: .68, height: 1, gazeY: 6},
			right: eyeStyle{open: .68, height: 1, gazeY: 6}, motion: motionRelaxed},
		both("sleepy", eyeStyle{open: .42, height: 1, gazeY: 7}),
		both("surprised", eyeStyle{open: 1, height: 1.16, gazeY: -4}),
		{name: "curious", left: eyeStyle{open: .65, height: 1, gazeX: 12, gazeY: -5},
			right: eyeStyle{open: 1, height: 1, gazeX: 12, gazeY: -5}},
		{name: "wink-left", left: eyeStyle{open: 1, height: 1},
			right: eyeStyle{open: 1, height: 1, gazeX: -8}, wink: winkScreenRight},
		{name: "wink-right", left: eyeStyle{open: 1, height: 1, gazeX: 8},
			right: eyeStyle{open: 1, height: 1}, wink: winkScreenLeft},
		both("bored", eyeStyle{open: .4, height: 1, gazeY: 10}),
	}
}

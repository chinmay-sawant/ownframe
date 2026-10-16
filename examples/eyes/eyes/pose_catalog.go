package eyes

func basePoses() []eyePose {
	return []eyePose{
		both("open", eyeStyle{open: 1, height: 1}),
		both("squint", eyeStyle{open: .34, height: 1, top: .5, bottom: .5}),
		{name: "angry", left: eyeStyle{open: .64, height: 1, top: 1, bottom: .08, slant: 11},
			right: eyeStyle{open: .64, height: 1, top: 1, bottom: .08, slant: -11}},
		both("spiral", eyeStyle{open: 1, height: 1}),
		{name: "relaxed", left: eyeStyle{open: .68, height: 1, top: .82, bottom: .3, gazeY: 6},
			right: eyeStyle{open: .68, height: 1, top: .82, bottom: .3, gazeY: 6}, motion: motionRelaxed},
		both("sleepy", eyeStyle{open: .42, height: 1, top: .9, bottom: .1, gazeY: 7}),
		both("surprised", eyeStyle{open: 1, height: 1.16, gazeY: -4}),
		{name: "curious", left: eyeStyle{open: .65, height: 1, top: .6, slant: -9, gazeX: 12, gazeY: -5},
			right: eyeStyle{open: 1, height: 1, slant: 4, gazeX: 12, gazeY: -5}},
		{name: "skeptical", left: eyeStyle{open: .42, height: 1, top: .8, slant: -8, gazeX: -8},
			right: eyeStyle{open: .88, height: 1, top: .2, slant: 5, gazeX: -8}},
		{name: "wink-left", left: eyeStyle{open: 1, height: 1, top: 1},
			right: eyeStyle{open: 1, height: 1, top: 1, gazeX: -8}, wink: winkScreenRight},
		{name: "wink-right", left: eyeStyle{open: 1, height: 1, top: 1, gazeX: 8},
			right: eyeStyle{open: 1, height: 1, top: 1}, wink: winkScreenLeft},
		both("bored", eyeStyle{open: .4, height: 1, top: .7, bottom: .3, gazeY: 10}),
	}
}

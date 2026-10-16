package eyes

func expressivePoses() []eyePose {
	return firstExpressivePoses()
}

func firstExpressivePoses() []eyePose {
	return []eyePose{
		{name: "sad", left: eyeStyle{open: .58, height: 1, top: .78, bottom: .2, slant: -5, gazeY: 10},
			right: eyeStyle{open: .58, height: 1, top: .78, bottom: .2, slant: 5, gazeY: 10}, motion: motionSad},
		{name: "scared", left: eyeStyle{open: 1, height: 1.16, top: .02, bottom: .03, slant: -3, gazeY: -2},
			right: eyeStyle{open: 1, height: 1.16, top: .02, bottom: .03, slant: 3, gazeY: -2}, motion: motionScared},
		{name: "frustrated", left: eyeStyle{open: .38, height: 1, top: .5, bottom: .5, gazeX: -5, gazeY: 10},
			right: eyeStyle{open: .38, height: 1, top: .5, bottom: .5, gazeX: 5, gazeY: 10}, motion: motionFrustrated},
	}
}

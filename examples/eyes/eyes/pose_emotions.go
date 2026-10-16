package eyes

func expressivePoses() []eyePose {
	return append(firstExpressivePoses(), moreExpressivePoses()...)
}

func firstExpressivePoses() []eyePose {
	return []eyePose{
		{name: "sad", left: eyeStyle{open: .58, height: 1, top: .78, bottom: .2, slant: -5, gazeY: 10},
			right: eyeStyle{open: .58, height: 1, top: .78, bottom: .2, slant: 5, gazeY: 10}, motion: motionSad},
		{name: "embarrassed", left: eyeStyle{open: .54, height: 1, top: .72, bottom: .2, gazeX: -8, gazeY: 10},
			right: eyeStyle{open: .54, height: 1, top: .72, bottom: .2, gazeX: -8, gazeY: 10}, motion: motionEmbarrassed},
		{name: "scared", left: eyeStyle{open: 1, height: 1.16, top: .02, bottom: .03, slant: -3, gazeY: -2},
			right: eyeStyle{open: 1, height: 1.16, top: .02, bottom: .03, slant: 3, gazeY: -2}, motion: motionScared},
		{name: "frustrated", left: eyeStyle{open: .5, height: 1, top: .94, slant: 9},
			right: eyeStyle{open: .5, height: 1, top: .94, slant: -9}, motion: motionFrustrated},
		{name: "suspicious", left: eyeStyle{open: .44, height: 1, top: .92, slant: -7, gazeX: 9},
			right: eyeStyle{open: .84, height: 1, top: .25, slant: 4, gazeX: 9}, motion: motionSuspicious},
	}
}

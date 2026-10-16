package eyes

func expressivePoses() []eyePose {
	return firstExpressivePoses()
}

func firstExpressivePoses() []eyePose {
	return []eyePose{
		{name: "sad", left: eyeStyle{open: .58, height: 1, gazeY: 10},
			right: eyeStyle{open: .58, height: 1, gazeY: 10}, motion: motionSad},
		{name: "scared", left: eyeStyle{open: 1, height: 1.16, gazeY: -2},
			right: eyeStyle{open: 1, height: 1.16, gazeY: -2}, motion: motionScared},
		{name: "frustrated", left: eyeStyle{open: .34, height: .92, gazeX: -5, gazeY: 10},
			right: eyeStyle{open: .34, height: .92, gazeX: 5, gazeY: 10}, motion: motionFrustrated},
	}
}

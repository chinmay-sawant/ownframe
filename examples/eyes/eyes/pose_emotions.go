package eyes

func expressivePoses() []eyePose {
	return firstExpressivePoses()
}

func firstExpressivePoses() []eyePose {
	return []eyePose{
		{name: "sad", left: eyeStyle{open: sadOpen, height: 1, top: sadTop, bottom: sadBottom, slant: sadSlantL, gazeY: sadGazeY},
			right: eyeStyle{open: sadOpen, height: 1, top: sadTop, bottom: sadBottom, slant: -sadSlantL, gazeY: sadGazeY}, motion: motionSad},
		{name: "scared", left: eyeStyle{open: 1, height: scaredHeight, top: scaredTop, bottom: scaredBottom, slant: -scaredSlant, gazeY: scaredGazeY},
			right: eyeStyle{open: 1, height: scaredHeight, top: scaredTop, bottom: scaredBottom, slant: scaredSlant, gazeY: scaredGazeY}, motion: motionScared},
		{name: "frustrated", left: eyeStyle{open: frustratedOpen, height: 1, top: squintLid, bottom: squintLid, gazeX: -frustratedGazeX, gazeY: frustratedGazeY},
			right: eyeStyle{open: frustratedOpen, height: 1, top: squintLid, bottom: squintLid, gazeX: frustratedGazeX, gazeY: frustratedGazeY}, motion: motionFrustrated},
	}
}

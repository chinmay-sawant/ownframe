package eyes

func moreExpressivePoses() []eyePose {
	return []eyePose{
		{name: "smug", left: eyeStyle{open: .54, height: 1, top: .55, slant: -5, gazeX: 7},
			right: eyeStyle{open: .72, height: 1, top: .24, slant: 3, gazeX: 7}, motion: motionSmug},
		{name: "mischievous", left: eyeStyle{open: .78, height: 1, top: .25, slant: -4, gazeX: 10},
			right: eyeStyle{open: .6, height: 1, top: .78, slant: 5, gazeX: 10}, motion: motionMischievous},
		{name: "disgusted", left: eyeStyle{open: .52, height: 1, top: .86, slant: -7, gazeX: -10},
			right: eyeStyle{open: .44, height: 1, top: .92, slant: 7, gazeX: -10}, motion: motionDisgusted},
		{name: "proud", left: eyeStyle{open: .68, height: 1, top: .3, slant: -3, gazeY: -5},
			right: eyeStyle{open: .68, height: 1, top: .3, slant: 3, gazeY: -5}, motion: motionProud},
		{name: "pleading", left: eyeStyle{open: .82, height: 1, top: .82, bottom: .22, slant: 5, gazeY: 8},
			right: eyeStyle{open: .82, height: 1, top: .82, bottom: .22, slant: -5, gazeY: 8}, motion: motionPleading},
	}
}

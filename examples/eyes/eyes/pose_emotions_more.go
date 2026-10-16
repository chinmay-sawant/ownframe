package eyes

func moreExpressivePoses() []eyePose {
	return []eyePose{
		{name: "smug", left: eyeStyle{open: .72, height: 1, top: .25, slant: -5, gazeX: 11,
			browUp: 20, browSlant: -7},
			right: eyeStyle{open: .42, height: 1, top: .72, slant: 8, gazeX: 11,
				browUp: 5, browSlant: 8}, motion: motionSmug},
		{name: "mischievous", left: eyeStyle{open: .58, height: 1, top: .3, slant: -6, gazeX: 14,
			browUp: 7, browSlant: -12},
			right: eyeStyle{open: .68, height: 1, top: .58, slant: 8, gazeX: 14,
				browUp: 18, browSlant: 10}, motion: motionMischievous},
		{name: "disgusted", left: eyeStyle{open: .4, height: 1, top: .88, bottom: .26, slant: -10, gazeX: 16, gazeY: -7},
			right: eyeStyle{open: .5, height: 1, top: .78, bottom: .18, slant: 5, gazeX: 16, gazeY: -7}, motion: motionDisgusted},
		{name: "proud", left: eyeStyle{open: .72, height: 1, top: .24, slant: -3, gazeY: -13},
			right: eyeStyle{open: .72, height: 1, top: .24, slant: 3, gazeY: -13}, motion: motionProud},
		{name: "pleading", left: eyeStyle{open: .98, height: 1.16, top: .4, bottom: .04,
			slant: -8, gazeY: -14, browUp: 24, browSlant: -18},
			right: eyeStyle{open: .98, height: 1.16, top: .4, bottom: .04,
				slant: 8, gazeY: -14, browUp: 24, browSlant: 18}, motion: motionPleading},
	}
}

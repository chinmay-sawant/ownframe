package main

import "github.com/chinmay-sawant/ownframe/internal/cssmotion"

func easing(value string, progress float64) float64 {
	return cssmotion.Ease(value, progress)
}

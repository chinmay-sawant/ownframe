package main

import (
	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

type paint = cssmotion.Paint

func capture(d *ownframe.Display, box ownframe.Box) []paint {
	return cssmotion.Capture(d, box)
}

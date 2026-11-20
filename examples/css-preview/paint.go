package main

import (
	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/motion"
)

type paint = motion.Paint

func capture(d *ownframe.Display, box ownframe.Box) []paint {
	return motion.Capture(d, box)
}

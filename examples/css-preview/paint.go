package main

import (
	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

type paint = cssmotion.Paint

func capture(d *ownframe.Display, box ownframe.Box) []paint {
	return cssmotion.Capture(d, box)
}

func inside(d *ownframe.Display, b ownframe.Box, o ownframe.DisplayOp) bool {
	return cssmotion.Contains(d, b, o)
}

func read(o *ownframe.DisplayOp) paint { return cssmotion.Read(o) }

func write(o *ownframe.DisplayOp, v paint) { cssmotion.Write(o, v) }

package snip

import "time"

// Fallback image, capture, color, and drawing sizes.
const (
	demoWidth       = 320
	demoHeight      = 200
	demoMinSize     = 8
	blueBase        = 128
	snipDirPerm     = 0o750
	snipFilePerm    = 0o600
	captureTimeout  = 10 * time.Second
	shortHexLen     = 3
	longHexLen      = 6
	redShift        = 16
	greenShift      = 8
	bresenhamFactor = 2
	arrowLen        = 9
	arrowWidth      = 4
	stripInset      = 2
	demoEdge        = 2
)

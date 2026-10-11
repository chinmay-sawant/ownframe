package bitmap

import (
	"math"

	"golang.org/x/image/math/fixed"
)

func fixedSpacing(v float64) fixed.Int26_6 { return fixed.Int26_6(math.Round(v * 64)) }

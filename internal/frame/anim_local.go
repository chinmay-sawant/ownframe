package frame

import (
	"github.com/chinmay-sawant/ownframe/internal/page"
	"github.com/chinmay-sawant/ownframe/internal/page/dirty"
)

// paintCenter is op's painted center in points, the origin scale and rotate
// turn about.
func paintCenter(op *page.DisplayOp) (float64, float64) {
	x0, y0, x1, y1 := dirty.LocalPaintBox(op)

	return (x0 + x1) / halfDivisor, (y0 + y1) / halfDivisor
}

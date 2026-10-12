package board

import (
	"context"
	"time"
)

// Rotate moves to the next feed without a Redraw: the view and the flaps
// take the new rows, and the clock keeps its own tick. The tick calls it when
// the rotation is due, and the dump calls it to measure the real rotation.
// Advance stays the full resync path.
func (a *App) Rotate(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	a.feed++
	a.lastRotate = time.Now()
	a.retarget(rowsAt(a.feed))

	return nil
}

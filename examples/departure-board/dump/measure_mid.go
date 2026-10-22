package dump

import (
	"context"
	"fmt"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/departure-board/board"
)

// midAction runs the scenario's mid action. "rotate" takes the tick's
// retarget path and must leave the page generation alone, so nothing after
// it rebuilds the buffer. Any other value is the Advance resync, and its
// Redraw time lands in RedrawNS.
func midAction(ctx context.Context, app *board.App, mid string, rep *Report) error {
	if mid == "rotate" {
		gen := app.Page().Generation()
		if err := app.Rotate(ctx); err != nil {
			return err
		}
		if got := app.Page().Generation(); got != gen {
			return fmt.Errorf("dump: rotate redrew the page: generation %d to %d", gen, got)
		}
		return nil
	}

	start := time.Now()
	if err := app.Advance(ctx); err != nil {
		return err
	}
	rep.RedrawNS += int64(time.Since(start))
	return nil
}

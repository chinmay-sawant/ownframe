package host

import "context"

// RedrawBatcher is a screen that can defer its draws across a pointer
// sequence. BeginRedrawBatch starts the deferral and EndRedrawBatch draws the
// final state once. A screen without it draws every step.
type RedrawBatcher interface {
	BeginRedrawBatch()
	EndRedrawBatch(ctx context.Context) error
}

// Batch runs fn inside one redraw batch when screen supports it. The final
// draw includes every state change fn made, so a focus step followed by a
// click no longer rebuilds the page twice.
func Batch(ctx context.Context, screen Screen, fn func() error) (err error) {
	batcher, ok := screen.(RedrawBatcher)
	if !ok {
		return fn()
	}

	batcher.BeginRedrawBatch()
	defer func() {
		if endErr := batcher.EndRedrawBatch(ctx); err == nil {
			err = endErr
		}
	}()

	return fn()
}

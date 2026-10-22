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
//
// The screen is any value and the batch role is what counts: a screen that
// does not batch draws every step, so fn runs on its own. A fake that
// implements the role alone can use the batch too.
//
// A panic inside fn ends the batch without a draw and keeps travelling, so
// the crash report sees the panic the caller raised.
func Batch(ctx context.Context, screen any, fn func() error) (err error) {
	batcher, ok := screen.(RedrawBatcher)
	if !ok {
		return fn()
	}

	batcher.BeginRedrawBatch()

	completed := false

	defer func() {
		if !completed {
			// fn panicked. The batch ends without a draw, so the panic
			// that is travelling reaches the crash report unchanged.
			endQuietly(ctx, batcher)

			return
		}

		if endErr := batcher.EndRedrawBatch(ctx); err == nil {
			err = endErr
		}
	}()

	err = fn()
	completed = true

	return err
}

// endQuietly ends a batch that must not draw. The end runs with a dead
// context, so a screen that checks the context first paints nothing, and a
// panic inside the end cannot replace the panic that is travelling.
func endQuietly(ctx context.Context, batcher RedrawBatcher) {
	if ctx == nil {
		return
	}

	dead, cancel := context.WithCancel(ctx)
	cancel()

	defer func() { _ = recover() }()

	_ = batcher.EndRedrawBatch(dead)
}

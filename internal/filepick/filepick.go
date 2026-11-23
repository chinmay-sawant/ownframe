// Package filepick opens the dialog that picks one file.
// The compiler selects the implementation for the current system:
// comdlg32 on Windows, zenity or kdialog on Linux, and osascript on macOS.
// A system without a dialog returns ErrNoDialog. Nil context never runs
// a dialog and returns ErrNilContext. User cancel returns ErrCanceled.
package filepick

import (
	"context"
	"errors"
	"fmt"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// ErrNilContext means Pick got a nil context.
var ErrNilContext = host.ErrNilContext

// ErrCanceled means the person closed the dialog or the context ended.
var ErrCanceled = errors.New("filepick: canceled")

// ErrNoDialog means this system has no file dialog.
var ErrNoDialog = errors.New("filepick: no dialog")

func requireCtx(ctx context.Context) error {
	if ctx == nil {
		return ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return ctxCanceled(err)
	}

	return nil
}

// ctxCanceled reports an ended context as ErrCanceled that also wraps the
// context error, so callers match either errors.Is(err, ErrCanceled) or
// errors.Is(err, context.Canceled) / context.DeadlineExceeded.
func ctxCanceled(ctxErr error) error {
	return fmt.Errorf("%w: %w", ErrCanceled, ctxErr)
}

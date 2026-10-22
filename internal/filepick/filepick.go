// Package filepick opens the dialog that picks one file.
// The compiler selects the implementation for the current system:
// comdlg32 on Windows, zenity or kdialog on Linux, and osascript on macOS.
// A system without a dialog returns ErrNoDialog. Nil context never runs
// a dialog and returns ErrNilContext. User cancel returns ErrCanceled.
package filepick

import (
	"context"
	"errors"

	"github.com/chinmay-sawant/ownframe/internal/fetch"
)

// ErrNilContext means Pick got a nil context.
var ErrNilContext = fetch.ErrNilContext

// ErrCanceled means the person closed the dialog or the context ended.
var ErrCanceled = errors.New("filepick: canceled")

// ErrNoDialog means this system has no file dialog.
var ErrNoDialog = errors.New("filepick: no dialog")

func requireCtx(ctx context.Context) error {
	if ctx == nil {
		return ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return ErrCanceled
	}

	return nil
}

// Package filepick opens the dialog that picks one file.
// The compiler selects the implementation for the current system:
// comdlg32 on Windows, zenity or kdialog on Linux, and osascript on macOS.
// A system without a dialog returns ok false. Nil context never runs
// a dialog and returns ok false; ErrNilContext is the canonical cause.
package filepick

import "github.com/chinmay-sawant/ownframe/internal/fetch"

// ErrNilContext means Pick got a nil context.
var ErrNilContext = fetch.ErrNilContext

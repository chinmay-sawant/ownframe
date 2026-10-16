// Package window shows a screen in a window.
//
// On a desktop, Run opens a normal window with a title bar. The person can
// drag the edges. The minimum size comes from the screen.
// On a phone, BindMobile hands NewGame to the system view, and the screen
// size is the phone. In a browser, build with GOOS=js GOARCH=wasm and Run
// uses the page instead of a desktop window. Resizing the browser changes
// the same frame.
//
// Mouse clicks and taps call Screen.Click. Typed text calls Type.
// Every key press and release calls KeyDown and KeyUp with a lowercase
// key name.
// Backspace calls Backspace. Ctrl-Backspace calls DeleteWord.
// Enter calls Submit. Ctrl or Command with C, V, X, A, Z, and Y call
// Copy, Paste, Cut, SelectAll, Undo, and Redo.
// The wheel scrolls a page that is taller or wider than the window.
package window

import (
	"context"
	"errors"
	"fmt"

	"github.com/chinmay-sawant/ownframe/internal/fetch"
	"github.com/chinmay-sawant/ownframe/internal/host"
)

const (
	backspaceRepeatAt    = 30
	backspaceRepeatEvery = 4
)

var (
	errNilApp = errors.New("window: nil screen")
	// Window keeps its prefix only as wrapped context; identity is the
	// canonical fetch.ErrNilContext so one errors.Is branch covers Run.
	errNilContext = fmt.Errorf("window: %w", fetch.ErrNilContext)
	errNoImage    = errors.New("window: no image")
)

// Run opens the desktop window, or the browser canvas when built for wasm.
// Run must be called from main, and it returns when the window closes.
func Run(ctx context.Context, app host.Screen) error {
	if ctx == nil {
		return errNilContext
	}

	if app == nil {
		return errNilApp
	}

	return runWindow(ctx, app, Options{})
}

package ownframe

import (
	"context"
	"errors"

	"github.com/chinmay-sawant/ownframe/internal/filepick"
	pagepkg "github.com/chinmay-sawant/ownframe/internal/page"
	"github.com/chinmay-sawant/ownframe/internal/window"
)

// WindowOptions controls desktop window appearance and input.
// The zero value opens the same window as Run. Risky surfaces stay off
// unless opted in: MousePassthrough passes pointer input through painted
// pixels, so keep it false unless the app is a click-through overlay.
type WindowOptions = window.Options

// RunWithOptions opens a page with optional desktop window settings.
// A startup or window failure is stored in a crash report file first; the
// returned error keeps the original without the file path. A canceled
// context or a deadline returns that error and writes no file. A panic
// still goes through the recover in savePanic. Use LastReportPath for the file.
func RunWithOptions(ctx context.Context, page *Page, options WindowOptions) (err error) {
	defer savePanic(page, &err)

	if err := prepare(ctx, page); err != nil {
		return reportRunErr(page, err)
	}

	pagepkg.InstallPicker(innerPage(page), filepick.Pick)
	ensureAudio()

	if err := window.RunWithOptions(ctx, innerPage(page), options); err != nil {
		return reportRunErr(page, err)
	}

	return nil
}

// reportRunErr writes a crash report for a startup or window failure.
// Stopping the window by canceling its context, or by a deadline, is a
// normal return. Those errors are handed back as they are.
func reportRunErr(page *Page, err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}

	return reportError(page, err)
}

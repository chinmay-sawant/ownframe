package ipc

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// Page returns the ownframe page Run and Serve display.
func (a *App) Page() *ownframe.Page {
	return a.page
}

// View returns the current status, received text, and message tape.
func (a *App) View() View {
	return a.view
}

// Cancel removes both demo.log listeners and the demo.double handler.
// A later Send finds no listener; Request returns ErrNoHandler.
func (a *App) Cancel() {
	if a.stopLog != nil {
		a.stopLog()
		a.stopLog = nil
	}

	if a.stopCount != nil {
		a.stopCount()
		a.stopCount = nil
	}

	if a.stopDouble != nil {
		a.stopDouble()
		a.stopDouble = nil
	}

	a.view.Live = false
}

// say stores the outcome and copies it to the top of the message tape.
func (a *App) say(bad bool, line string) {
	a.view.Status = line
	a.view.Bad = bad

	a.view.Events = append([]string{line}, a.view.Events...)
	if len(a.view.Events) > maxEvents {
		a.view.Events = a.view.Events[:maxEvents]
	}
}

// Redraw fills the template and renders the current size.
func (a *App) Redraw(ctx context.Context) error {
	return a.page.Redraw(ctx)
}

// Boxes returns the last hit-test boxes.
func (a *App) Boxes() []ownframe.Box {
	return a.page.Boxes()
}

// Click hit-tests the page and runs the control under the point.
func (a *App) Click(ctx context.Context, x, y float64) error {
	return a.page.Click(ctx, x, y)
}

// Type edits the focused control. The library owns the field text.
func (a *App) Type(ctx context.Context, text string) error {
	return a.page.Type(ctx, text)
}

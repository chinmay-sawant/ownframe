// Package synth implements the Pocket Synthesizer example.
package synth

import (
	"context"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// App is the pocket synthesizer application.
type App struct {
	page      *ownframe.Page
	view      View
	audio     *AudioHost
	tickCount uint64
	seqStep   int
	lastStep  time.Time
}

// New creates and initializes a new Pocket Synthesizer page.
func New() (*App, error) {
	page, err := ownframe.New(ownframe.Config{
		DevTools:  true,
		Title:     "Pocket Synthesizer",
		HTML:      buildHTML(),
		Width:     DefaultWidth,
		Height:    DefaultHeight,
		MinWidth:  MinWidth,
		MinHeight: MinHeight,
	})
	if err != nil {
		return nil, err
	}

	app := &App{
		page:  page,
		view:  initialView(),
		audio: NewAudioHost(),
	}

	page.Handle(ownframe.Handlers{
		Click:   app.onClick,
		KeyDown: app.onKeyDown,
	})
	page.SetTick(app.Tick)
	page.SetData(&app.view)

	return app, nil
}

// Page returns the underlying ownframe Page.
func (a *App) Page() *ownframe.Page {
	return a.page
}

// View returns a copy of the current synthesizer view state.
func (a *App) View() View {
	return a.view
}

// Audio returns the synthesizer audio host.
func (a *App) Audio() *AudioHost {
	return a.audio
}

// Redraw parses the HTML template and renders the page display.
func (a *App) Redraw(ctx context.Context) error {
	return a.page.Redraw(ctx)
}

// Close cleans up audio resources.
func (a *App) Close() {
	if a.audio != nil {
		a.audio.Close()
	}
}

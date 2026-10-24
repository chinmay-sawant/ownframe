package main

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/chinmay-sawant/ownframe"
)

//go:embed animations.html
var animationsHTML string

const (
	defaultWidth  = 1140
	defaultHeight = 900
	togglePlayID  = "toggle-play"
)

// ViewState is the complete template data model. Every card's motion is a CSS
// declaration in animations.html, so the only runtime state is the shared play
// switch.
type ViewState struct {
	IsPlaying bool
}

// App manages the CSS animation showcase. The automatic page runtime drives
// every card from its declared @keyframes; the app holds view state and the one
// click handler.
type App struct {
	page *ownframe.Page
	view ViewState
}

// newApp constructs the gallery and publishes its initial state.
func newApp(devtools bool) (*App, error) {
	app := &App{view: ViewState{IsPlaying: true}}

	p, err := ownframe.New(ownframe.Config{
		Title:  "CSS Animation Showcase",
		HTML:   animationsHTML,
		Width:  defaultWidth,
		Height: defaultHeight,
		// The dock is open by default so the Frame, Elements, and Ops panels
		// are one key away; pass -dev=false to close it.
		DevTools: devtools,
		// The 3D and perspective cards need the retained box tree for the
		// re-projection pass, so keep it from the first layout.
		RetainState: true,
	})
	if err != nil {
		return nil, fmt.Errorf("create animations page: %w", err)
	}

	app.page = p
	p.SetData(&app.view)
	if err := p.Redraw(context.Background()); err != nil {
		return nil, fmt.Errorf("initial redraw: %w", err)
	}

	p.Handle(ownframe.Handlers{
		Click: app.handleClick,
	})

	return app, nil
}

// Page returns the underlying ownframe.Page.
func (a *App) Page() *ownframe.Page {
	return a.page
}

// handleClick toggles the shared play switch. The button only flips the view;
// the body's is-paused class reaches every card through the stylesheet rule.
func (a *App) handleClick(ctx context.Context, box ownframe.Box) error {
	if box.ID != togglePlayID {
		return nil
	}

	a.view.IsPlaying = !a.view.IsPlaying
	a.page.SetData(&a.view)

	return a.page.Redraw(ctx)
}

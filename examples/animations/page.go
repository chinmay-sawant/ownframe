package main

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

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
// declaration in animations.html; playback selects all cards or one card.
type ViewState struct {
	IsPlaying bool
	Solo      string
	Reset     bool
}

// App manages the CSS animation showcase. The automatic page runtime drives
// every card from its declared @keyframes; the app controls playback selection.
type App struct {
	page *ownframe.Page
	view ViewState
}

// newApp constructs the gallery and publishes its initial state.
func newApp(devtools, performance bool) (*App, error) {
	app := &App{view: ViewState{IsPlaying: true}}

	p, err := ownframe.New(ownframe.Config{
		Title:  "CSS Animation Showcase",
		HTML:   animationsHTML,
		Width:  defaultWidth,
		Height: defaultHeight,
		// The dock is open by default so the Frame, Elements, and Ops panels
		// are one key away; pass -dev=false to close it.
		DevTools: devtools,
		// The Frame tab's PERFORMANCE, PIPELINE DETAIL, and MEMORY rows read
		// "-" until the page records timings. This is the gallery's own
		// profiling view, so it is on unless -performance=false.
		Perf: performance,
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

// Remove the selected track in the click batch, then rebind it on the frame
// tick. Two redraws inside a click would be merged by native pointer batching.
func (a *App) handleClick(ctx context.Context, box ownframe.Box) error {
	if box.ID == togglePlayID {
		a.view.IsPlaying = !a.view.IsPlaying
		a.view.Solo = ""
		a.view.Reset = false
		a.page.SetTick(nil)
		return a.redraw(ctx)
	}
	target, ok := strings.CutPrefix(box.ID, "play-")
	if !ok {
		return nil
	}
	for _, candidate := range a.page.Boxes() {
		if candidate.ID != target || len(candidate.Animations) == 0 {
			continue
		}
		a.view.IsPlaying = false
		a.view.Solo = target
		a.view.Reset = true
		a.page.SetTick(a.finishReplay)
		return a.redraw(ctx)
	}
	return nil
}

func (a *App) finishReplay(ctx context.Context) error {
	a.view.Reset = false
	a.page.SetTick(nil)
	return a.redraw(ctx)
}

func (a *App) redraw(ctx context.Context) error {
	a.page.SetData(&a.view)
	return a.page.Redraw(ctx)
}

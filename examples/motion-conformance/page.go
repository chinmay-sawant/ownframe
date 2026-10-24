package main

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/chinmay-sawant/ownframe"
)

//go:embed index.html
var indexHTML string

const (
	defaultWidth  = 1200
	defaultHeight = 900

	colourToggleID    = "toggle-transition-colour"
	opacityToggleID   = "toggle-transition-opacity"
	transformToggleID = "toggle-transition-transform"
)

// transitionState is the class-toggle state the three transition cards print.
// A button flips one field and the class reaches the card through one SetData
// and Redraw; the transition itself is CSS.
type transitionState struct {
	ColourOn    bool
	OpacityOn   bool
	TransformOn bool
}

// newPage builds the conformance page. Every animation and transition is
// declared in index.html; the page registers no tick callback and builds no
// player. The three transition cards add a button that toggles one class, the
// same class-toggle shape the css-preview catalog cards use.
//
// RetainState is on from the first layout because the 3D and perspective cards
// need the retained box tree for the re-projection pass. A page that starts
// without it pays one owed redraw before the first 3D sample moves.
func newPage() (*ownframe.Page, error) {
	state := &transitionState{}

	p, err := ownframe.New(ownframe.Config{
		Title:       "CSS motion conformance",
		HTML:        indexHTML,
		Width:       defaultWidth,
		Height:      defaultHeight,
		RetainState: true,
	})
	if err != nil {
		return nil, fmt.Errorf("create motion-conformance page: %w", err)
	}

	p.SetData(state)
	if err := p.Redraw(context.Background()); err != nil {
		return nil, fmt.Errorf("initial redraw: %w", err)
	}

	p.Handle(ownframe.Handlers{
		Click: func(ctx context.Context, box ownframe.Box) error {
			return toggleTransition(ctx, p, state, box.ID)
		},
	})

	return p, nil
}

// toggleTransition flips the class one transition card's button controls and
// publishes it. The handler changes no paint directly: the class change is the
// style change the automatic transition driver captures. Other clicks pass
// through.
func toggleTransition(ctx context.Context, p *ownframe.Page, state *transitionState, id string) error {
	switch id {
	case colourToggleID:
		state.ColourOn = !state.ColourOn
	case opacityToggleID:
		state.OpacityOn = !state.OpacityOn
	case transformToggleID:
		state.TransformOn = !state.TransformOn
	default:
		return nil
	}

	p.SetData(state)

	return p.Redraw(ctx)
}

package page

import "github.com/chinmay-sawant/ownframe/internal/render"

// renderState is the input, theme, and image state Redraw and PNG share.
func (p *Page) renderState() render.State {
	state := render.State{Hover: p.hoverID(), Active: p.activeID(), Theme: p.theme, Images: p.imageBytes}
	state.RetainState = p.retainState
	if p.formState() != nil {
		state.Focus = p.formState().focusID
	}

	return state
}

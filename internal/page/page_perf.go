package page

import (
	"github.com/chinmay-sawant/ownframe/internal/page/draw"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

// renderState is the input, theme, and image state Redraw and PNG share.
func (p *Page) renderState() render.State {
	focus := ""
	if st := p.formState(); st != nil {
		focus = st.focusID
	}

	return draw.RenderState(p.hoverID(), p.activeID(), focus, p.theme, p.imageBytes, p.retainState)
}

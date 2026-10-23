package page

import (
	"context"

	"github.com/chinmay-sawant/blinkless/css"

	"github.com/chinmay-sawant/ownframe/internal/page/draw"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

// styledDocument returns a styled document for source. When the executed
// source is unchanged the cached tree is reused and only the viewport and the
// pointer state change, so the sheets are not recollected. A fresh cache
// counts one parse and one cascade.
func (p *Page) styledDocument(
	ctx context.Context,
	source string,
	state render.State,
) (*css.Document, error) {
	styled, err := draw.Style(ctx, p.cache, source, p.viewport(), state)
	if err != nil {
		return nil, err
	}

	p.cache = styled.Cache
	if styled.Fresh {
		p.statePseudos = styled.Pseudos
		p.stats.parses++
		p.stats.cascades++
	}

	return styled.Doc, nil
}

// invalidateCache drops the parsed tree so the next Redraw rebuilds it.
func (p *Page) invalidateCache() {
	p.cache = nil
	// Unknown sheets keep every state redraw until the next Redraw reports
	// what the new sheets can match.
	p.statePseudos = draw.UnknownPseudos()
}

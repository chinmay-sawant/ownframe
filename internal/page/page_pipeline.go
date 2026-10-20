package page

import (
	"context"

	"github.com/chinmay-sawant/blinkless/css"

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
	if p.cache != nil && p.cache.Source() == source {
		return p.cache.Relayout(ctx, p.width, p.height, state)
	}

	cache, err := render.NewCache(ctx, source, p.width, p.height, state)
	if err != nil {
		return nil, err
	}

	p.cache = cache
	p.statePseudos = cache.StatePseudos()
	p.stats.parses++
	p.stats.cascades++

	return cache.Styled(), nil
}

// invalidateCache drops the parsed tree so the next Redraw rebuilds it.
func (p *Page) invalidateCache() {
	p.cache = nil
	// Unknown sheets keep every state redraw until the next Redraw reports
	// what the new sheets can match.
	p.statePseudos = render.StatePseudos{Hover: true, Active: true, Focus: true}
}

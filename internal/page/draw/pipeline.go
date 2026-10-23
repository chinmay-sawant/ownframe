package draw

import (
	"context"

	"github.com/chinmay-sawant/blinkless/css"
	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

// Styled is a styled document and the cache it came from.
// Fresh means this call parsed and cascaded, so the caller counts both.
type Styled struct {
	Doc     *css.Document
	Cache   *render.Cache
	Fresh   bool
	Pseudos render.StatePseudos
}

// Style reuses cache when its source matches and only the viewport and the
// pointer state change, so the sheets are not recollected. A miss parses
// and cascades once.
func Style(
	ctx context.Context,
	cache *render.Cache,
	source string,
	view Viewport,
	state render.State,
) (Styled, error) {
	if cache != nil && cache.Source() == source {
		doc, err := cache.Relayout(ctx, view.Width, view.Height, state)
		if err != nil {
			return Styled{}, err
		}

		return Styled{Doc: doc, Cache: cache}, nil
	}

	next, err := render.NewCache(ctx, source, view.Width, view.Height, state)
	if err != nil {
		return Styled{}, err
	}

	return Styled{
		Doc:     next.Styled(),
		Cache:   next,
		Fresh:   true,
		Pseudos: next.StatePseudos(),
	}, nil
}

// UnknownPseudos keeps every interactive state when the sheets are not
// known yet. The next style replaces it with what the document can match.
func UnknownPseudos() render.StatePseudos {
	return render.StatePseudos{Hover: true, Active: true, Focus: true}
}

// List lays a styled document out with the state's image resolver.
func List(ctx context.Context, styled *css.Document, state render.State) (*layout.Display, error) {
	return render.DisplayListDocumentOptions(ctx, styled, layout.Options{
		Images:      state.Images,
		RetainState: state.RetainState,
	})
}

// RenderState is the pointer, theme, and image input Redraw and PNG share.
func RenderState(
	hover, active, focus string,
	theme *css.Sheet,
	images func(string) ([]byte, error),
	retain bool,
) render.State {
	return render.State{
		Hover:       hover,
		Active:      active,
		Focus:       focus,
		Theme:       theme,
		Images:      images,
		RetainState: retain,
	}
}

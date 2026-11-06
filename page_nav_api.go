package ownframe

import (
	"context"
	"fmt"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

// HTML returns the template source at the current history index.
func (p *Page) HTML() string { return innerPage(p).HTML() }

// LoadOptions controls whether Load navigates or updates the current document.
type LoadOptions struct {
	// ReplaceCurrent keeps the current history index and forward entries.
	// Controls with unchanged IDs and types keep user values and focus.
	ReplaceCurrent bool
	// AdoptTree records html as the current source after a caller has already
	// patched the cached tree to match it. It never parses or draws. The
	// builder uses it after an incremental update. The source must be a
	// builder document.
	AdoptTree bool
}

// Load parses html and draws it. By default it adds a navigation entry.
// Supply one LoadOptions value to replace the current document instead.
func (p *Page) Load(ctx context.Context, html string, options ...LoadOptions) error {
	if len(options) > 1 {
		return fmt.Errorf("ownframe: Load accepts at most one options value")
	}
	if len(options) == 1 && options[0].AdoptTree {
		return page.AdoptPatchedSource(innerPage(p), html)
	}
	if len(options) == 1 && options[0].ReplaceCurrent {
		return page.ReplaceHTML(ctx, innerPage(p), html)
	}
	return innerPage(p).Load(ctx, html)
}

// Back draws the previous template source.
func (p *Page) Back(ctx context.Context) error {
	return innerPage(p).Back(ctx)
}

// Forward draws the next template source.
func (p *Page) Forward(ctx context.Context) error {
	return innerPage(p).Forward(ctx)
}

// Route maps a data-action value to HTML for Click to load.
func (p *Page) Route(action, html string) error {
	return innerPage(p).Route(action, html)
}

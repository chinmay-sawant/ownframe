package ownframe

import (
	"context"
	"fmt"
	"github.com/chinmay-sawant/ownframe/internal/page"
)

// HTML returns the template source at the current history index.
func (p *Page) HTML() string {
	inner := innerPage(p)
	if inner == nil {
		return ""
	}

	return inner.HTML()
}

// LoadOptions controls how Load interprets HTML and updates history.
type LoadOptions struct {
	// ReplaceCurrent keeps the current history index and forward entries.
	// Controls with unchanged IDs and types keep user values and focus.
	ReplaceCurrent bool
	// StaticHTML renders the source without executing Go template actions.
	StaticHTML bool
	// AdoptTree records html after a caller has patched the cached tree
	// to match it. It replaces the current entry without parsing or drawing
	// and requires StaticHTML.
	AdoptTree bool
}

// Load parses html as a Go template, adds a navigation entry, and draws it.
// LoadOptions can opt into literal HTML or replace the current entry.
func (p *Page) Load(ctx context.Context, html string, options ...LoadOptions) error {
	if len(options) > 1 {
		return fmt.Errorf("ownframe: Load accepts at most one options value")
	}
	inner := innerPage(p)
	if inner == nil {
		return ErrNilPage
	}
	var opts LoadOptions
	if len(options) == 1 {
		opts = options[0]
	}
	switch {
	case opts.AdoptTree:
		if !opts.StaticHTML {
			return fmt.Errorf("ownframe: Load with AdoptTree requires StaticHTML")
		}
		return page.AdoptPatchedSource(inner, html)
	case opts.ReplaceCurrent:
		return page.ReplaceHTML(ctx, inner, html, opts.StaticHTML)
	default:
		return page.LoadHTML(ctx, inner, html, opts.StaticHTML)
	}
}

// Back draws the previous template source.
func (p *Page) Back(ctx context.Context) error {
	inner := innerPage(p)
	if inner == nil {
		return ErrNilPage
	}

	return inner.Back(ctx)
}

// Forward draws the next template source.
func (p *Page) Forward(ctx context.Context) error {
	inner := innerPage(p)
	if inner == nil {
		return ErrNilPage
	}

	return inner.Forward(ctx)
}

// Route maps a data-action value to HTML for Click to load.
func (p *Page) Route(action, html string) error {
	inner := innerPage(p)
	if inner == nil {
		return ErrNilPage
	}

	return inner.Route(action, html)
}

var (
	// ErrNoHistory means Back or Forward has no entry.
	ErrNoHistory = page.ErrNoHistory
	// ErrInvalidLink means a dispatched or clicked URL is malformed.
	ErrInvalidLink = page.ErrInvalidLink
	// ErrUnsupportedLinkScheme means the scheme is outside its allowlist.
	ErrUnsupportedLinkScheme = page.ErrUnsupportedLinkScheme
	// ErrBadLinkScheme means Config contains a blocked or malformed scheme.
	ErrBadLinkScheme = page.ErrBadLinkScheme
	// ErrNoDeepLinkHandler means no incoming URL callback is registered.
	ErrNoDeepLinkHandler = page.ErrNoDeepLinkHandler
)

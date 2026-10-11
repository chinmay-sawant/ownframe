package page

import "context"

// PrepareMobile prepares a page for its native mobile viewport.
func PrepareMobile(ctx context.Context, p *Page) error {
	if ctx == nil {
		return errNilContext
	}
	if p == nil {
		return ErrNilPage
	}
	p.mobileViewport = true
	return Prepare(ctx, p)
}

package page

import (
	"context"
	"errors"
	"html/template"
	"strings"
)

// ErrNoHistory means Back or Forward has no entry.
var ErrNoHistory = errors.New("ownframe: no history")

// HTML returns the template source at the current history index.
func (p *Page) HTML() string { return p.nav.current() }

// Load parses html, drops later history, and draws that page. html is
// trusted author content: Go template actions in it may read fields and
// methods of the page data. ownframe does not sandbox template execution.
func (p *Page) Load(ctx context.Context, html string) error {
	if err := useContext(ctx); err != nil {
		return err
	}
	if len(html) > MaxSourceSize {
		return ErrSourceTooLarge
	}
	if strings.TrimSpace(html) == "" {
		return ErrEmptyHTML
	}
	tpl, err := template.New("page").Parse(html)
	if err != nil {
		return err
	}
	p.nav.push(html)
	p.tpl = tpl
	p.invalidateAccessibilityDocument()
	p.invalidateCache()
	p.markFull()
	return p.Redraw(ctx)
}

// Back draws the previous template source.
func (p *Page) Back(ctx context.Context) error {
	if !p.nav.canBack() {
		return ErrNoHistory
	}
	return p.drawAt(ctx, p.nav.backIndex())
}

// Forward draws the next template source.
func (p *Page) Forward(ctx context.Context) error {
	if !p.nav.canForward() {
		return ErrNoHistory
	}
	return p.drawAt(ctx, p.nav.forwardIndex())
}

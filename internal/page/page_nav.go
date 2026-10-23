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
func (p *Page) HTML() string { return p.nav.Current() }

// Load parses html, drops later history, and draws that page. html is
// trusted author content: Go template actions in it may read fields and
// methods of the page data. ownframe does not sandbox template execution.
func (p *Page) Load(ctx context.Context, html string) error {
	if err := useContext(ctx); err != nil {
		return err
	}
	if len(html) > MaxSourceSize {
		return sourceTooLarge("html", int64(len(html)))
	}
	if strings.TrimSpace(html) == "" {
		return ErrEmptyHTML
	}
	if err := checkDOMLimits(html); err != nil {
		return err
	}
	tpl, err := template.New("page").Parse(html)
	if err != nil {
		return err
	}
	p.nav.Push(html)
	p.tpl = tpl
	p.invalidateAccessibilityDocument()
	p.invalidateCache()
	p.markFull()
	return p.Redraw(ctx)
}

// Back draws the previous template source.
func (p *Page) Back(ctx context.Context) error {
	if err := useContext(ctx); err != nil {
		return err
	}
	if !p.nav.CanBack() {
		return ErrNoHistory
	}
	return p.drawAt(ctx, p.nav.BackIndex())
}

// Forward draws the next template source.
func (p *Page) Forward(ctx context.Context) error {
	if err := useContext(ctx); err != nil {
		return err
	}
	if !p.nav.CanForward() {
		return ErrNoHistory
	}
	return p.drawAt(ctx, p.nav.ForwardIndex())
}

// Route maps a data-action value to HTML for Click to load. html is
// trusted author content parsed as a Go template at load time; ownframe
// does not sandbox template execution. An empty action is ignored. The
// same action replaces its HTML.
func (p *Page) Route(action, html string) error {
	if action == "" {
		return nil
	}
	if len(html) > MaxSourceSize {
		return sourceTooLarge("route:"+action, int64(len(html)))
	}
	if err := checkDOMLimits(html); err != nil {
		return err
	}
	p.nav.Route(action, html)
	return nil
}

func (p *Page) drawAt(ctx context.Context, at int) error {
	if err := useContext(ctx); err != nil {
		return err
	}
	tpl, err := template.New("page").Parse(p.nav.SourceAt(at))
	if err != nil {
		return err
	}
	p.nav.GoTo(at)
	p.tpl = tpl
	p.invalidateAccessibilityDocument()
	p.invalidateCache()
	p.markFull()
	return p.Redraw(ctx)
}

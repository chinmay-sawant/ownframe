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
func (p *Page) HTML() string {
	return p.past[p.pastAt]
}

// Load parses html, drops later history, and draws that page.
func (p *Page) Load(ctx context.Context, html string) error {
	if err := useContext(ctx); err != nil {
		return err
	}

	if strings.TrimSpace(html) == "" {
		return ErrEmptyHTML
	}

	tpl, err := template.New("page").Parse(html)
	if err != nil {
		return err
	}

	p.past = append(p.past[:p.pastAt+1], html)
	p.pastAt = len(p.past) - 1
	p.tpl = tpl
	p.invalidateAccessibilityDocument()
	p.invalidateCache()
	p.markFull()

	return p.Redraw(ctx)
}

// Back draws the previous template source.
func (p *Page) Back(ctx context.Context) error {
	if p.pastAt == 0 {
		return ErrNoHistory
	}

	return p.drawAt(ctx, p.pastAt-1)
}

// Forward draws the next template source.
func (p *Page) Forward(ctx context.Context) error {
	if p.pastAt+1 >= len(p.past) {
		return ErrNoHistory
	}

	return p.drawAt(ctx, p.pastAt+1)
}

// Route maps a data-action value to HTML for Click to load.
// An empty action is ignored. The same action replaces its HTML.
func (p *Page) Route(action, html string) {
	if action == "" {
		return
	}

	if p.routes == nil {
		p.routes = map[string]string{}
	}

	p.routes[action] = html
}

func (p *Page) drawAt(ctx context.Context, at int) error {
	if err := useContext(ctx); err != nil {
		return err
	}

	tpl, err := template.New("page").Parse(p.past[at])
	if err != nil {
		return err
	}

	p.pastAt = at
	p.tpl = tpl
	p.invalidateAccessibilityDocument()
	p.invalidateCache()
	p.markFull()

	return p.Redraw(ctx)
}

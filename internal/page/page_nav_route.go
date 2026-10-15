package page

import (
	"context"
	"html/template"
)

// Route maps a data-action value to HTML for Click to load. html is
// trusted author content parsed as a Go template at load time; ownframe
// does not sandbox template execution. An empty action is ignored. The
// same action replaces its HTML.
func (p *Page) Route(action, html string) error {
	if action == "" {
		return nil
	}
	if len(html) > MaxSourceSize {
		return ErrSourceTooLarge
	}
	p.nav.route(action, html)
	return nil
}

func (p *Page) drawAt(ctx context.Context, at int) error {
	if err := useContext(ctx); err != nil {
		return err
	}
	tpl, err := template.New("page").Parse(p.nav.sourceAt(at))
	if err != nil {
		return err
	}
	p.nav.goTo(at)
	p.tpl = tpl
	p.invalidateAccessibilityDocument()
	p.invalidateCache()
	p.markFull()
	return p.Redraw(ctx)
}

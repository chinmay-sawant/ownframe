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
	p.markFormDirty()
	p.invalidateAccessibilityDocument()
	p.invalidateCache()
	p.markFull()
	return p.Redraw(ctx)
}

// ReplaceHTML updates a document without adding or dropping history entries.
// Like file reload, stable control IDs keep live values and focus.
func ReplaceHTML(ctx context.Context, p *Page, source string) error {
	if err := useContext(ctx); err != nil {
		return err
	}
	if len(source) > MaxSourceSize {
		return sourceTooLarge("html", int64(len(source)))
	}
	if strings.TrimSpace(source) == "" {
		return ErrEmptyHTML
	}
	if err := checkDOMLimits(source); err != nil {
		return err
	}
	tpl, err := template.New("page").Parse(source)
	if err != nil {
		return err
	}
	previous, previousSource := p.tpl, p.nav.Current()
	previousFrame, previousBoxes := p.draw, p.boxes
	previousBody, previousSynced := p.source, p.lastSyncBody
	previousForm := p.formState()
	previousRequest, previousHasRequest := p.ff.requestFocus, p.ff.hasFocusRequest
	if previousForm != nil {
		next := *previousForm
		p.setFormState(&next)
	}
	p.tpl = tpl
	p.nav.Retarget(source)
	p.invalidateAccessibilityDocument()
	p.markFormDirty()
	p.invalidateCache()
	if err := p.Redraw(ctx); err != nil {
		p.tpl = previous
		p.draw, p.boxes = previousFrame, previousBoxes
		p.source, p.lastSyncBody = previousBody, previousSynced
		p.nav.Retarget(previousSource)
		p.setFormState(previousForm)
		p.ff.requestFocus, p.ff.hasFocusRequest = previousRequest, previousHasRequest
		p.markFormDirty()
		p.invalidateCache()
		return err
	}
	if p.redrawBatch == 0 {
		p.resolveStates()
	}
	return nil
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
	p.markFormDirty()
	p.invalidateAccessibilityDocument()
	p.invalidateCache()
	p.markFull()
	return p.Redraw(ctx)
}

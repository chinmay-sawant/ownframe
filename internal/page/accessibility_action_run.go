package page

import (
	"context"
	"unicode/utf8"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

func (p *Page) accessibilityActivate(ctx context.Context, id string) error {
	c, ok := p.control(id)
	if !ok || c.Disabled {
		return host.ErrAccessibilityDisabled
	}
	box := p.boxByID(id)
	if c.Tag == "button" && box.Action != "" {
		if html, ok := p.nav.Lookup(box.Action); ok {
			return p.Load(ctx, html)
		}
	}
	return p.clickControl(ctx, box)
}

func (p *Page) accessibilitySetValue(ctx context.Context, id, value string) error {
	if !utf8.ValidString(value) {
		return host.ErrAccessibilityRange
	}
	c, ok := p.control(id)
	if !ok || c.Disabled {
		return host.ErrAccessibilityDisabled
	}
	if !canEdit(c) || c.Type == "file" {
		return host.ErrAccessibilityAction
	}
	if err := p.Focus(ctx, id); err != nil {
		return err
	}
	if p.FocusID() != id {
		return host.ErrAccessibilityDisabled
	}
	c, _ = p.control(id)
	return p.IMEReplace(ctx, 0, len(c.Value), value, len(value))
}

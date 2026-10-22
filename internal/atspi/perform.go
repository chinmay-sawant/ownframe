package atspi

import (
	"context"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// Perform runs an AT-SPI action against the page on the caller's loop.
func (p *Publisher) Perform(ctx context.Context, acc host.Accessible, path, action, value string) error {
	if acc == nil {
		return host.ErrAccessibilityUnavailable
	}
	obj, ok := p.lookup(path)
	if !ok {
		return host.ErrAccessibilityNode
	}
	kind, ok := ActionKind(action)
	if !ok {
		return host.ErrAccessibilityAction
	}
	if obj.Password {
		value = ""
	}
	return acc.AccessibilityPerform(ctx, host.AccessibilityRequest{
		Document: p.doc,
		Node:     obj.ID,
		Kind:     kind,
		Value:    value,
	})
}

func (p *Publisher) lookup(path string) (Object, bool) {
	for _, obj := range p.app.Objects {
		if obj.Path == path {
			return obj, true
		}
	}
	return Object{}, false
}

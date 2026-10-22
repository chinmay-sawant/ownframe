package page

import (
	"context"

	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/page/a11y"
)

// AccessibilityPerform runs an advertised action on the page-owning loop.
func (p *Page) AccessibilityPerform(ctx context.Context, request host.AccessibilityRequest) error {
	if err := useContext(ctx); err != nil {
		return err
	}
	tree, err := p.AccessibilityTree()
	if err != nil {
		return err
	}
	if request.Document != tree.Document {
		return host.ErrAccessibilityStale
	}
	node, ok := a11y.Lookup(tree, request.Node)
	if !ok {
		return host.ErrAccessibilityNode
	}
	if node.Disabled && (request.Kind != host.AccessibilityFocus || !node.Focusable) {
		return host.ErrAccessibilityDisabled
	}
	if !a11y.HasAction(node, request.Kind) {
		return host.ErrAccessibilityAction
	}

	switch request.Kind {
	case host.AccessibilityFocus:
		if err := p.Focus(ctx, node.ElementID); err != nil {
			return err
		}
		if p.FocusID() != node.ElementID {
			return host.ErrAccessibilityDisabled
		}
		return nil
	case host.AccessibilityActivate:
		return p.accessibilityActivate(ctx, node.ElementID)
	case host.AccessibilitySetValue:
		return p.accessibilitySetValue(ctx, node.ElementID, request.Value)
	default:
		return host.ErrAccessibilityAction
	}
}

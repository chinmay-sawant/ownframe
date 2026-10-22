package a11y

import (
	"context"
	"unicode/utf8"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/host"
)

// Actor is the page behavior one accessibility action runs.
type Actor interface {
	Focus(ctx context.Context, id string) error
	FocusID() string
	Control(id string) (Control, bool)
	BoxByID(id string) layout.Box
	Lookup(action string) (string, bool)
	Load(ctx context.Context, html string) error
	ClickControl(ctx context.Context, box layout.Box) error
	IMEReplace(ctx context.Context, start, end int, insert string, caret int) error
}

// Perform runs one advertised action against tree.
// The caller has already checked the context and loaded the current tree.
func Perform(ctx context.Context, tree host.AccessibilityTree, request host.AccessibilityRequest, actor Actor) error {
	if request.Document != tree.Document {
		return host.ErrAccessibilityStale
	}
	node, ok := Lookup(tree, request.Node)
	if !ok {
		return host.ErrAccessibilityNode
	}
	allowsFocus := request.Kind == host.AccessibilityFocus && node.Focusable
	if node.Disabled && !allowsFocus {
		return host.ErrAccessibilityDisabled
	}
	if !HasAction(node, request.Kind) {
		return host.ErrAccessibilityAction
	}

	switch request.Kind {
	case host.AccessibilityFocus:
		return focus(ctx, actor, node.ElementID)
	case host.AccessibilityActivate:
		return activate(ctx, actor, node.ElementID)
	case host.AccessibilitySetValue:
		return setValue(ctx, actor, node.ElementID, request.Value)
	default:
		return host.ErrAccessibilityAction
	}
}

func focus(ctx context.Context, actor Actor, id string) error {
	if err := actor.Focus(ctx, id); err != nil {
		return err
	}
	if actor.FocusID() != id {
		return host.ErrAccessibilityDisabled
	}
	return nil
}

func activate(ctx context.Context, actor Actor, id string) error {
	c, ok := actor.Control(id)
	if !ok || c.Disabled {
		return host.ErrAccessibilityDisabled
	}
	box := actor.BoxByID(id)
	if c.Tag == "button" && box.Action != "" {
		if html, ok := actor.Lookup(box.Action); ok {
			return actor.Load(ctx, html)
		}
	}
	return actor.ClickControl(ctx, box)
}

func setValue(ctx context.Context, actor Actor, id, value string) error {
	if !utf8.ValidString(value) {
		return host.ErrAccessibilityRange
	}
	c, ok := actor.Control(id)
	if !ok || c.Disabled {
		return host.ErrAccessibilityDisabled
	}
	if !c.Editable || c.Type == "file" {
		return host.ErrAccessibilityAction
	}
	if err := actor.Focus(ctx, id); err != nil {
		return err
	}
	if actor.FocusID() != id {
		return host.ErrAccessibilityDisabled
	}
	c, _ = actor.Control(id)
	return actor.IMEReplace(
		ctx,
		0,
		len(c.Value),
		value,
		len(value),
	)
}

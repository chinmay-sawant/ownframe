package pagetest

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// Click hits the centre of the box with id.
func Click(ctx context.Context, p *ownframe.Page, id string) error {
	box, ok := BoxByID(p, id)
	if !ok {
		return ErrNoBox
	}
	return p.Click(ctx, box.X+box.W/2, box.Y+box.H/2)
}

// ClickAction hits the centre of the box with data-action.
func ClickAction(ctx context.Context, p *ownframe.Page, action string) error {
	box, ok := BoxByAction(p, action)
	if !ok {
		return ErrNoBox
	}
	return p.Click(ctx, box.X+box.W/2, box.Y+box.H/2)
}

// Type appends text to the focused field.
func Type(ctx context.Context, p *ownframe.Page, text string) error {
	if p == nil {
		return ErrNoBox
	}
	return p.Edit().Type(ctx, text)
}

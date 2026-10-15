package render

import (
	"context"
	"fmt"

	"github.com/chinmay-sawant/blinkless/css"
)

// Styled returns the styled document for the cached tree.
func (c *Cache) Styled() *css.Document {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.styled
}

// Relayout re-cascades the cached sheets for a new viewport and pointer
// state. It never parses the HTML or recollects the sheets.
func (c *Cache) Relayout(
	ctx context.Context,
	width, height int,
	state State,
) (*css.Document, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	engineMu.Lock()
	defer engineMu.Unlock()

	styled, err := css.Relayout(ctx, c.styled, width, height, state.Focus, state.Hover, state.Active)
	if err != nil {
		return nil, fmt.Errorf("render: relayout: %w", err)
	}

	c.styled = styled

	return styled, nil
}

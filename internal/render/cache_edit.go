package render

import (
	"context"
	"errors"

	"github.com/chinmay-sawant/blinkless/css"
)

// SetAttribute edits an attribute of an element in the cached tree and reports
// whether the element exists. An empty name changes nothing and reports
// false. The edit lands on the same tree the cascade read, so the next
// relayout sees it.
func (c *Cache) SetAttribute(elementID, name, value string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.doc == nil || elementID == "" || name == "" {
		return false
	}

	return c.doc.SetAttribute(elementID, name, value)
}

// SetText replaces the text content of an element in the cached tree and
// reports whether the element exists.
//
// The text becomes the element's only child and every child it had is
// dropped, the way the DOM textContent setter works. An empty text leaves
// the element without children. The edit lands on the same tree the cascade
// read, so the next relayout sees it.
func (c *Cache) SetText(elementID, text string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.doc == nil || elementID == "" {
		return false
	}

	return c.doc.SetText(elementID, text)
}

// RelayoutFor re-collects the sheets for the cached tree at the given
// viewport and state and returns the styled document. The cached one stays
// until the caller commits it with SetStyled, so a caller that refuses the
// result keeps the page it had.
func (c *Cache) RelayoutFor(ctx context.Context, width, height int, state State) (*css.Document, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.styled == nil {
		return nil, errors.New("render: no styled document")
	}

	engineMu.Lock()
	defer engineMu.Unlock()

	styled, err := css.Relayout(ctx, c.styled, width, height, state.Focus, state.Hover, state.Active)
	if err != nil {
		return nil, err
	}

	return styled, nil
}

// SetStyled replaces the cached styled document.
func (c *Cache) SetStyled(styled *css.Document) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.styled = styled
}

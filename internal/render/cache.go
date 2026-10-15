package render

import (
	"context"
	"fmt"
	"sync"

	"github.com/chinmay-sawant/blinkless/css"
	"github.com/chinmay-sawant/blinkless/html"
)

// Cache holds one parsed HTML tree and the styled document built from it.
// The first draw parses and cascades once; every later draw relayouts the
// cached document at the current viewport and pointer state, so a size or a
// state change never parses the HTML or recollects the stylesheets. A caller
// drops the cache when the executed source or the theme changes.
//
// Concurrency contract: a Cache is safe for concurrent method calls, which
// serialize through its mutex. Relayout replaces the styled document returned
// by Styled; a caller that still holds a *css.Document from a previous call
// must not read or mutate it after starting a new Relayout. Cache methods do
// not lock the returned *css.Document, so the same Cache must not be shared
// across overlapping Relayout or Styled consumers. Multiple Pages keep their
// own Cache and do not share one.
type Cache struct {
	mu     sync.Mutex
	source string
	doc    *html.Document
	styled *css.Document
}

// NewCache parses source and applies the document sheets once, at the given
// viewport and pointer state.
func NewCache(
	ctx context.Context,
	source string,
	width, height int,
	state State,
) (*Cache, error) {
	engineMu.Lock()
	defer engineMu.Unlock()
	doc, err := html.Parse([]byte(source))
	if err != nil {
		return nil, fmt.Errorf("render: parse: %w", err)
	}

	styled, err := css.Apply(ctx, doc, state.options(width, height))
	if err != nil {
		return nil, fmt.Errorf("render: css: %w", err)
	}

	return &Cache{source: source, doc: doc, styled: styled}, nil
}

// Source returns the executed source the cached tree was parsed from.
func (c *Cache) Source() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.source
}

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

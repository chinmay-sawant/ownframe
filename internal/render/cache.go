package render

import (
	"context"
	"fmt"
	"sync"

	"github.com/chinmay-sawant/blinkless/css"
	"github.com/chinmay-sawant/blinkless/html"
	"github.com/chinmay-sawant/ownframe/internal/fetch"
)

// Cache holds one parsed HTML tree and the styled document built from it.
// The first draw parses and cascades once; every later draw relayouts the
// cached document at the current viewport and pointer state, so a size or a
// state change never parses the HTML. The sheets are collected again on every
// relayout, because a caller can edit the tree. A caller drops the cache when
// the executed source or the theme changes.
//
// Concurrency contract: a Cache is safe for concurrent method calls, which
// serialize through its mutex. Relayout replaces the styled document returned
// by Styled; a holder must not read or mutate an older doc after a Relayout.
// Call Styled and Relayout from the page loop only; never retain the doc
// across a Relayout or share the Cache across concurrent loops. Multiple
// Pages keep their own Cache and do not share one. Tick and Draw stay on the
// same goroutine, so no live race occurs under that single-owner use.
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
	if ctx == nil {
		return nil, fetch.ErrNilContext
	}
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

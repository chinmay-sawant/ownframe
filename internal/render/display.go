package render

import (
	"context"
	"fmt"

	"github.com/chinmay-sawant/blinkless/css"
	"github.com/chinmay-sawant/blinkless/html"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/fetch"
)

// DisplayOp is one vector op from layout under a local name.
// Read payloads via accessors; zero value has none.
type DisplayOp = layout.DisplayOp

// Display is a retained display list plus its canvas size.
type Display = layout.Display

// DisplayList parses source and returns vector ops for width, height.
// Paint stays the entry for pictures, including GET /frame.png.
func DisplayList(ctx context.Context, source string, width, height int) (*Display, error) {
	return DisplayListState(ctx, source, width, height, State{})
}

// DisplayListState is DisplayList with pointer and focus state.
func DisplayListState(ctx context.Context, source string, width, height int, state State) (*Display, error) {
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

	display, err := layout.DisplayListOptions(ctx, styled, layout.Options{Images: state.Images})
	if err != nil {
		return nil, fmt.Errorf("render: display: %w", err)
	}

	return display, nil
}

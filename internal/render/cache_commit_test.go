package render_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/blinkless/css"
	"github.com/chinmay-sawant/blinkless/html"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

// TestCacheRelayoutForDoesNotCommit checks the commit contract: the caller
// keeps the styled document it had until it commits the new one.
func TestCacheRelayoutForDoesNotCommit(t *testing.T) {
	ctx := context.Background()
	cache, err := render.NewCache(ctx, editSource, 320, 200, render.State{})
	if err != nil {
		t.Fatal(err)
	}

	kept := cache.Styled()
	styled, err := cache.RelayoutFor(ctx, 320, 200, render.State{})
	if err != nil {
		t.Fatal(err)
	}
	if cache.Styled() != kept {
		t.Fatal("RelayoutFor committed the styled document")
	}

	cache.SetStyled(styled)
	if cache.Styled() != styled {
		t.Fatal("SetStyled did not commit the styled document")
	}
}

// TestCacheStatePseudosFollowTheStyledDocument checks that the cache reports
// the state rules of the styled document it holds, and not the ones it was
// built with.
func TestCacheStatePseudosFollowTheStyledDocument(t *testing.T) {
	ctx := context.Background()
	cache, err := render.NewCache(ctx, editSource, 320, 200, render.State{})
	if err != nil {
		t.Fatal(err)
	}
	if cache.StatePseudos().Hover {
		t.Fatal("the sheets have no hover rule")
	}

	doc, err := html.Parse([]byte(`<style>#b:hover{width:20px}</style><div id="b"></div>`))
	if err != nil {
		t.Fatal(err)
	}
	styled, err := css.Apply(ctx, doc, css.Options{WidthPx: 320, HeightPx: 200, Media: "screen"})
	if err != nil {
		t.Fatal(err)
	}
	cache.SetStyled(styled)

	if !cache.StatePseudos().Hover {
		t.Fatal("the hover rule of the committed document was not reported")
	}
}

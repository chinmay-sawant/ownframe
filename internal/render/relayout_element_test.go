package render_test

import (
	"context"
	"sync"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

func TestRelayoutElementConcurrentWithParse(t *testing.T) {
	ctx := context.Background()
	cache, err := render.NewCache(ctx, `<p id="p">hi</p>`, 200, 100, render.State{RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	display, err := render.DisplayListDocumentOptions(
		ctx,
		cache.Styled(),
		layout.Options{RetainState: true},
	)
	if err != nil {
		t.Fatal(err)
	}

	styled := cache.Styled()
	var wg sync.WaitGroup
	wg.Go(func() {
		for range 20 {
			_, err := render.NewCache(ctx, `<p id="q">other</p>`, 200, 100, render.State{})
			if err != nil {
				t.Error(err)
				return
			}
		}
	})
	wg.Go(func() {
		for range 20 {
			_, err := render.RelayoutElement(ctx, styled, display, "p", layout.Options{})
			if err != nil {
				t.Error(err)
				return
			}
		}
	})
	wg.Wait()
}

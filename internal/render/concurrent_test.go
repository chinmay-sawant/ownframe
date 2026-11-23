package render_test

import (
	"context"
	"sync"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

func TestConcurrentIndependentPages(t *testing.T) {
	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			for range 3 {
				cache, err := render.NewCache(context.Background(), `<p>Hello</p>`, 200, 100, render.State{})
				if err != nil {
					t.Error(err)
					return
				}
				if _, err := render.DisplayListDocument(context.Background(), cache.Styled(), nil); err != nil {
					t.Error(err)
					return
				}
			}
		})
	}
	wg.Wait()
}

// TestParallelPagesRenderAndRelayoutUnderRace is the HYG-ENG-01 gate. Several
// pages parse, style, relayout, and build display lists at once with no
// package-level lock: blinkless fixed Node.EnsureAttrs and
// prepare.SheetCache concurrency (BLK-CONC-01/02), so the engine calls run
// concurrently under `go test -race ./internal/render ./internal/page`.
func TestParallelPagesRenderAndRelayoutUnderRace(t *testing.T) {
	const source = `<style>
		.card { width: 40px; height: 20px; background: #123 }
		.card:hover { background: #456 }
		@media (min-width: 300px) { .card { width: 80px } }
	</style><div class="card" id="a">a</div><div class="card" id="b">b</div>`

	var wg sync.WaitGroup
	for page := range 6 {
		wg.Go(func() {
			ctx := context.Background()
			cache, err := render.NewCache(ctx, source, 200, 120, render.State{})
			if err != nil {
				t.Error(err)
				return
			}
			for step := range 8 {
				width := 200 + 120*((step+page)%2)
				styled, err := cache.Relayout(ctx, width, 120, render.State{Hover: "a"})
				if err != nil {
					t.Error(err)
					return
				}
				if _, err := render.DisplayListDocument(ctx, styled, nil); err != nil {
					t.Error(err)
					return
				}
				_ = cache.StatePseudos()
			}
		})
	}
	wg.Wait()
}

// TestConcurrentReadOnlyPassesOnOneDocument runs display-list passes over one
// styled document from several goroutines, the access pattern that makes the
// lazy Node.EnsureAttrs initialization race that the removed engine lock used to hide.
func TestConcurrentReadOnlyPassesOnOneDocument(t *testing.T) {
	ctx := context.Background()
	cache, err := render.NewCache(ctx, `<div id="a" class="x"><p id="b" data-k="v">t</p></div>`, 200, 120, render.State{})
	if err != nil {
		t.Fatal(err)
	}
	styled := cache.Styled()
	var wg sync.WaitGroup
	for range 6 {
		wg.Go(func() {
			for range 6 {
				if _, err := render.DisplayListDocument(ctx, styled, nil); err != nil {
					t.Error(err)
					return
				}
			}
		})
	}
	wg.Wait()
}

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

func TestMotionWrappersConcurrentWithParse(t *testing.T) {
	ctx := context.Background()
	src := `<div style="width:180px;height:80px">` +
		`<div id="p" style="width:100px;height:20px">hi</div>` +
		`<svg id="s" width="20" height="20" viewBox="0 0 20 20"><rect id="r" width="10" height="10" fill="#112233"/></svg>` +
		`</div>`
	cache, err := render.NewCache(ctx, src, 200, 100, render.State{RetainState: true})
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
		for range 15 {
			if _, err := render.NewCache(ctx, `<p id="q">other</p>`, 200, 100, render.State{}); err != nil {
				t.Error(err)
				return
			}
		}
	})
	wg.Go(func() {
		for range 15 {
			_, err := render.RelayoutElementWithOverrides(
				ctx,
				styled,
				display,
				"p",
				[]layout.Override{{ElementID: "p", Property: "width", Value: "120px"}},
				layout.Options{RetainState: true},
			)
			if err != nil {
				t.Error(err)
				return
			}
		}
	})
	wg.Go(func() {
		for range 15 {
			_, err := render.RestampTransforms(
				ctx,
				styled,
				display,
				[]layout.TransformOverride{{ElementID: "p", Transform: "translateX(4px)", HasTransform: true}},
				layout.Options{RetainState: true},
			)
			if err != nil {
				t.Error(err)
				return
			}
		}
	})
	wg.Go(func() {
		for range 15 {
			_, err := render.RestyleInlineSVG(
				ctx,
				styled,
				display,
				[]layout.SVGPaintOverride{{ElementID: "s", PartID: "r", Fill: "#445566"}},
				layout.Options{RetainState: true},
			)
			if err != nil {
				t.Error(err)
				return
			}
		}
	})
	wg.Wait()
}

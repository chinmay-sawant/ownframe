package draw

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

func TestPipelineStyleReusesCache(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	view := Viewport{Width: 40, Height: 40}
	first, err := Style(ctx, nil, `<p>a</p>`, view, render.State{})
	if err != nil || !first.Fresh || first.Cache == nil || first.Doc == nil {
		t.Fatal(err)
	}

	second, err := Style(ctx, first.Cache, `<p>a</p>`, view, render.State{})
	if err != nil || second.Fresh || second.Cache != first.Cache || second.Doc == nil {
		t.Fatal("reuse")
	}

	third, err := Style(ctx, first.Cache, `<p>b</p>`, view, render.State{})
	if err != nil || !third.Fresh || third.Cache == first.Cache {
		t.Fatal("fresh")
	}
}

func TestPipelineUnknownPseudos(t *testing.T) {
	t.Parallel()

	got := UnknownPseudos()
	if !got.Hover || !got.Active || !got.Focus {
		t.Fatalf("%+v", got)
	}
}

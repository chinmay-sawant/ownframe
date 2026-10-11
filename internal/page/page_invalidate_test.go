package page_test

import (
	"context"
	"testing"
)

func TestSetDataChangingOutputRebuildsCache(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newCachePage(t, `<p>{{.N}}</p>`)
	p.SetData(struct{ N int }{N: 1})

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	p.SetData(struct{ N int }{N: 2})
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if got := p.Stats().Parses; got != 2 {
		t.Fatalf("parses = %d, want 2", got)
	}
}

func TestSetThemeInvalidatesCache(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newCachePage(t, `<p id="a">static</p>`)

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if err := p.SetTheme(`#a{color:#123456}`); err != nil {
		t.Fatal(err)
	}

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if got := p.Stats().Parses; got != 2 {
		t.Fatalf("parses = %d, want 2", got)
	}
}

func TestNavigationInvalidatesCache(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newCachePage(t, `<p>one</p>`)

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if err := p.Load(ctx, `<p>two</p>`); err != nil {
		t.Fatal(err)
	}

	if err := p.Back(ctx); err != nil {
		t.Fatal(err)
	}

	if err := p.Forward(ctx); err != nil {
		t.Fatal(err)
	}

	if got := p.Stats().Parses; got != 4 {
		t.Fatalf("parses = %d, want 4", got)
	}
}

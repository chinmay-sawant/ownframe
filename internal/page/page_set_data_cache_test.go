package page_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

func TestSetDataKeepsCacheWhenOutputIsUnchanged(t *testing.T) {
	p, err := page.New(page.Config{HTML: `<p>stable</p>`, Width: 320, Height: 200, Perf: true})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	parses := p.Stats().Parses
	p.SetData(struct{ Unused string }{Unused: "new value"})
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if got := p.Stats().Parses; got != parses {
		t.Fatalf("parses = %d after unchanged output, want %d", got, parses)
	}
}

func TestSetDataConcurrentWithRedraw(t *testing.T) {
	t.Parallel()
	p, err := page.New(page.Config{HTML: `<p>{{.Text}}</p>`, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	done := make(chan struct{})
	go func() {
		for i := 0; i < 50; i++ {
			p.SetData(map[string]string{"Text": "hello"})
		}
		close(done)
	}()
	for {
		select {
		case <-done:
			return
		default:
			_ = p.Redraw(ctx)
		}
	}
}


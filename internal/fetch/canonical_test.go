package fetch_test

import (
	"errors"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/fetch"
	"github.com/chinmay-sawant/ownframe/internal/ipc"
	"github.com/chinmay-sawant/ownframe/internal/page"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

func TestNilContextCanonical(t *testing.T) {
	t.Parallel()
	if _, err := fetch.Do(nil, "GET", "https://example.com", nil, nil); !errors.Is(err, fetch.ErrNilContext) { //nolint:staticcheck // intentionally passes nil to verify ErrNilContext
		t.Fatalf("fetch err = %v", err)
	}
	if _, err := ipc.Request(nil, "nil-canon", "x"); !errors.Is(err, fetch.ErrNilContext) { //nolint:staticcheck // intentionally passes nil to verify ErrNilContext
		t.Fatalf("ipc err = %v", err)
	}
	p, err := page.New(page.Config{HTML: "<p>x</p>", Width: 100, Height: 100})
	if err != nil {
		t.Fatal(err)
	}
	if err := page.Prepare(nil, p); !errors.Is(err, fetch.ErrNilContext) { //nolint:staticcheck // intentionally passes nil to verify ErrNilContext
		t.Fatalf("prepare err = %v", err)
	}
	if err := p.Back(nil); !errors.Is(err, fetch.ErrNilContext) {
		t.Fatalf("back err = %v", err)
	}
	if err := p.Forward(nil); !errors.Is(err, fetch.ErrNilContext) {
		t.Fatalf("forward err = %v", err)
	}
	if _, err := render.DisplayListState(nil, "<p>x</p>", 100, 100, render.State{}); !errors.Is(err, fetch.ErrNilContext) {
		t.Fatalf("display err = %v", err)
	}
	if _, err := render.NewCache(nil, "<p>x</p>", 100, 100, render.State{}); !errors.Is(err, fetch.ErrNilContext) {
		t.Fatalf("cache err = %v", err)
	}
}

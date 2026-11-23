package pagetest_test

import (
	"context"
	"errors"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/pagetest"
)

// TestBoxByActionIndexed checks the page's action index answers the helper:
// the later box wins a repeated data-action, and misses and the empty action
// report false.
func TestBoxByActionIndexed(t *testing.T) {
	p, err := ownframe.New(ownframe.Config{
		HTML: `<div id="a" data-action="save" style="width:20px;height:20px">a</div>` +
			`<div id="b" data-action="save" style="width:20px;height:20px">b</div>`,
		Width:  320,
		Height: 200,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	box, ok := pagetest.BoxByAction(p, "save")
	if !ok {
		t.Fatal("save was not found")
	}
	if box.ID != "b" {
		t.Fatalf("BoxByAction(save) = %q, want the later box b", box.ID)
	}

	if _, ok := pagetest.BoxByAction(p, "missing"); ok {
		t.Fatal("a missing action must not resolve")
	}
	if _, ok := pagetest.BoxByAction(p, ""); ok {
		t.Fatal("the empty action must not resolve")
	}
}

// TestBoxByIDOnBitmapFallbackAndNilPage pins HYG-IPC-01: a page that paints
// through the bitmap fallback has no retained display, yet BoxByID still
// finds the box; Type on a nil page reports ErrNilPage.
func TestBoxByIDOnBitmapFallbackAndNilPage(t *testing.T) {
	p, err := ownframe.New(ownframe.Config{
		HTML:   `<style>div { width: 60px; height: 40px; background: #123456; border-radius: 50% / 20%; }</style><div id="ell"></div>`,
		Width:  320,
		Height: 400,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	if p.Probe().Display() != nil {
		t.Skip("elliptical fill no longer falls back to the bitmap path")
	}
	if _, ok := pagetest.BoxByID(p, "ell"); !ok {
		t.Fatal("BoxByID missed the box on a bitmap-fallback page")
	}
	if err := pagetest.Type(context.Background(), nil, "x"); !errors.Is(err, pagetest.ErrNilPage) {
		t.Fatalf("Type(nil) = %v, want ErrNilPage", err)
	}
}

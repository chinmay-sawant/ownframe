package main

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/viewtransition"
)

func TestCatalogViewTransitionNameReachesOwnFrameFrames(t *testing.T) {
	old := catalogViewDisplay(t, `<style>.c{display:block;width:40px;height:20px;background:red;view-transition-name:card}</style><div id="target" class="c">old</div>`)
	next := catalogViewDisplay(t, `<style>.c{display:block;width:100px;height:20px;background:blue;view-transition-name:card}</style><div id="target" class="c">new</div>`)
	tr, err := viewtransition.New(old, next, viewtransition.Scope{OldRoot: -1, NewRoot: -1}, nil)
	if err != nil {
		t.Fatal(err)
	}
	f := tr.At(.5)
	if len(f.Groups) != 1 || f.Groups[0].Name != "card" || f.Groups[0].Bounds.W != 70 || len(f.Groups[0].Old) == 0 || len(f.Groups[0].New) == 0 {
		t.Fatalf("paired frame = %+v", f.Groups)
	}
	display := f.Display()
	want := len(f.Base.Ops) + len(f.Groups[0].Old) + len(f.Groups[0].New)
	if len(display.Ops) != want || len(display.OpOwners) != want || len(display.Order) != want {
		t.Fatalf("render display ops/order/owners=%d/%d/%d, want %d", len(display.Ops), len(display.Order), len(display.OpOwners), want)
	}
}

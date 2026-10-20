package main

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/viewtransition"
)

func TestCatalogViewTransitionClassChangesOwnFrameFrame(t *testing.T) {
	d := catalogViewDisplay(t, `<style>.c{display:block;width:40px;height:20px;background:red;view-transition-name:card;view-transition-class:promo}</style><div id="target" class="c">card</div>`)
	scope := viewtransition.Scope{OldRoot: -1, NewRoot: -1}
	plain, err := viewtransition.New(d, d, scope, nil)
	if err != nil {
		t.Fatal(err)
	}
	alpha, scale := .5, 1.5
	styled, err := viewtransition.New(d, d, scope, map[string]viewtransition.ClassStyle{"promo": {Opacity: &alpha, Scale: &scale}})
	if err != nil {
		t.Fatal(err)
	}
	plainFrame, styledFrame := plain.At(.5), styled.At(.5)
	a, b := plainFrame.Groups[0], styledFrame.Groups[0]
	if len(b.Classes) != 1 || b.Classes[0] != "promo" || b.New[0].W <= a.New[0].W || b.New[0].Opacity() >= a.New[0].Opacity() {
		t.Fatalf("class frames did not differ: plain=%+v styled=%+v", a, b)
	}
	plainDisplay, styledDisplay := plainFrame.Display(), styledFrame.Display()
	if len(styledDisplay.Ops) != len(plainDisplay.Ops) || len(styledDisplay.Order) != len(styledDisplay.Ops) || len(styledDisplay.OpOwners) != len(styledDisplay.Ops) {
		t.Fatalf("class display ops/order/owners = %d/%d/%d, plain ops %d", len(styledDisplay.Ops), len(styledDisplay.Order), len(styledDisplay.OpOwners), len(plainDisplay.Ops))
	}
	if styledDisplay.Ops[len(styledDisplay.Ops)-1].W <= plainDisplay.Ops[len(plainDisplay.Ops)-1].W || styledDisplay.Ops[len(styledDisplay.Ops)-1].Opacity() >= plainDisplay.Ops[len(plainDisplay.Ops)-1].Opacity() {
		t.Fatal("view-transition-class style did not reach composed OwnFrame paint")
	}
}

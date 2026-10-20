package main

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/viewtransition"
)

func TestCatalogViewTransitionScopeLimitsOwnFrameFrames(t *testing.T) {
	scopeValue := "all"
	source := `<style>.outer{view-transition-name:outer}.scope{view-transition-scope:` + scopeValue + `;view-transition-name:scope}.inner{display:block;view-transition-name:inside}.away{view-transition-name:away}</style><main class="outer"><section id="scope" class="scope"><div class="inner">inside</div></section><p class="away">away</p></main>`
	d := catalogViewDisplay(t, source)
	doc, err := viewtransition.New(d, d, viewtransition.Scope{OldRoot: -1, NewRoot: -1}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := doc.Groups(); len(got) != 2 || got[0].Name != "outer" || got[1].Name != "away" {
		t.Fatalf("document transition groups = %+v", got)
	}
	root := catalogViewIndex(d, "scope")
	if root < 0 {
		t.Fatal("scoped element missing from Blinkless display")
	}
	local, err := viewtransition.New(d, d, viewtransition.Scope{OldRoot: root, NewRoot: root}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := local.Groups(); len(got) != 2 || got[0].Name != "scope" || got[1].Name != "inside" {
		t.Fatalf("scoped transition groups = %+v", got)
	}
	frame := local.At(.5)
	display := frame.Display()
	if len(display.Ops) <= len(frame.Base.Ops) || len(display.Order) != len(display.Ops) {
		t.Fatalf("scoped frame display ops/order = %d/%d, base ops %d", len(display.Ops), len(display.Order), len(frame.Base.Ops))
	}
}

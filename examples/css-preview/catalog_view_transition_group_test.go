package main

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/viewtransition"
)

func TestCatalogViewTransitionGroupBuildsOwnFrameHierarchy(t *testing.T) {
	groupMode := "nearest"
	d := catalogViewDisplay(t, `<style>.p{display:block;width:80px;height:60px;background:red;view-transition-name:parent;view-transition-group:contain}.c{display:block;width:30px;height:20px;background:blue;view-transition-name:child;view-transition-group:`+groupMode+`}</style><main id="parent" class="p"><b id="child" class="c">item</b></main>`)
	tr, err := viewtransition.New(d, d, viewtransition.Scope{OldRoot: -1, NewRoot: -1}, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, group := range tr.Groups() {
		if group.Name == "child" {
			if group.Parent != "parent" || group.Mode != groupMode {
				t.Fatalf("child group = %+v", group)
			}
			frame := tr.At(.5)
			display := frame.Display()
			if len(display.Ops) <= len(frame.Base.Ops) || len(display.Order) != len(display.Ops) || len(display.OpOwners) != len(display.Ops) {
				t.Fatalf("composed group display = ops/order/owners %d/%d/%d, base ops %d", len(display.Ops), len(display.Order), len(display.OpOwners), len(frame.Base.Ops))
			}
			return
		}
	}
	t.Fatal("child view-transition group missing")
}

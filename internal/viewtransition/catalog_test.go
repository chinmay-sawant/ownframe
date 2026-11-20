package viewtransition

import (
	"testing"
)

func TestCatalogViewTransitionClassChangesOwnFrameFrame(t *testing.T) {
	d := snapshot(t, `<style>.c{display:block;width:40px;height:20px;background:red;view-transition-name:card;view-transition-class:promo}</style><div id="target" class="c">card</div>`)
	scope := Scope{OldRoot: -1, NewRoot: -1}
	plain, err := New(d, d, scope, nil)
	if err != nil {
		t.Fatal(err)
	}
	alpha, scale := .5, 1.5
	styled, err := New(d, d, scope, map[string]ClassStyle{"promo": {Opacity: &alpha, Scale: &scale}})
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

func TestCatalogViewTransitionGroupBuildsOwnFrameHierarchy(t *testing.T) {
	groupMode := "nearest"
	d := snapshot(t, `<style>.p{display:block;width:80px;height:60px;background:red;view-transition-name:parent;view-transition-group:contain}.c{display:block;width:30px;height:20px;background:blue;view-transition-name:child;view-transition-group:`+groupMode+`}</style><main id="parent" class="p"><b id="child" class="c">item</b></main>`)
	tr, err := New(d, d, Scope{OldRoot: -1, NewRoot: -1}, nil)
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

func TestCatalogViewTransitionNameReachesOwnFrameFrames(t *testing.T) {
	old := snapshot(t, `<style>.c{display:block;width:40px;height:20px;background:red;view-transition-name:card}</style><div id="target" class="c">old</div>`)
	next := snapshot(t, `<style>.c{display:block;width:100px;height:20px;background:blue;view-transition-name:card}</style><div id="target" class="c">new</div>`)
	tr, err := New(old, next, Scope{OldRoot: -1, NewRoot: -1}, nil)
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

func TestCatalogViewTransitionScopeLimitsOwnFrameFrames(t *testing.T) {
	scopeValue := "all"
	source := `<style>.outer{view-transition-name:outer}.scope{view-transition-scope:` + scopeValue + `;view-transition-name:scope}.inner{display:block;view-transition-name:inside}.away{view-transition-name:away}</style><main class="outer"><section id="scope" class="scope"><div class="inner">inside</div></section><p class="away">away</p></main>`
	d := snapshot(t, source)
	doc, err := New(d, d, Scope{OldRoot: -1, NewRoot: -1}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := doc.Groups(); len(got) != 2 || got[0].Name != "outer" || got[1].Name != "away" {
		t.Fatalf("document transition groups = %+v", got)
	}
	root := boxNamed(d, "scope")
	if root < 0 {
		t.Fatal("scoped element missing from Blinkless display")
	}
	local, err := New(d, d, Scope{OldRoot: root, NewRoot: root}, nil)
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

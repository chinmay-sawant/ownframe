package page

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

// TestScopedPatchSkipsUnrelatedCapture pins the scope-gated transition
// capture: a patch outside every transitioning element must not run the
// capture walk. The id-less transitioning box would be counted by a
// whole-display capture, so the skipped commit must leave TransitionsNoID
// at 0. The layouts delta proves the patch still committed.
func TestScopedPatchSkipsUnrelatedCapture(t *testing.T) {
	source := `<div id="a" style="width:10px;height:10px">x</div><div style="width:10px;height:10px;transition: opacity 1s">v</div>`
	p, err := New(Config{HTML: source, Width: 200, Height: 100, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	before := p.Stats().Layouts
	text := "y"
	if err := p.PatchElement(ctx, "a", ElementEdit{Text: &text}); err != nil {
		t.Fatal(err)
	}
	if got := p.Stats().Layouts; got != before+1 {
		t.Fatalf("layouts = %d, want %d (patch did not commit)", got, before+1)
	}
	if got := p.Motion().TransitionsNoID; got != 0 {
		t.Fatalf("TransitionsNoID = %d, want 0 (capture ran outside the scope)", got)
	}
}

// TestScopedPatchBindsInScopeTransition is the control: patching the
// transitioning element itself must run the full capture, bind the run, and
// keep the id-less diagnostic. The scope gate must not over-skip.
func TestScopedPatchBindsInScopeTransition(t *testing.T) {
	source := `<style>.on{opacity:0.5}</style><div id="t" class="off" style="width:10px;height:10px;transition: opacity 1s">v</div><div style="width:10px;height:10px;transition: opacity 1s">u</div>`
	p, err := New(Config{HTML: source, Width: 200, Height: 100, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := p.PatchElement(ctx, "t", ElementEdit{Attribute: "class", Value: "on"}); err != nil {
		t.Fatal(err)
	}
	if got := p.Motion().TransitionsBound; got != 1 {
		t.Fatalf("bound transitions = %d, want 1", got)
	}
	if got := p.Motion().TransitionsNoID; got != 1 {
		t.Fatalf("TransitionsNoID = %d, want 1 (diagnostic lost)", got)
	}
}

// TestScopeDeclaresTransitions pins the scope predicate: transitioning boxes
// match by id, by id-bearing ancestor, or not at all. Id-less boxes match
// only through an ancestor in scope, because the cascade anchors hover
// scopes at id-bearing boxes.
func TestScopeDeclaresTransitions(t *testing.T) {
	trans := []layout.Transition{{Property: "opacity"}}
	mk := func(id string, parent int, withTrans bool) Box {
		var tr []layout.Transition
		if withTrans {
			tr = trans
		}
		return Box{ID: id, ParentIndex: parent, Transitions: tr}
	}
	cases := []struct {
		name  string
		boxes []Box
		roots []string
		want  bool
	}{
		{"no transitions", []Box{mk("a", -1, false)}, []string{"a"}, false},
		{"root declares", []Box{mk("a", -1, true)}, []string{"a"}, true},
		{"outside scope", []Box{mk("a", -1, false), mk("b", -1, true)}, []string{"a"}, false},
		{"child of root", []Box{mk("a", -1, false), mk("b", 0, true)}, []string{"a"}, true},
		{"id-less under root", []Box{mk("a", -1, false), mk("", 0, true)}, []string{"a"}, true},
		{"id-less outside scope", []Box{mk("a", -1, false), mk("", -1, true)}, []string{"a"}, false},
		{"nil roots", []Box{mk("a", -1, true)}, nil, false},
		{"empty roots", []Box{mk("a", -1, true)}, []string{}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := &Display{Boxes: tc.boxes}
			if got := scopeDeclaresTransitions(d, tc.roots); got != tc.want {
				t.Fatalf("scopeDeclaresTransitions = %v, want %v", got, tc.want)
			}
		})
	}
}

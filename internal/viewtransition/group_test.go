package viewtransition

import (
	"errors"
	"testing"
)

func TestViewTransitionGroupBuildsNestedNamedHierarchy(t *testing.T) {
	markup := `<style>.a{view-transition-name:outer;view-transition-group:contain}.b{view-transition-name:middle;view-transition-group:normal}.c{view-transition-name:near;view-transition-group:nearest}.d{view-transition-name:custom;view-transition-group:outer}</style><main class="a"><section class="b"><b class="c">near</b></section><i class="d">custom</i></main>`
	d := snapshot(t, markup)
	tr, err := New(d, d, Scope{DocumentRoot, DocumentRoot}, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"outer": "", "middle": "outer", "near": "middle", "custom": "outer"}
	for _, g := range tr.Groups() {
		if p, ok := want[g.Name]; !ok || p != g.Parent {
			t.Errorf("group %q parent=%q, want %q", g.Name, g.Parent, p)
		}
	}
}

func TestViewTransitionDuplicateNameFailsDiscovery(t *testing.T) {
	d := snapshot(t, `<style>.n{view-transition-name:repeat}</style><p class="n">one</p><p class="n">two</p>`)
	if _, err := New(d, d, Scope{DocumentRoot, DocumentRoot}, nil); !errors.Is(err, ErrDuplicateName) {
		t.Fatalf("duplicate name error=%v, want %v", err, ErrDuplicateName)
	}
}

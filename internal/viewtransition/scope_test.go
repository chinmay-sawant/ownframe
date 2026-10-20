package viewtransition

import "testing"

func TestViewTransitionScopeLimitsDiscoveryToItsSubtree(t *testing.T) {
	source := `<style>.outer{view-transition-name:outer}.scope{view-transition-scope:all;view-transition-name:scope}.inner{display:block;view-transition-name:inside}.away{view-transition-name:away}</style><main class="outer"><section id="scope" class="scope"><div class="inner">inside</div></section><p class="away">away</p></main>`
	d := snapshot(t, source)
	doc, err := New(d, d, Scope{DocumentRoot, DocumentRoot}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if names := doc.Groups(); len(names) != 2 || names[0].Name != "outer" || names[1].Name != "away" {
		t.Fatalf("document scope names=%v", names)
	}
	i := boxNamed(d, "scope")
	if i < 0 {
		t.Fatal("scope element missing from Blinkless display")
	}
	scoped, err := New(d, d, Scope{i, i}, nil)
	if err != nil {
		t.Fatal(err)
	}
	got := scoped.Groups()
	if len(got) != 2 || got[0].Name != "scope" || got[1].Name != "inside" {
		t.Fatalf("element scope names=%v boxes=%+v", got, d.Boxes)
	}
}

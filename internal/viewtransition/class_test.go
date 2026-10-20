package viewtransition

import "testing"

func TestViewTransitionClassChangesGeneratedLayerStyle(t *testing.T) {
	old, next := pairedDisplays(t)
	plain, err := New(old, next, Scope{DocumentRoot, DocumentRoot}, nil)
	if err != nil {
		t.Fatal(err)
	}
	alpha, scale := 0.5, 1.25
	styled, err := New(old, next, Scope{DocumentRoot, DocumentRoot}, map[string]ClassStyle{"card": {Opacity: &alpha, Scale: &scale}})
	if err != nil {
		t.Fatal(err)
	}
	a, b := plain.At(0.5).Groups[1], styled.At(0.5).Groups[1]
	if len(b.Classes) != 2 || b.Classes[0] != "card" || b.Classes[1] != "promo" {
		t.Fatalf("classes=%v", b.Classes)
	}
	if len(a.New) == 0 || len(b.New) == 0 {
		t.Fatal("named child has no captured paint")
	}
	if b.New[0].Opacity() >= a.New[0].Opacity() || b.New[0].W == a.New[0].W {
		t.Fatal("class style did not change generated paint")
	}
}

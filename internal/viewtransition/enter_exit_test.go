package viewtransition

import "testing"

func TestViewTransitionFadesNamesEnteringAndLeaving(t *testing.T) {
	old := snapshot(t, `<style>.old{display:block;view-transition-name:leaving}</style><div class="old">old</div>`)
	next := snapshot(t, `<style>.new{display:block;view-transition-name:entering}</style><div class="new">new</div>`)
	tr, err := New(old, next, Scope{DocumentRoot, DocumentRoot}, nil)
	if err != nil {
		t.Fatal(err)
	}
	f := tr.At(0.25)
	if len(f.Groups) != 2 || f.Groups[0].Name != "entering" || len(f.Groups[0].Old) != 0 || len(f.Groups[0].New) == 0 {
		t.Fatalf("entering layer=%+v", f.Groups)
	}
	if len(f.Groups[1].New) != 0 || len(f.Groups[1].Old) == 0 {
		t.Fatalf("leaving layer=%+v", f.Groups[1])
	}
	if f.Groups[0].NewOpacity != 0.25 || f.Groups[1].OldOpacity != 0.75 {
		t.Fatalf("transition alphas=%v/%v", f.Groups[0].NewOpacity, f.Groups[1].OldOpacity)
	}
}

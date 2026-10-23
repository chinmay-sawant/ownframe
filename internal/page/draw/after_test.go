package draw

import "testing"

func TestRedrawAfterDefersOneDraw(t *testing.T) {
	t.Parallel()

	var a After
	if a.Take() {
		t.Fatal("idle")
	}

	ran := false
	a.Set(func() {
		ran = true
		if !a.Busy() {
			t.Fatal("not busy")
		}
		a.Ask()
		a.Run()
	})
	a.Run()
	if !ran || !a.Take() || a.Take() {
		t.Fatal("ask")
	}

	a.Done()
	a.Ask()
	if !a.Take() {
		t.Fatal("again")
	}
	a.Set(nil)
	a.Run()
}

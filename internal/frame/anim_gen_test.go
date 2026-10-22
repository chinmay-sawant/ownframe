package frame

import (
	"context"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

// genHTML is one card; data moves its edge so a Redraw replaces the list.
const genHTML = `<style>#card{position:absolute;left:{{.X}}px;top:40px;` +
	`width:120px;height:80px;background:#3366cc}</style><div id="card"></div>`

// genCard builds the card page and returns the page, list, and card fill.
func genCard(t *testing.T) (*page.Page, *page.Display, *page.DisplayOp) {
	t.Helper()
	s, err := page.New(page.Config{HTML: genHTML, Width: 640, Height: 480})
	if err != nil {
		t.Fatal(err)
	}
	s.SetData(struct{ X int }{X: 40})
	if err := s.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	d := s.Display()
	if d == nil {
		t.Fatal("no display list")
	}
	box := page.Box{X: 20, Y: 20, W: 300, H: 200}
	if f := Fill(d, box, [3]float64{0.2, 0.4, 0.8}); f != nil {
		return s, d, f
	}
	t.Fatal("card op not found")
	return nil, nil, nil
}

// A bound animation tracks the live list, then dies on Redraw: Step goes
// inactive with no damage and leaves the freed operations untouched.
func TestBoundAnimDiesOnRedraw(t *testing.T) {
	t.Parallel()
	s, d, fill := genCard(t)
	a := TranslateTo(d, []*page.DisplayOp{fill}, 30, 0, time.Second, Linear).Bind(s)
	if a.Generation() != s.Generation() {
		t.Fatal("generation not recorded")
	}
	if _, live := a.Step(t0); !live {
		t.Fatal("bound animation dead before redraw")
	}
	x0 := fill.X
	s.SetData(struct{ X int }{X: 90})
	if err := s.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	if s.Display() == d {
		t.Fatal("redraw kept the display list")
	}
	dmg, live := a.Step(t0.Add(time.Second))
	if live || a.Active() || !dmg.Empty() {
		t.Fatal("bound animation alive after redraw")
	}
	if fill.X != x0 {
		t.Fatalf("freed op patched to %v", fill.X)
	}
}

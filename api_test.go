package ownframe_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestAPINewSetDataHandleRedraw(t *testing.T) {
	p, err := ownframe.New(ownframe.Config{
		HTML:   `<input id="e" value="x"><p id="n">{{.N}}</p>`,
		Width:  320,
		Height: 200,
	})
	if err != nil {
		t.Fatal(err)
	}

	p.SetData(struct{ N int }{N: 1})
	p.Handle(ownframe.Handlers{
		Click: func(context.Context, ownframe.Box) error { return nil },
	})
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.SetFormValue("e", "hi")
	if got := p.FormValue("e"); got != "hi" {
		t.Fatalf("FormValue = %q", got)
	}
	if p.Testing().Generation() == 0 {
		t.Fatal("generation is 0")
	}
}

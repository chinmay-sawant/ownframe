package ownframe_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestTestingTickTakeDirty(t *testing.T) {
	p, err := ownframe.New(ownframe.Config{
		HTML:   `<p id="n">hi</p>`,
		Width:  320,
		Height: 200,
	})
	if err != nil {
		t.Fatal(err)
	}
	p.SetTick(func(context.Context) error { return nil })
	if err := p.Testing().Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.Probe().TakeDirty()
}

func TestTestingClickAfterRedraw(t *testing.T) {
	p, err := ownframe.New(ownframe.Config{
		HTML:   `<button id="go">n</button>`,
		Width:  320,
		Height: 200,
	})
	if err != nil {
		t.Fatal(err)
	}
	var hit string
	p.Handle(ownframe.Handlers{
		Click: func(_ context.Context, b ownframe.Box) error {
			hit = b.ID
			return nil
		},
	})
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	var x, y float64
	for _, b := range p.Boxes() {
		if b.ID == "go" {
			x, y = b.X+b.W/2, b.Y+b.H/2
		}
	}
	if err := p.Click(context.Background(), x, y); err != nil {
		t.Fatal(err)
	}
	if hit != "go" {
		t.Fatalf("hit = %q", hit)
	}
}

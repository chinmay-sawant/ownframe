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
	p.Testing().TakeDirty()
}

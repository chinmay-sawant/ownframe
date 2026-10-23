package page

import (
	"context"
	"testing"
)

func TestTickNoOpStaysClean(t *testing.T) {
	t.Parallel()

	p := tickBarPage(t)
	p.SetTick(func(context.Context) error { return nil })
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if p.FrameDirty() {
		t.Fatal("no-op tick set FrameDirty")
	}
	if _, ok := p.TakeDirty(); ok {
		t.Fatal("no-op tick marked dirty")
	}
}

func TestUseFrameDirtyStillSetsFlag(t *testing.T) {
	t.Parallel()

	p := tickBarPage(t)
	p.SetTick(func(context.Context) error {
		p.UseFrameDirty()

		return nil
	})
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !p.FrameDirty() {
		t.Fatal("UseFrameDirty alias did not set the flag")
	}
	if _, ok := p.TakeDirty(); ok {
		t.Fatal("alias-only tick marked a rect")
	}
}

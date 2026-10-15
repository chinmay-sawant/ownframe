package page

import (
	"context"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestCaretBlinkHidesAndShows(t *testing.T) {
	ctx := context.Background()
	p, err := New(Config{HTML: `<input id="e" type="text" value="ab">`, Width: 300, Height: 100})
	if err != nil {
		t.Fatal(err)
	}

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if err := p.Focus(ctx, "e"); err != nil {
		t.Fatal(err)
	}

	if got := p.display.Ops[len(p.display.Ops)-1].Kind; got != layout.DisplayOpFillRect {
		t.Fatalf("visible caret kind = %v, want fill rect", got)
	}

	p.formState().blinkAt = time.Now().Add(-caretBlinkInterval)
	if err := p.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	if got := p.display.Ops[len(p.display.Ops)-1].Kind; got != layout.DisplayOpNoop {
		t.Fatalf("hidden caret kind = %v, want noop", got)
	}

	if err := p.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	if got := p.display.Ops[len(p.display.Ops)-1].Kind; got != layout.DisplayOpNoop {
		t.Fatalf("caret returned early with kind %v", got)
	}

	p.formState().blinkAt = time.Now().Add(-caretBlinkInterval)
	if err := p.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	if got := p.display.Ops[len(p.display.Ops)-1].Kind; got != layout.DisplayOpFillRect {
		t.Fatalf("caret did not return, kind = %v", got)
	}
}

func TestEditResetsBlink(t *testing.T) {
	ctx := context.Background()
	p, err := New(Config{HTML: `<input id="e" type="text" value="ab">`, Width: 300, Height: 100})
	if err != nil {
		t.Fatal(err)
	}

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if err := p.Focus(ctx, "e"); err != nil {
		t.Fatal(err)
	}

	p.formState().blinkAt = time.Now().Add(-caretBlinkInterval)
	if err := p.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	if p.formState().blinkOn {
		t.Fatal("blink did not turn off")
	}

	if err := p.Type(ctx, "c"); err != nil {
		t.Fatal(err)
	}

	if !p.formState().blinkOn {
		t.Fatal("typing did not reset the blink")
	}
}

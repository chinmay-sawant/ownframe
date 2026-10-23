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

	if got := p.draw.Display.Ops[len(p.draw.Display.Ops)-1].Kind; got != layout.DisplayOpFillRect {
		t.Fatalf("visible caret kind = %v, want fill rect", got)
	}

	st := p.formState().snap()
	st.BlinkAt = time.Now().Add(-caretBlinkInterval)
	p.formState().store(&st)
	if err := p.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	if got := p.draw.Display.Ops[len(p.draw.Display.Ops)-1].Kind; got != layout.DisplayOpNoop {
		t.Fatalf("hidden caret kind = %v, want noop", got)
	}

	if err := p.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	if got := p.draw.Display.Ops[len(p.draw.Display.Ops)-1].Kind; got != layout.DisplayOpNoop {
		t.Fatalf("caret returned early with kind %v", got)
	}

	st = p.formState().snap()
	st.BlinkAt = time.Now().Add(-caretBlinkInterval)
	p.formState().store(&st)
	if err := p.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	if got := p.draw.Display.Ops[len(p.draw.Display.Ops)-1].Kind; got != layout.DisplayOpFillRect {
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

	st := p.formState().snap()
	st.BlinkAt = time.Now().Add(-caretBlinkInterval)
	p.formState().store(&st)
	if err := p.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	if p.formState().snap().BlinkOn {
		t.Fatal("blink did not turn off")
	}

	if err := p.Type(ctx, "c"); err != nil {
		t.Fatal(err)
	}

	if !p.formState().snap().BlinkOn {
		t.Fatal("typing did not reset the blink")
	}
}

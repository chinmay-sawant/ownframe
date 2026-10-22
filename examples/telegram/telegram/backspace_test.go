package telegram_test

import (
	"context"
	"testing"
)

// TestComposerBackspace drives the page half of a soft-keyboard Backspace.
// The window uses the plain path for a key and IMEReplace for the IME's
// surrounding-text commit. The value, the caret, and the pinned bars must
// survive both.
func TestComposerBackspace(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)
	click(t, ctx, app, "", "open-anna")
	click(t, ctx, app, "compose", "")

	app.SetInsets(28, 32)
	if err := app.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	app.Page().SetScrollOffset(0, 120)
	if err := app.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if err := app.Type(ctx, "hello world"); err != nil {
		t.Fatal(err)
	}

	if err := app.Page().Backspace(ctx); err != nil {
		t.Fatal(err)
	}

	if got := app.Page().FormValue("compose"); got != "hello worl" {
		t.Fatalf("value = %q", got)
	}

	_, caret, before, after, ok := app.Page().IMEContext()
	if !ok || caret != 10 || before != "hello worl" || after != "" {
		t.Fatalf("guard ok=%v caret=%d before=%q after=%q", ok, caret, before, after)
	}

	// The IME commit for a backspace after "hello": the window hands the
	// page prefix+text+suffix, and the caret belongs after the prefix, not
	// at the end of the field.
	if err := app.Page().Testing().IMEReplace(ctx, 0, 10, "hell worl", 4); err != nil {
		t.Fatal(err)
	}

	if got := app.Page().FormValue("compose"); got != "hell worl" {
		t.Fatalf("ime value = %q", got)
	}

	_, caret, before, after, ok = app.Page().IMEContext()
	if !ok || caret != 4 || before != "hell" || after != " worl" {
		t.Fatalf("ime ok=%v caret=%d before=%q after=%q", ok, caret, before, after)
	}

	_, viewH := app.Page().Size()
	view := app.View()
	if view.BarTop != 120 || view.BottomTop != 120+viewH-view.InsetBottom-63-view.InsetTop {
		t.Fatalf("pins = %d, %d", view.BarTop, view.BottomTop)
	}
}

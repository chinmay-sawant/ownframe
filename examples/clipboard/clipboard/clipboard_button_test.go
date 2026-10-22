package clipboard_test

import (
	"context"
	"strings"
	"testing"
)

func TestButtonsWriteStatusAndText(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)

	click(t, ctx, app, "copy")
	if got := app.View().Status; !strings.Contains(got, "nothing selected") {
		t.Fatalf("empty copy status = %q", got)
	}

	click(t, ctx, app, "left")
	if err := app.Type(ctx, "hello"); err != nil {
		t.Fatal(err)
	}

	click(t, ctx, app, "copy")
	if got := app.View().Status; !strings.Contains(got, "copied: hello") {
		t.Fatalf("copy status = %q", got)
	}

	click(t, ctx, app, "cut")
	if got := app.FormValue("left"); got != "" {
		t.Fatalf("after cut = %q, want empty", got)
	}

	if got := app.View().Status; !strings.Contains(got, "cut: hello") {
		t.Fatalf("cut status = %q", got)
	}

	click(t, ctx, app, "pastego")
	if got := app.FormValue("left"); !strings.Contains(got, "from Go") {
		t.Fatalf("after pastego = %q, want from Go", got)
	}

	click(t, ctx, app, "selectall")
	if !app.Page().Form().FormSelected("left") {
		t.Fatal("selectall did not select the left field")
	}
}

// TestSelectAllSurvivesCaretPlacement mirrors the window press: the caret
// lands from the press point before the click handler runs.
func TestSelectAllSurvivesCaretPlacement(t *testing.T) {
	ctx := context.Background()
	app := newApp(t, ctx)

	click(t, ctx, app, "left")
	if err := app.Type(ctx, "hello"); err != nil {
		t.Fatal(err)
	}

	box := boxByID(t, app, "selectall")
	if err := app.Page().Testing().SelectAt(ctx, box.X+box.W/2, box.Y+box.H/2); err != nil {
		t.Fatal(err)
	}

	click(t, ctx, app, "selectall")
	if !app.Page().Form().FormSelected("left") {
		t.Fatal("selectall lost the selection after the caret placement")
	}
}

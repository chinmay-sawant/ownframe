package clipboard_test

import (
	"context"
	"strings"
	"testing"
)

// A button click blurs the form first, so a button must act on the field
// the user clicked last, not always on the left field.
func TestButtonsUseLastFocusedField(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)

	click(t, ctx, app, "right")
	if err := app.Type(ctx, "rr"); err != nil {
		t.Fatal(err)
	}

	click(t, ctx, app, "copy")
	if got := app.View().Status; !strings.Contains(got, "copied: rr") {
		t.Fatalf("copy status = %q, want copied: rr", got)
	}

	click(t, ctx, app, "selectall")
	if !app.Page().Form().FormSelected("right") {
		t.Fatal("selectall did not select the right field")
	}

	click(t, ctx, app, "pastego")
	if got := app.FormValue("right"); !strings.Contains(got, "from Go") {
		t.Fatalf("right = %q, want from Go", got)
	}

	click(t, ctx, app, "cut")
	if got := app.FormValue("right"); got != "" {
		t.Fatalf("right after cut = %q, want empty", got)
	}

	if got := app.View().Status; !strings.Contains(got, "cut: ") {
		t.Fatalf("cut status = %q, want cut: ...", got)
	}

	if got := app.FormValue("left"); got != "" {
		t.Fatalf("left = %q, want untouched", got)
	}
}

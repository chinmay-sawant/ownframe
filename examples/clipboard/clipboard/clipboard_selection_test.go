package clipboard_test

import (
	"context"
	"testing"
)

func TestPasteGoReplacesSelectedTextAfterButtonBlur(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)
	click(t, ctx, app, "left")
	if err := app.Type(ctx, "replace this whole value"); err != nil {
		t.Fatal(err)
	}
	click(t, ctx, app, "selectall")
	click(t, ctx, app, "pastego")

	if got := app.FormValue("left"); got != "from Go" {
		t.Fatalf("after paste = %q, want from Go", got)
	}
}

package clipboard_test

import (
	"context"
	"strings"
	"testing"

	osclip "github.com/chinmay-sawant/ownframe/internal/clipboard"
)

func TestCopyAndCutWriteClipboard(t *testing.T) {
	ctx := context.Background()
	app := newApp(t, ctx)

	click(t, ctx, app, "left")
	if err := app.Type(ctx, "hello"); err != nil {
		t.Fatal(err)
	}

	click(t, ctx, app, "copy")
	got, err := osclip.Read()
	if err != nil {
		t.Fatal(err)
	}

	if got != "hello" {
		t.Fatalf("clipboard after copy = %q, want hello", got)
	}

	click(t, ctx, app, "cut")
	got, err = osclip.Read()
	if err != nil {
		t.Fatal(err)
	}

	if got != "hello" {
		t.Fatalf("clipboard after cut = %q, want hello", got)
	}
}

func TestPastePullsClipboardText(t *testing.T) {
	ctx := context.Background()
	app := newApp(t, ctx)

	osclip.Write("first\nsecond")

	click(t, ctx, app, "left")
	click(t, ctx, app, "paste")

	if got := app.FormValue("left"); got != "firstsecond" {
		t.Fatalf("pasted = %q, want firstsecond", got)
	}

	if got := app.View().Status; !strings.Contains(got, "pasted: firstsecond") {
		t.Fatalf("status = %q", got)
	}
}

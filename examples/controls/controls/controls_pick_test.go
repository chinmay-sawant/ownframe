package controls_test

import (
	"context"
	"strings"
	"testing"
)

func TestDocPickerUpdatesStatus(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)

	app.Page().SetPicker(func(context.Context, string) (string, error) {
		return "/tmp/notes.txt", nil
	})

	click(t, ctx, app, "doc")

	if got := app.FormValue("doc"); got != "/tmp/notes.txt" {
		t.Fatalf("doc = %q", got)
	}

	if got := boxText(t, app, "doc"); got != "/tmp/notes.txt" {
		t.Fatalf("doc text = %q", got)
	}

	if got := boxText(t, app, "status"); !strings.Contains(got, "doc=/tmp/notes.txt") {
		t.Fatalf("status = %q", got)
	}
}

package page_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/filepick"
	"github.com/chinmay-sawant/ownframe/internal/page"
)

func TestFileControlCancelKeepsValue(t *testing.T) {
	t.Parallel()

	screen := formPage(t, `<input id="f" type="file" value="old.txt"`+wide+`>`)

	calls := 0
	page.InstallPicker(screen, func(context.Context, string) (string, error) {
		calls++

		return "", filepick.ErrCanceled
	})

	clickID(t, screen, "f")
	if calls != 1 {
		t.Fatalf("calls %d", calls)
	}

	if screen.FormValue("f") != "old.txt" {
		t.Fatalf("value %q", screen.FormValue("f"))
	}
}

func TestFileControlDisabledSkipsPicker(t *testing.T) {
	t.Parallel()

	screen := formPage(t, `<input id="f" type="file" disabled`+wide+`>`)

	calls := 0
	page.InstallPicker(screen, func(context.Context, string) (string, error) {
		calls++

		return "/tmp/a.txt", nil
	})

	clickID(t, screen, "f")
	if calls != 0 || screen.FormValue("f") != "" {
		t.Fatalf("calls %d value %q", calls, screen.FormValue("f"))
	}
}

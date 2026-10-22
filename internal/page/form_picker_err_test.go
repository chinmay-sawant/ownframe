package page_test

import (
	"context"
	"errors"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/filepick"
	"github.com/chinmay-sawant/ownframe/internal/page"
)

func clickFile(t *testing.T, screen *page.Page, id string) error {
	t.Helper()
	x, y := boxCenter(t, screen, id)

	return screen.Click(context.Background(), x, y)
}

func TestFileControlPickerErrorKeepsValue(t *testing.T) {
	t.Parallel()

	screen := formPage(t, `<input id="f" type="file" value="old.txt"`+wide+`>`)
	want := errors.New("picker failed")
	page.InstallPicker(screen, func(context.Context, string) (string, error) {
		return "", want
	})

	err := clickFile(t, screen, "f")
	if !errors.Is(err, want) {
		t.Fatalf("click %v", err)
	}
	if screen.FormValue("f") != "old.txt" {
		t.Fatalf("value %q", screen.FormValue("f"))
	}
}

func TestFileControlNoDialog(t *testing.T) {
	t.Parallel()

	screen := formPage(t, `<input id="f" type="file" value="old.txt"`+wide+`>`)
	page.InstallPicker(screen, func(context.Context, string) (string, error) {
		return "", filepick.ErrNoDialog
	})

	err := clickFile(t, screen, "f")
	if !errors.Is(err, filepick.ErrNoDialog) {
		t.Fatalf("click %v", err)
	}
	if screen.FormValue("f") != "old.txt" {
		t.Fatalf("value %q", screen.FormValue("f"))
	}
	if screen.FocusedField() != "f" {
		t.Fatalf("focus %q", screen.FocusedField())
	}
}

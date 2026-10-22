package page_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

func TestFileControlUsesPicker(t *testing.T) {
	t.Parallel()

	screen := formPage(t, `<input id="f" type="file"`+wide+`>`)

	var title string
	page.InstallPicker(screen, func(_ context.Context, got string) (string, error) {
		title = got

		return "/tmp/report.pdf", nil
	})

	clickID(t, screen, "f")
	if screen.FormValue("f") != "/tmp/report.pdf" {
		t.Fatalf("value %q", screen.FormValue("f"))
	}

	if boxText(t, screen, "f") != "/tmp/report.pdf" {
		t.Fatalf("text %q", boxText(t, screen, "f"))
	}

	if title != "ownframe" {
		t.Fatalf("title %q", title)
	}
}

func TestFileControlChangeAndBind(t *testing.T) {
	t.Parallel()

	view := &struct{ Doc string }{}
	screen := formPage(t, `<input id="f" type="file" data-bind="Doc"`+wide+`>`)
	screen.SetData(view)

	var events []string
	screen.Handle(page.Handlers{
		BeforeEdit: func(_ context.Context, b page.Box) error {
			events = append(events, "before:"+b.ID)

			return nil
		},
		Change: func(_ context.Context, b page.Box) error {
			events = append(events, "change:"+b.ID)

			return nil
		},
	})
	page.InstallPicker(screen, func(context.Context, string) (string, error) {
		return "/tmp/x.txt", nil
	})

	clickID(t, screen, "f")

	if view.Doc != "/tmp/x.txt" {
		t.Fatalf("bound %q", view.Doc)
	}

	if len(events) != 2 || events[0] != "before:f" || events[1] != "change:f" {
		t.Fatalf("events %v", events)
	}
}

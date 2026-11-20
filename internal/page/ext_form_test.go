package page_test

import (
	"context"
	"errors"
	"github.com/chinmay-sawant/ownframe/internal/filepick"
	"github.com/chinmay-sawant/ownframe/internal/page"
	"strings"
	"testing"
)

func TestFormBlurSkipsType(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	body := `<input id="t" type="text"` + wide + `><div id="plain">x</div>`
	screen := formPage(t, body)
	clickID(t, screen, "t")
	if err := screen.Type(ctx, "ab"); err != nil {
		t.Fatal(err)
	}

	clickID(t, screen, "plain")
	if err := screen.Type(ctx, "z"); err != nil {
		t.Fatal(err)
	}

	if screen.FormValue("t") != "ab" {
		t.Fatalf("value %q", screen.FormValue("t"))
	}
}

func TestFormLoadClearsValue(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen := formPage(t, `<input id="t" type="text"`+wide+`>`)
	clickID(t, screen, "t")
	if err := screen.Type(ctx, "ab"); err != nil {
		t.Fatal(err)
	}

	if err := screen.Load(ctx, `<div id="z">Z</div>`); err != nil {
		t.Fatal(err)
	}

	if screen.FormValue("t") != "" {
		t.Fatalf("value %q", screen.FormValue("t"))
	}
}

func TestFormCheckbox(t *testing.T) {
	t.Parallel()

	screen := formPage(t, `<input id="c" type="checkbox">`)
	if screen.FormChecked("c") {
		t.Fatal("starts checked")
	}

	clickID(t, screen, "c")
	if !screen.FormChecked("c") {
		t.Fatal("not checked")
	}
}

func TestFormRadios(t *testing.T) {
	t.Parallel()

	body := `<input id="a" type="radio" name="g" checked style="display:block">` +
		`<input id="b" type="radio" name="g" style="display:block">`
	screen := formPage(t, body)
	if !screen.FormChecked("a") || screen.FormChecked("b") {
		t.Fatal("initial radios")
	}

	clickID(t, screen, "b")
	if screen.FormChecked("a") || !screen.FormChecked("b") {
		t.Fatal("radio group")
	}
}

func TestDisabledControlsSkipHandler(t *testing.T) {
	t.Parallel()
	check := func(t *testing.T, body, id string, val func(*page.Page) string) {
		t.Helper()
		s := formPage(t, body)
		calls := 0
		s.Handle(page.Handlers{Click: func(context.Context, page.Box) error { calls++; return nil }})
		page.InstallPicker(s, func(context.Context, string) (string, error) { calls++; return "/tmp/x.txt", nil })
		before := val(s)
		gen := s.Generation()
		clickID(t, s, id)
		if calls != 0 {
			t.Fatalf("%s calls = %d, want 0", id, calls)
		}
		if got := val(s); got != before {
			t.Fatalf("%s value = %q, want %q", id, got, before)
		}
		if s.Generation() != gen {
			t.Fatalf("%s gen moved", id)
		}
	}
	t.Run("checkbox", func(t *testing.T) {
		t.Parallel()
		check(t, `<input id="c" type="checkbox" disabled>`, "c", func(s *page.Page) string {
			if s.FormChecked("c") {
				return "1"
			}
			return "0"
		})
	})
	t.Run("text", func(t *testing.T) {
		t.Parallel()
		check(t, `<input id="t" type="text" value="keep" disabled`+wide+`>`, "t", func(s *page.Page) string { return s.FormValue("t") })
	})
	t.Run("file", func(t *testing.T) {
		t.Parallel()
		check(t, `<input id="f" type="file" disabled`+wide+`>`, "f", func(s *page.Page) string { return s.FormValue("f") })
	})
}

func TestFormSelectCycles(t *testing.T) {
	t.Parallel()

	body := `<select id="s"` + wide + `><option>Red</option>` +
		`<option selected>Blue</option></select>`
	screen := formPage(t, body)
	clickID(t, screen, "s")
	if strings.TrimSpace(boxText(t, screen, "s")) != "Red" {
		t.Fatalf("text %q", boxText(t, screen, "s"))
	}
}

func TestFormFile(t *testing.T) {
	t.Parallel()

	screen := formPage(t, `<input id="f" type="file"`+wide+`>`)
	clickID(t, screen, "f")
	if err := screen.Type(context.Background(), "notes.txt"); err != nil {
		t.Fatal(err)
	}

	if boxText(t, screen, "f") != "notes.txt" {
		t.Fatalf("text %q", boxText(t, screen, "f"))
	}
}

func TestFormTextareaBackspace(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen := formPage(t, `<textarea id="a"`+wide+`></textarea>`)
	clickID(t, screen, "a")
	if err := screen.Type(ctx, "hi"); err != nil {
		t.Fatal(err)
	}

	if err := screen.Backspace(ctx); err != nil {
		t.Fatal(err)
	}

	if screen.FormValue("a") != "h" {
		t.Fatalf("value %q", screen.FormValue("a"))
	}
}

func TestFormDisabledCheckbox(t *testing.T) {
	t.Parallel()

	screen := formPage(t, `<input id="c" type="checkbox" disabled>`)
	clickID(t, screen, "c")
	if screen.FormChecked("c") {
		t.Fatal("disabled checked")
	}
}

func formPage(t *testing.T, body string) *page.Page {
	t.Helper()

	screen, err := page.New(page.Config{
		HTML:   `<body style="margin:8px;font-family:sans-serif">` + body + `</body>`,
		Width:  640,
		Height: 480,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err = screen.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	return screen
}

func clickID(t *testing.T, screen *page.Page, id string) {
	t.Helper()

	x, y := boxCenter(t, screen, id)
	if err := screen.Click(context.Background(), x, y); err != nil {
		t.Fatal(err)
	}
}

const wide = ` style="display:block;width:220px;height:28px"`

func TestFormTypeText(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen := formPage(t, `<input id="t" type="text"`+wide+`>`)
	clickID(t, screen, "t")
	if err := screen.Type(ctx, "ab"); err != nil {
		t.Fatal(err)
	}

	if screen.FormValue("t") != "ab" || boxText(t, screen, "t") != "ab" {
		t.Fatalf("value %q text %q", screen.FormValue("t"), boxText(t, screen, "t"))
	}
}

func TestFormPasswordCopy(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen := formPage(t, `<input id="pw" type="password"`+wide+`>`)
	clickID(t, screen, "pw")
	if err := screen.Type(ctx, "secret"); err != nil {
		t.Fatal(err)
	}

	if screen.FormValue("pw") != "secret" {
		t.Fatalf("value %q", screen.FormValue("pw"))
	}

	got := boxText(t, screen, "pw")
	if strings.Contains(got, "secret") || !maskText(got) {
		t.Fatalf("box %q", got)
	}

	text, ok, err := screen.Copy(ctx)
	if err != nil || !ok || text != "secret" {
		t.Fatalf("copy %q %v %v", text, ok, err)
	}
}

func maskText(s string) bool {
	if s == "" {
		return false
	}

	for _, r := range s {
		switch r {
		case '•', '*', '●', '·', '∗', '⬤':
		default:
			return false
		}
	}

	return true
}

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

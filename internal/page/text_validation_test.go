package page_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

func TestTextValidationLabelFocus(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen := formPage(t, `<label for="email" style="display:block;width:140px;height:24px">Email<span style="display:inline-block;width:16px;height:16px">*</span></label>`+
		`<input id="email" type="text"`+wide+`>`+
		`<input id="other" type="text"`+wide+`>`)
	clickID(t, screen, "other")
	if screen.FocusID() != "other" {
		t.Fatalf("setup focus %q", screen.FocusID())
	}

	clickBox(t, screen, func(box page.Box) bool {
		return box.Tag == "span" && box.Text == "*"
	})
	if screen.FocusID() != "email" {
		t.Fatalf("label click focus %q", screen.FocusID())
	}

	clickID(t, screen, "other")
	clickBox(t, screen, func(box page.Box) bool {
		return box.Tag == "label"
	})
	if screen.FocusID() != "email" {
		t.Fatalf("label text focus %q", screen.FocusID())
	}

	if err := screen.Type(ctx, "a"); err != nil {
		t.Fatal(err)
	}
	if screen.FormValue("email") != "a" {
		t.Fatalf("typed %q", screen.FormValue("email"))
	}
}

func TestTextValidationOTPSlots(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	screen := formPage(t, `<input id="code-0" type="text" inputmode="numeric" maxlength="1"`+wide+`>`+
		`<input id="code-1" type="text" inputmode="numeric" maxlength="1"`+wide+`>`+
		`<input id="code-2" type="text" inputmode="numeric" maxlength="1"`+wide+`>`)
	clickID(t, screen, "code-0")
	if err := screen.Type(ctx, "a"); err != nil {
		t.Fatal(err)
	}
	if screen.FormValue("code-0") != "" || screen.FocusID() != "code-0" {
		t.Fatalf("non-digit value %q focus %q", screen.FormValue("code-0"), screen.FocusID())
	}

	if err := screen.Type(ctx, "4"); err != nil {
		t.Fatal(err)
	}
	if screen.FormValue("code-0") != "4" || screen.FocusID() != "code-1" {
		t.Fatalf("digit value %q focus %q", screen.FormValue("code-0"), screen.FocusID())
	}

	if err := screen.Backspace(ctx); err != nil {
		t.Fatal(err)
	}
	if screen.FocusID() != "code-0" || screen.FormValue("code-0") != "4" || screen.FormValue("code-1") != "" {
		t.Fatalf("backspace focus %q code-0 %q code-1 %q", screen.FocusID(), screen.FormValue("code-0"), screen.FormValue("code-1"))
	}

	clickID(t, screen, "code-2")
	if err := screen.Type(ctx, "9"); err != nil {
		t.Fatal(err)
	}
	if screen.FormValue("code-2") != "9" || screen.FocusID() != "code-2" {
		t.Fatalf("last slot value %q focus %q", screen.FormValue("code-2"), screen.FocusID())
	}
}

func clickBox(t *testing.T, screen *page.Page, match func(page.Box) bool) {
	t.Helper()

	for _, box := range screen.Boxes() {
		if box.W <= 0 || box.H <= 0 || !match(box) {
			continue
		}

		if err := screen.Click(context.Background(), box.X+box.W/2, box.Y+box.H/2); err != nil {
			t.Fatal(err)
		}

		return
	}

	t.Fatal("no matching box")
}

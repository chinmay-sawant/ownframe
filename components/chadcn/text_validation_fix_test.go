package chadcn_test

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

func TestTextValidationFieldMarksControl(t *testing.T) {
	t.Parallel()
	source := renderFormTree(t, chadcn.Field(chadcn.FieldOptions{
		Label:       "Name on card",
		Description: "As printed on the card",
		For:         "card-name",
		Required:    true,
		Error:       "Enter the name",
		Control:     chadcn.Input("card-name", ""),
	}))
	for _, want := range []string{
		`<label class="of-field-label" for="card-name">`,
		`required="required"`,
		`aria-required="true"`,
		`aria-invalid="true"`,
		`aria-describedby="card-name-error"`,
		`<p class="of-field-error" id="card-name-error" role="alert">Enter the name</p>`,
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("field missing %q:\n%s", want, source)
		}
	}

	area := renderFormTree(t, chadcn.Field(chadcn.FieldOptions{
		For:      "note",
		Required: true,
		Error:    "Required",
		Control:  chadcn.Textarea("note", ""),
	}))
	for _, want := range []string{`required="required"`, `aria-required="true"`, `aria-describedby="note-error"`} {
		if !strings.Contains(area, want) {
			t.Fatalf("textarea field missing %q:\n%s", want, area)
		}
	}
}

func TestTextValidationFieldSkipsHiddenControl(t *testing.T) {
	t.Parallel()
	source := renderFormTree(t, chadcn.Field(chadcn.FieldOptions{
		Label:    "Webhook",
		For:      "hook",
		Required: true,
		Error:    "Enter a URL",
		Control: chadcn.InputGroup(chadcn.InputGroupOptions{
			ID:     "hook",
			Prefix: ui.Text("https://"),
		}),
	}))
	if strings.Contains(source, "aria-required") || strings.Contains(source, "aria-invalid") || strings.Contains(source, "required=") {
		t.Fatalf("wrapper was marked:\n%s", source)
	}
	if !strings.Contains(source, `for="hook"`) || !strings.Contains(source, "Enter a URL") {
		t.Fatalf("label or error dropped:\n%s", source)
	}
}

func TestTextValidationInputGroupFillsRow(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	page, err := ui.NewPage(ownframe.Config{
		Width:  480,
		Height: 200,
		Theme:  chadcn.Stylesheet,
	}, ui.Div().Width(ui.Px(360)).Children(
		chadcn.InputGroup(chadcn.InputGroupOptions{
			ID:     "handle",
			Value:  "acme",
			Prefix: ui.Text("@"),
		}),
		chadcn.InputGroup(chadcn.InputGroupOptions{
			ID:     "hook",
			Value:  "hooks.acme.dev",
			Prefix: ui.Text("https://"),
			Addon:  chadcn.Button("Copy", chadcn.ButtonOptions{Size: chadcn.ButtonSmall}).ID("copy"),
		}),
	))
	if err != nil {
		t.Fatal(err)
	}
	if err := page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	handle := boxByID(t, page, "handle")
	handleRow := boxAt(t, page, handle.ParentIndex)
	if handle.X+handle.W < handleRow.X+handleRow.W-2 {
		t.Fatalf("bare field right %.1f stops before row %.1f", handle.X+handle.W, handleRow.X+handleRow.W)
	}
	if handle.H+1 < handleRow.H-2 {
		t.Fatalf("bare field height %.1f does not fill row %.1f", handle.H, handleRow.H)
	}

	hook := boxByID(t, page, "hook")
	button := boxByID(t, page, "copy")
	if hook.X+hook.W > button.X {
		t.Fatalf("field covers addon button: field right %.1f button %.1f", hook.X+hook.W, button.X)
	}
	if hook.W < 120 {
		t.Fatalf("field did not fill the space before the addon: width %.1f", hook.W)
	}
}

func boxByID(t *testing.T, page *ownframe.Page, id string) ownframe.Box {
	t.Helper()
	for _, box := range page.Boxes() {
		if box.ID == id && box.W > 0 && box.H > 0 {
			return box
		}
	}
	t.Fatalf("no box id=%q", id)
	return ownframe.Box{}
}

func boxAt(t *testing.T, page *ownframe.Page, index int) ownframe.Box {
	t.Helper()
	boxes := page.Boxes()
	if index < 0 || index >= len(boxes) {
		t.Fatalf("box index %d of %d", index, len(boxes))
	}
	return boxes[index]
}

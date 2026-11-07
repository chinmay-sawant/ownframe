package chadcn_test

import (
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// Field and TextField derive stable part ids from the control id so a label,
// description, or error text edit takes the View.Update patch path instead of
// falling back to a full document render.
func fix2FieldRoot(label, description, errText string) *ui.Element {
	return probeFormRoot(chadcn.Field(chadcn.FieldOptions{
		Label:       label,
		Description: description,
		Error:       errText,
		For:         "card-name",
		Control:     chadcn.Input("card-name", ""),
	}))
}

func TestProbePatchFix2Field(t *testing.T) {
	t.Parallel()

	t.Run("LabelEdit", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, fix2FieldRoot("Name on card", "As printed", ""))
		expectProbePatch(t, v, ctx, fix2FieldRoot("Name on embossed card", "As printed", ""))
		if source := v.Page.HTML(); !strings.Contains(source, "Name on embossed card") {
			t.Fatalf("patched page missing new label:\n%s", source)
		}
	})

	t.Run("DescriptionEdit", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, fix2FieldRoot("Name on card", "As printed", ""))
		expectProbePatch(t, v, ctx, fix2FieldRoot("Name on card", "Exactly as embossed", ""))
		if source := v.Page.HTML(); !strings.Contains(source, "Exactly as embossed") {
			t.Fatalf("patched page missing new description:\n%s", source)
		}
	})

	t.Run("ErrorEdit", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, fix2FieldRoot("Card number", "16 digits", "Too short"))
		expectProbePatch(t, v, ctx, fix2FieldRoot("Card number", "16 digits", "Enter all 16 digits"))
		if source := v.Page.HTML(); !strings.Contains(source, "Enter all 16 digits") {
			t.Fatalf("patched page missing new error:\n%s", source)
		}
	})

	t.Run("TextFieldLabelEdit", func(t *testing.T) {
		t.Parallel()
		build := func(label string) *ui.Element {
			return probeFormRoot(chadcn.TextField("message", label, "hello"))
		}
		v, ctx := mountProbeView(t, build("Message"))
		expectProbePatch(t, v, ctx, build("Subject"))
		if source := v.Page.HTML(); !strings.Contains(source, ">Subject</label>") {
			t.Fatalf("patched page missing new text field label:\n%s", source)
		}
		if got := v.Page.Form().FormValue("message"); got != "hello" {
			t.Fatalf("live input value = %q, want %q", got, "hello")
		}
	})

	t.Run("LabelEditKeepsLiveValue", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, fix2FieldRoot("Name on card", "As printed", ""))
		if err := v.Page.Form().Focus(ctx, "card-name"); err != nil {
			t.Fatal(err)
		}
		if err := v.Page.Edit().Type(ctx, "Ada"); err != nil {
			t.Fatal(err)
		}
		expectProbePatch(t, v, ctx, fix2FieldRoot("Name on embossed card", "As printed", ""))
		if got := v.Page.Form().FormValue("card-name"); got != "Ada" {
			t.Fatalf("live input value = %q, want %q", got, "Ada")
		}
	})

	t.Run("IdenticalNoop", func(t *testing.T) {
		t.Parallel()
		build := func() *ui.Element { return fix2FieldRoot("Name on card", "As printed", "Too short") }
		v, ctx := mountProbeView(t, build())
		expectProbeNoop(t, v, ctx, build())
	})
}

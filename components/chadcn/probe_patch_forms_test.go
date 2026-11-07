package chadcn_test

import (
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

func probeFormRoot(el *ui.Element) *ui.Element {
	return ui.Div().ID("probe-root").Children(el)
}

func selVals() []chadcn.SelectValue {
	return []chadcn.SelectValue{{Value: "a", Label: "Alpha"}, {Value: "b", Label: "Beta"}}
}

func probeFieldSeen(label, description, errText string) *ui.Element {
	return probeFormRoot(chadcn.Field(chadcn.FieldOptions{
		Label:       label,
		Description: description,
		Error:       errText,
		For:         "card-name",
		Control:     chadcn.Input("card-name", ""),
	}))
}

func TestProbePatchForms(t *testing.T) {
	t.Parallel()

	t.Run("InputNoop", func(t *testing.T) {
		t.Parallel()
		build := func() *ui.Element { return probeFormRoot(chadcn.Input("email", "ada@example.com")) }
		v, ctx := mountProbeView(t, build())
		expectProbeNoop(t, v, ctx, build())
	})

	t.Run("TextareaNoop", func(t *testing.T) {
		t.Parallel()
		build := func() *ui.Element { return probeFormRoot(chadcn.Textarea("note", "line one")) }
		v, ctx := mountProbeView(t, build())
		expectProbeNoop(t, v, ctx, build())
	})

	t.Run("NativeSelectNoop", func(t *testing.T) {
		t.Parallel()
		build := func() *ui.Element { return probeFormRoot(chadcn.NativeSelect("ns-1", selVals(), "a")) }
		v, ctx := mountProbeView(t, build())
		expectProbeNoop(t, v, ctx, build())
	})

	t.Run("InputOTPNoop", func(t *testing.T) {
		t.Parallel()
		build := func() *ui.Element {
			return probeFormRoot(chadcn.InputOTP(chadcn.InputOTPOptions{ID: "code", Length: 4, Value: "12"}))
		}
		v, ctx := mountProbeView(t, build())
		expectProbeNoop(t, v, ctx, build())
	})

	t.Run("InputGroupNoop", func(t *testing.T) {
		t.Parallel()
		build := func() *ui.Element {
			return probeFormRoot(chadcn.InputGroup(chadcn.InputGroupOptions{
				ID: "search", Label: "Search", Value: "repo",
				Prefix: ui.Text("@"),
			}))
		}
		v, ctx := mountProbeView(t, build())
		expectProbeNoop(t, v, ctx, build())
	})

	t.Run("SearchFieldNoop", func(t *testing.T) {
		t.Parallel()
		build := func() *ui.Element { return probeFormRoot(chadcn.SearchField("mail-search", "Search mail", "unread")) }
		v, ctx := mountProbeView(t, build())
		expectProbeNoop(t, v, ctx, build())
	})

	t.Run("TextFieldNoop", func(t *testing.T) {
		t.Parallel()
		build := func() *ui.Element { return probeFormRoot(chadcn.TextField("message", "Message", "hello")) }
		v, ctx := mountProbeView(t, build())
		expectProbeNoop(t, v, ctx, build())
	})

	t.Run("FieldNoop", func(t *testing.T) {
		t.Parallel()
		build := func() *ui.Element { return probeFieldSeen("Name", "As printed", "") }
		v, ctx := mountProbeView(t, build())
		expectProbeNoop(t, v, ctx, build())
	})

	t.Run("ChoiceNoop", func(t *testing.T) {
		t.Parallel()
		build := func() *ui.Element {
			return probeFormRoot(chadcn.ChoiceWithID("choice-1", "Nightly backups", "Keeps 14 days.", false))
		}
		v, ctx := mountProbeView(t, build())
		expectProbeNoop(t, v, ctx, build())
	})

	t.Run("InputPlaceholder", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, probeFormRoot(chadcn.InputWithOptions("email", "", chadcn.InputOptions{
			Placeholder: "you@example.com",
		})))
		expectProbePatch(t, v, ctx, probeFormRoot(chadcn.InputWithOptions("email", "", chadcn.InputOptions{
			Placeholder: "name@example.com",
		})))
	})

	t.Run("SearchFieldLabel", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, probeFormRoot(chadcn.SearchField("mail-search", "Search mail", "")))
		expectProbePatch(t, v, ctx, probeFormRoot(chadcn.SearchField("mail-search", "Find mail", "")))
	})

	t.Run("TextareaPlaceholder", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, probeFormRoot(chadcn.TextareaWithOptions("note", "", chadcn.TextareaOptions{
			Placeholder: "Add a note",
		})))
		expectProbePatch(t, v, ctx, probeFormRoot(chadcn.TextareaWithOptions("note", "", chadcn.TextareaOptions{
			Placeholder: "Add a longer note",
		})))
	})

	t.Run("InputOTPValue", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, probeFormRoot(chadcn.InputOTP(chadcn.InputOTPOptions{
			ID: "code", Length: 4, Value: "12",
		})))
		expectProbePatch(t, v, ctx, probeFormRoot(chadcn.InputOTP(chadcn.InputOTPOptions{
			ID: "code", Length: 4, Value: "34",
		})))
	})

	t.Run("StandaloneDescription", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, probeFormRoot(chadcn.FieldDescription("Short help").ID("help-1")))
		expectProbePatch(t, v, ctx, probeFormRoot(chadcn.FieldDescription("Longer help text").ID("help-1")))
	})

	t.Run("StandaloneError", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, probeFormRoot(chadcn.FieldError("Required").ID("err-1")))
		expectProbePatch(t, v, ctx, probeFormRoot(chadcn.FieldError("Enter a value").ID("err-1")))
	})

	t.Run("FieldErrorEdit", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, probeFieldSeen("Card number", "16 digits", "Too short"))
		expectProbePatch(t, v, ctx, probeFieldSeen("Card number", "16 digits", "Enter all 16 digits"))
	})

	t.Run("FieldErrorEditWrappedControl", func(t *testing.T) {
		t.Parallel()
		build := func(errText string) *ui.Element {
			return probeFormRoot(chadcn.Field(chadcn.FieldOptions{
				Label: "Webhook",
				For:   "hook",
				Error: errText,
				Control: chadcn.InputGroup(chadcn.InputGroupOptions{
					ID:     "hook",
					Prefix: ui.Text("https://"),
				}),
			}))
		}
		v, ctx := mountProbeView(t, build("Enter a URL"))
		expectProbePatch(t, v, ctx, build("Use an https URL"))
	})

	t.Run("InputGroupDisabled", func(t *testing.T) {
		t.Parallel()
		build := func(disabled bool) *ui.Element {
			return probeFormRoot(chadcn.InputGroup(chadcn.InputGroupOptions{
				ID: "search", Label: "Search", Value: "repo",
				Prefix:   ui.Text("@"),
				Disabled: disabled,
			}))
		}
		v, ctx := mountProbeView(t, build(false))
		expectProbePatch(t, v, ctx, build(true))
	})

	t.Run("ChoiceCaption", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, probeFormRoot(chadcn.ChoiceWithID("choice-1", "Nightly backups", "Keeps 7 days.", false)))
		expectProbePatch(t, v, ctx, probeFormRoot(chadcn.ChoiceWithID("choice-1", "Nightly backups", "Keeps 14 days.", false)))
	})

	t.Run("NativeSelectKeepsLiveValue", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, probeFormRoot(chadcn.NativeSelect("ns-1", selVals(), "a")))
		if err := v.Update(ctx, probeFormRoot(chadcn.NativeSelect("ns-1", selVals(), "b"))); err != nil {
			t.Fatal(err)
		}
		// The rebuilt tree carries the new selection in the rendered source.
		if source := v.Page.HTML(); !strings.Contains(source, `<option selected="selected" value="b">`) {
			t.Fatalf("rebuilt source lost the new selection:\n%s", source)
		}
		// Form sync owns the live value, so the untouched control keeps "a".
		if got := v.Page.Form().FormValue("ns-1"); got != "a" {
			t.Fatalf("live select value = %q, want %q", got, "a")
		}
		found := false
		for _, box := range v.Page.Boxes() {
			if box.ID == "ns-1" && box.W > 0 && box.H > 0 {
				found = true
			}
		}
		if !found {
			t.Fatal("no laid-out box for ns-1")
		}
	})
}

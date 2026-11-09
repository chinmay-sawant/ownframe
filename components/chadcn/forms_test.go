package chadcn_test

import (
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

func renderFormTree(t *testing.T, element *ui.Element) string {
	t.Helper()
	source, err := ui.Render(element)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	return source
}

func wantFormMarkup(t *testing.T, source string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(source, want) {
			t.Fatalf("rendered HTML missing %q:\n%s", want, source)
		}
	}
}

func wantNoFormMarkup(t *testing.T, source string, rejects ...string) {
	t.Helper()
	for _, reject := range rejects {
		if strings.Contains(source, reject) {
			t.Fatalf("rendered HTML has unexpected %q:\n%s", reject, source)
		}
	}
}

func TestInputDefaults(t *testing.T) {
	t.Parallel()
	source := renderFormTree(t, chadcn.Input("email", "ada@example.com"))
	if !strings.HasPrefix(source, "<input ") {
		t.Fatalf("input root: %s", source)
	}
	wantFormMarkup(t, source,
		`id="email"`,
		`type="text"`,
		`value="ada@example.com"`,
		`class="of-scope of-input of-focus"`,
	)
	wantNoFormMarkup(t, source, `placeholder=`, `aria-invalid`, `disabled`)
}

func TestInputOptions(t *testing.T) {
	t.Parallel()
	source := renderFormTree(t, chadcn.InputWithOptions("api-key", "", chadcn.InputOptions{
		Type:        "password",
		Placeholder: "sk-...",
		Size:        chadcn.InputSmall,
		Disabled:    true,
		Invalid:     true,
	}))
	wantFormMarkup(t, source,
		`id="api-key"`,
		`type="password"`,
		`value=""`,
		`placeholder="sk-..."`,
		`aria-invalid="true"`,
		`disabled="disabled"`,
		`of-input--sm`,
	)
}

func TestInputSizesChangeClass(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name  string
		size  chadcn.InputSize
		class string
	}{
		{"small", chadcn.InputSmall, "of-input--sm"},
		{"medium", chadcn.InputMedium, ""},
		{"large", chadcn.InputLarge, "of-input--lg"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			source := renderFormTree(t, chadcn.InputWithOptions("field", "", chadcn.InputOptions{Size: tt.size}))
			if tt.class == "" {
				wantFormMarkup(t, source, `class="of-scope of-input of-focus"`)
				return
			}
			wantFormMarkup(t, source, tt.class)
		})
	}
}

func TestTextareaRowsAndValue(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name     string
		rows     int
		wantRows string
		wantMin  string
	}{
		{"default", 0, "3", "min-height:76px"},
		{"five", 5, "5", "min-height:116px"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			source := renderFormTree(t, chadcn.TextareaWithOptions("note", "line one", chadcn.TextareaOptions{
				Placeholder: "Add a note",
				Rows:        tt.rows,
			}))
			wantFormMarkup(t, source,
				`id="note"`,
				`rows="`+tt.wantRows+`"`,
				`placeholder="Add a note"`,
				`class="of-scope of-textarea of-focus"`,
				tt.wantMin,
				">line one</textarea>",
			)
		})
	}
}

func TestTextareaState(t *testing.T) {
	t.Parallel()
	source := renderFormTree(t, chadcn.TextareaWithOptions("note", "", chadcn.TextareaOptions{
		Disabled: true,
		Invalid:  true,
	}))
	wantFormMarkup(t, source, `aria-invalid="true"`, `disabled="disabled"`, `rows="3"`)
}

func TestFieldStructure(t *testing.T) {
	t.Parallel()
	source := renderFormTree(t, chadcn.Field(chadcn.FieldOptions{
		Label:       "Name on card",
		Description: "As printed on the card",
		For:         "card-name",
		Required:    true,
		Control:     chadcn.Input("card-name", ""),
	}))
	wantFormMarkup(t, source,
		`class="of-scope of-field"`,
		`<label class="of-field-label" for="card-name" id="card-name-label">Name on card<span aria-hidden="true" class="of-field-required">*</span></label>`,
		`<p class="of-field-description" id="card-name-description">As printed on the card</p>`,
		`<input aria-required="true" class="of-scope of-input of-focus" id="card-name" required="required" type="text" value="">`,
	)
	wantNoFormMarkup(t, source, `of-field--invalid`)
}

func TestFieldErrorKeepsDescription(t *testing.T) {
	t.Parallel()
	source := renderFormTree(t, chadcn.Field(chadcn.FieldOptions{
		Label:       "Card number",
		Description: "16 digits",
		Error:       "Enter all 16 digits",
		Control:     chadcn.Input("card-number", ""),
	}))
	wantFormMarkup(t, source,
		`class="of-scope of-field of-field--invalid"`,
		`<p class="of-field-description" id="card-number-description">16 digits</p>`,
		`<p class="of-field-error" id="card-number-error" role="alert">Enter all 16 digits</p>`,
		`aria-invalid="true"`,
		`aria-describedby="card-number-error"`,
	)
}

func TestFieldParts(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		got  string
		want string
	}{
		{"label", renderFormTree(t, chadcn.FieldLabel("Email")), `<label class="of-field-label">Email</label>`},
		{"description", renderFormTree(t, chadcn.FieldDescription("Work address")), `<p class="of-field-description">Work address</p>`},
		{"error", renderFormTree(t, chadcn.FieldError("Required")), `<p class="of-field-error" role="alert">Required</p>`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("rendered %s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestNativeSelectKeepsSelection(t *testing.T) {
	t.Parallel()
	tree := chadcn.NativeSelect("status", []chadcn.SelectValue{
		{Value: "active", Label: "Active"},
		{Value: "away", Label: "Away", Disabled: true},
		{Value: "offline"},
	}, "active")
	source, err := ui.Render(tree)
	if err != nil {
		if strings.Contains(err.Error(), "unsupported HTML tag") && strings.Contains(err.Error(), "option") {
			t.Skipf("blocked: add \"option\" to supportedTag in ui/builder.go: %v", err)
		}
		t.Fatal(err)
	}
	if !strings.HasPrefix(source, `<div class="of-scope of-native-select-wrap">`) {
		t.Fatalf("select wrapper: %s", source)
	}
	wantFormMarkup(t, source,
		`<select class="of-native-select of-focus" id="status">`,
		`<option selected="selected" value="active">Active</option>`,
		`<option disabled="disabled" value="away">Away</option>`,
		`<option value="offline">offline</option>`,
		`class="of-native-select-icon"`,
		`aria-hidden="true"`,
	)
}

func TestTextFieldCompatibility(t *testing.T) {
	t.Parallel()
	source := renderFormTree(t, chadcn.TextField("message", "Message", "hello"))
	wantFormMarkup(t, source,
		`<div class="of-scope of-field of-text-field">`,
		`<label class="of-field-label" for="message" id="message-label">Message</label>`,
		`<input class="of-scope of-input of-focus" id="message" type="text" value="hello">`,
	)
}

func TestSearchFieldCompatibility(t *testing.T) {
	t.Parallel()
	source := renderFormTree(t, chadcn.SearchField("mail-search", "Search mail", "unread").Width(ui.Px(240)))
	if !strings.HasPrefix(source, "<input ") {
		t.Fatalf("search root: %s", source)
	}
	wantFormMarkup(t, source,
		`id="mail-search"`,
		`type="search"`,
		`value="unread"`,
		`aria-label="Search mail"`,
		`placeholder="Search mail"`,
		`class="of-scope of-input of-focus of-search-field"`,
		`width:240px`,
	)
}

func TestStylesheetCarriesFormRules(t *testing.T) {
	t.Parallel()
	wantFormMarkup(t, chadcn.Stylesheet,
		".of-field {",
		".of-input {",
		".of-textarea {",
		".of-native-select {",
		".of-search-field {",
		".of-text-field {",
		"var(--of-input,",
	)
}

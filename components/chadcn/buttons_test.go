package chadcn_test

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/theme"
)

func renderButtonTree(t *testing.T, root *ui.Element) string {
	t.Helper()
	source, err := ui.Render(root)
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func TestButtonVariantsAndSizesMapToClasses(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		variant chadcn.ButtonVariant
		want    string
	}{
		{chadcn.ButtonTonal, "of-button--tonal"},
		{chadcn.ButtonFilled, "of-button--filled"},
		{chadcn.ButtonOutlined, "of-button--outlined"},
		{chadcn.ButtonText, "of-button--text"},
		{chadcn.ButtonDanger, "of-button--danger"},
	} {
		source := renderButtonTree(t, chadcn.Button("Save", chadcn.ButtonOptions{Variant: tt.variant}))
		if !strings.Contains(source, tt.want) {
			t.Fatalf("variant %d missing class %q in %s", tt.variant, tt.want, source)
		}
	}
	for _, tt := range []struct {
		size chadcn.ButtonSize
		want string
	}{
		{chadcn.ButtonSmall, "of-button--small"},
		{chadcn.ButtonMedium, "of-button--medium"},
		{chadcn.ButtonLarge, "of-button--large"},
	} {
		source := renderButtonTree(t, chadcn.Button("Save", chadcn.ButtonOptions{Size: tt.size}))
		if !strings.Contains(source, tt.want) {
			t.Fatalf("size %d missing class %q in %s", tt.size, tt.want, source)
		}
	}
}

func TestButtonDefaultsAndIcon(t *testing.T) {
	t.Parallel()
	source := renderButtonTree(t, chadcn.Button("Save", chadcn.ButtonOptions{}))
	for _, want := range []string{
		`<button`,
		`type="button"`,
		`class="of-button of-button--tonal of-button--medium"`,
		`<span class="of-button-label">Save</span>`,
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("default button missing %q in %s", want, source)
		}
	}
	if strings.Contains(source, "aria-busy") || strings.Contains(source, "disabled") {
		t.Fatalf("default button carries a state: %s", source)
	}

	withIcon := renderButtonTree(t, chadcn.Button("Send", chadcn.ButtonOptions{Icon: ui.Span().Text("+")}))
	if !strings.Contains(withIcon, `<span aria-hidden="true" class="of-button-icon">`) ||
		!strings.Contains(withIcon, `<span>+</span>`) {
		t.Fatalf("icon slot missing: %s", withIcon)
	}
	if strings.Index(withIcon, "of-button-icon") > strings.Index(withIcon, "of-button-label") {
		t.Fatalf("icon should paint before the label: %s", withIcon)
	}
}

func TestButtonLoadingAndDisabled(t *testing.T) {
	t.Parallel()
	source := renderButtonTree(t, chadcn.Button("Delete", chadcn.ButtonOptions{
		Variant:  chadcn.ButtonDanger,
		Disabled: true,
		Loading:  true,
	}))
	for _, want := range []string{
		`aria-busy="true"`,
		`disabled="disabled"`,
		`class="of-button of-button--danger of-button--medium"`,
		`class="of-button-spinner"`,
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("loading button missing %q in %s", want, source)
		}
	}

	iconOnly := renderButtonTree(t, chadcn.Button("", chadcn.ButtonOptions{Icon: ui.Span().Text("^")}))
	if !strings.Contains(iconOnly, "of-button--icon") || strings.Contains(iconOnly, "of-button-label") {
		t.Fatalf("icon-only button markup: %s", iconOnly)
	}
}

func TestButtonAppliesThemeOverrides(t *testing.T) {
	t.Parallel()
	source := renderButtonTree(t, chadcn.Button("Filled", chadcn.ButtonOptions{
		Variant:   chadcn.ButtonFilled,
		Overrides: theme.Overrides{"--of-primary": "#123456"},
	}))
	if !strings.Contains(source, `style="--of-primary:#123456;"`) {
		t.Fatalf("override missing: %s", source)
	}
}

func TestButtonGroupJoinsChildren(t *testing.T) {
	t.Parallel()
	source := renderButtonTree(t, chadcn.ButtonGroup(chadcn.ButtonGroupOptions{Label: "View"},
		chadcn.Button("List", chadcn.ButtonOptions{}),
		chadcn.Button("Grid", chadcn.ButtonOptions{}),
	))
	for _, want := range []string{`class="of-button-group"`, `role="group"`, `aria-label="View"`, "List", "Grid"} {
		if !strings.Contains(source, want) {
			t.Fatalf("button group missing %q in %s", want, source)
		}
	}
	if strings.Count(source, "<button") != 2 {
		t.Fatalf("button group should keep both buttons: %s", source)
	}

	vertical := renderButtonTree(t, chadcn.ButtonGroup(chadcn.ButtonGroupOptions{Vertical: true},
		chadcn.Button("Top", chadcn.ButtonOptions{}),
		chadcn.Button("Bottom", chadcn.ButtonOptions{}),
	))
	if !strings.Contains(vertical, `class="of-button-group of-button-group--vertical"`) {
		t.Fatalf("vertical group class missing: %s", vertical)
	}
}

func TestTogglePressedState(t *testing.T) {
	t.Parallel()
	pressed := renderButtonTree(t, chadcn.Toggle("Bookmark", true, chadcn.ToggleOptions{Icon: ui.Span().Text("*")}))
	for _, want := range []string{
		`aria-pressed="true"`,
		`class="of-toggle of-toggle--medium"`,
		`<span aria-hidden="true" class="of-toggle-icon">`,
		`<span class="of-toggle-label">Bookmark</span>`,
		`type="button"`,
	} {
		if !strings.Contains(pressed, want) {
			t.Fatalf("pressed toggle missing %q in %s", want, pressed)
		}
	}

	idle := renderButtonTree(t, chadcn.Toggle("Bookmark", false))
	if !strings.Contains(idle, `aria-pressed="false"`) {
		t.Fatalf("idle toggle lost aria-pressed: %s", idle)
	}

	disabled := renderButtonTree(t, chadcn.Toggle("Bold", false, chadcn.ToggleOptions{Disabled: true, Size: chadcn.ButtonSmall}))
	if !strings.Contains(disabled, `disabled="disabled"`) || !strings.Contains(disabled, "of-toggle--small") {
		t.Fatalf("disabled toggle markup: %s", disabled)
	}

	iconOnly := renderButtonTree(t, chadcn.Toggle("", true, chadcn.ToggleOptions{Icon: ui.Span().Text("B")}))
	if !strings.Contains(iconOnly, "of-toggle--icon") || strings.Contains(iconOnly, "of-toggle-label") {
		t.Fatalf("icon-only toggle markup: %s", iconOnly)
	}
}

func TestToggleGroupRendersSelection(t *testing.T) {
	t.Parallel()
	source := renderButtonTree(t, chadcn.ToggleGroup(chadcn.ToggleGroupOptions{
		Label:    "Text style",
		Selected: []string{"bold"},
		Items: []chadcn.ToggleItem{
			{ID: "bold-toggle", Value: "bold", Label: "B"},
			{ID: "italic-toggle", Value: "italic", Label: "I"},
			{ID: "underline-toggle", Value: "underline", Label: "U", Disabled: true},
		},
	}))
	for _, want := range []string{
		`role="group"`,
		`aria-label="Text style"`,
		`id="bold-toggle"`,
		`data-value="bold"`,
		`data-value="italic"`,
		`data-value="underline"`,
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("toggle group missing %q in %s", want, source)
		}
	}
	if got := strings.Count(source, `aria-pressed="true"`); got != 1 {
		t.Fatalf("pressed count=%d want 1: %s", got, source)
	}
	if got := strings.Count(source, `aria-pressed="false"`); got != 2 {
		t.Fatalf("unpressed count=%d want 2: %s", got, source)
	}
	if !strings.Contains(source, `disabled="disabled"`) {
		t.Fatalf("disabled group item lost its state: %s", source)
	}

	vertical := renderButtonTree(t, chadcn.ToggleGroup(chadcn.ToggleGroupOptions{
		Vertical: true,
		Multiple: true,
		Selected: []string{"a", "b"},
		Items:    []chadcn.ToggleItem{{Value: "a", Label: "A"}, {Value: "b", Label: "B"}},
	}))
	if !strings.Contains(vertical, `class="of-toggle-group of-toggle-group--vertical"`) {
		t.Fatalf("vertical group class missing: %s", vertical)
	}
	if !strings.Contains(vertical, `data-multiple="true"`) {
		t.Fatalf("multiple flag missing: %s", vertical)
	}
	if got := strings.Count(vertical, `aria-pressed="true"`); got != 2 {
		t.Fatalf("multiple selection pressed count=%d want 2: %s", got, vertical)
	}
}

func TestToggleGroupWiresSelectHandler(t *testing.T) {
	t.Parallel()
	source := renderButtonTree(t, chadcn.ToggleGroup(chadcn.ToggleGroupOptions{
		Selected: []string{"bold"},
		OnSelect: func(context.Context, string) error { return nil },
		Items: []chadcn.ToggleItem{
			{ID: "bold-toggle", Value: "bold", Label: "B"},
			{ID: "italic-toggle", Value: "italic", Label: "I"},
		},
	}))
	if got := strings.Count(source, "data-action="); got != 2 {
		t.Fatalf("select handler wiring count=%d want 2: %s", got, source)
	}
	for _, want := range []string{`id="bold-toggle"`, `id="italic-toggle"`} {
		if !strings.Contains(source, want) {
			t.Fatalf("group item lost stable id %q in %s", want, source)
		}
	}

	withoutHandler := renderButtonTree(t, chadcn.ToggleGroup(chadcn.ToggleGroupOptions{
		Items: []chadcn.ToggleItem{{Value: "bold", Label: "B"}},
	}))
	if strings.Contains(withoutHandler, "data-action=") {
		t.Fatalf("group without OnSelect should not emit an action: %s", withoutHandler)
	}
}

func TestStylesheetCarriesButtonRules(t *testing.T) {
	t.Parallel()
	for _, want := range []string{
		".of-button {",
		".of-button--filled",
		".of-button--danger",
		".of-button--icon",
		".of-button-spinner",
		".of-button-group--vertical",
		".of-toggle[aria-pressed=\"true\"]",
		".of-toggle-group",
	} {
		if !strings.Contains(chadcn.Stylesheet, want) {
			t.Fatalf("stylesheet missing %q", want)
		}
	}
}

package chadcn_test

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/pagetest"
	"github.com/chinmay-sawant/ownframe/ui"
)

func renderFormControlsMarkup(t *testing.T, root *ui.Element) string {
	t.Helper()
	source, err := ui.Render(root)
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func TestCheckboxControlledState(t *testing.T) {
	t.Parallel()
	unchecked := renderFormControlsMarkup(t, chadcn.Checkbox("Accept terms", false))
	for _, want := range []string{
		`class="of-checkbox"`,
		`class="of-checkbox-label"`,
		`role="checkbox"`,
		`aria-checked="false"`,
		`type="button"`,
	} {
		if !strings.Contains(unchecked, want) {
			t.Fatalf("unchecked checkbox missing %q: %s", want, unchecked)
		}
	}
	if strings.Contains(unchecked, "✓") {
		t.Fatal("unchecked checkbox painted a check mark")
	}

	checked := renderFormControlsMarkup(t, chadcn.Checkbox("Accept terms", true))
	for _, want := range []string{
		`aria-checked="true"`,
		`class="of-checkbox of-checkbox--checked"`,
		`class="of-checkbox-mark"`,
		"✓",
	} {
		if !strings.Contains(checked, want) {
			t.Fatalf("checked checkbox missing %q: %s", want, checked)
		}
	}

	withOptions := renderFormControlsMarkup(t, chadcn.CheckboxWithOptions("Notifications", false, chadcn.CheckboxOptions{
		Description: "You can enable or disable notifications at any time.",
		Disabled:    true,
	}))
	for _, want := range []string{
		`class="of-checkbox-description"`,
		"You can enable or disable notifications at any time.",
		`disabled="disabled"`,
	} {
		if !strings.Contains(withOptions, want) {
			t.Fatalf("checkbox options missing %q: %s", want, withOptions)
		}
	}
}

func TestSwitchControlledState(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name    string
		checked bool
		want    string
	}{
		{name: "off", want: `aria-checked="false"`},
		{name: "on", checked: true, want: `aria-checked="true"`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			source := renderFormControlsMarkup(t, chadcn.Switch("Airplane mode", tt.checked))
			for _, want := range []string{
				`class="of-switch"`,
				`class="of-switch-track"`,
				`class="of-switch-thumb"`,
				`class="of-switch-label"`,
				`role="switch"`,
				tt.want,
			} {
				if !strings.Contains(source, want) {
					t.Fatalf("switch missing %q: %s", want, source)
				}
			}
		})
	}

	disabled := renderFormControlsMarkup(t, chadcn.SwitchWithOptions("Airplane mode", false, chadcn.SwitchOptions{Disabled: true}))
	if !strings.Contains(disabled, `disabled="disabled"`) {
		t.Fatalf("disabled switch missing disabled attribute: %s", disabled)
	}
}

func TestRadioGroupControlledSelection(t *testing.T) {
	t.Parallel()
	source := renderFormControlsMarkup(t, chadcn.RadioGroup(chadcn.RadioGroupOptions{
		ID:    "density",
		Label: "Density",
		Value: "comfortable",
		Items: []chadcn.RadioItem{
			{Value: "default", Label: "Default"},
			{Value: "comfortable", Label: "Comfortable", Description: "Balanced spacing"},
			{Value: "compact", Label: "Compact", Disabled: true},
		},
	}))
	for _, want := range []string{
		`class="of-radio-group"`,
		`role="radiogroup"`,
		`aria-label="Density"`,
		`id="density-item-0"`,
		`id="density-item-1"`,
		`id="density-item-2"`,
		`class="of-radio-group-description"`,
		"Balanced spacing",
		`disabled="disabled"`,
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("radio group missing %q: %s", want, source)
		}
	}
	if got := strings.Count(source, `role="radio"`); got != 3 {
		t.Fatalf("radio buttons=%d want 3: %s", got, source)
	}
	if !strings.Contains(source, `class="of-radio-group-item of-radio-group-item--checked"`) {
		t.Fatalf("selected radio has no checked class: %s", source)
	}
	if got := strings.Count(source, `aria-checked="true"`); got != 1 {
		t.Fatalf("checked radios=%d want 1: %s", got, source)
	}
	if got := strings.Count(source, `aria-checked="false"`); got != 2 {
		t.Fatalf("unchecked radios=%d want 2: %s", got, source)
	}
}

func TestRadioGroupOnSelect(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	selected := "default"
	chosen := ""
	view, err := ui.Mount(ownframe.Config{Width: 480, Height: 320}, chadcn.RadioGroup(chadcn.RadioGroupOptions{
		ID:    "density",
		Value: selected,
		Items: []chadcn.RadioItem{
			{Value: "default", Label: "Default"},
			{Value: "comfortable", Label: "Comfortable"},
		},
		OnSelect: func(_ context.Context, value string) error {
			chosen = value
			selected = value
			return nil
		},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if err := view.Page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.Click(ctx, view.Page, "density-item-0"); err != nil {
		t.Fatal(err)
	}
	if chosen != "" {
		t.Fatalf("selected radio reported %q", chosen)
	}
	if err := pagetest.Click(ctx, view.Page, "density-item-1"); err != nil {
		t.Fatal(err)
	}
	if chosen != "comfortable" || selected != "comfortable" {
		t.Fatalf("OnSelect got %q selected=%q", chosen, selected)
	}
}

func TestSliderRangeSemantics(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		opts    chadcn.SliderOptions
		minimum string
		maximum string
		now     string
		fill    string
	}{
		{name: "default", opts: chadcn.SliderOptions{Label: "Volume"}, minimum: "0", maximum: "100", now: "0", fill: "width:0%"},
		{name: "middle", opts: chadcn.SliderOptions{Label: "Volume", Value: 25, Min: 0, Max: 50, Step: 5}, minimum: "0", maximum: "50", now: "25", fill: "width:50%"},
		{name: "snapped", opts: chadcn.SliderOptions{Label: "Volume", Value: 23, Min: 0, Max: 50, Step: 5}, minimum: "0", maximum: "50", now: "25", fill: "width:50%"},
		{name: "clamped", opts: chadcn.SliderOptions{Label: "Volume", Value: 75, Min: 0, Max: 50}, minimum: "0", maximum: "50", now: "50", fill: "width:100%"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			source := renderFormControlsMarkup(t, chadcn.Slider(tt.opts))
			for _, want := range []string{
				`class="of-slider"`,
				`role="slider"`,
				`class="of-slider-track"`,
				`class="of-slider-rail"`,
				`class="of-slider-range"`,
				`class="of-slider-thumb"`,
				`aria-label="Volume"`,
				`aria-valuemin="` + tt.minimum + `"`,
				`aria-valuemax="` + tt.maximum + `"`,
				`aria-valuenow="` + tt.now + `"`,
				`aria-orientation="horizontal"`,
				`tabindex="0"`,
				tt.fill,
			} {
				if !strings.Contains(source, want) {
					t.Fatalf("slider missing %q: %s", want, source)
				}
			}
		})
	}
}

func TestSliderDisabledState(t *testing.T) {
	t.Parallel()
	source := renderFormControlsMarkup(t, chadcn.Slider(chadcn.SliderOptions{Label: "Volume", Value: 25, Disabled: true}))
	for _, want := range []string{`disabled="disabled"`, `aria-disabled="true"`, `aria-valuenow="25"`} {
		if !strings.Contains(source, want) {
			t.Fatalf("disabled slider missing %q: %s", want, source)
		}
	}
	if strings.Contains(source, `tabindex="0"`) {
		t.Fatalf("disabled slider stayed focusable: %s", source)
	}
}

func TestInputGroupSlots(t *testing.T) {
	t.Parallel()
	source := renderFormControlsMarkup(t, chadcn.InputGroup(chadcn.InputGroupOptions{
		ID:          "search",
		Label:       "Search",
		Type:        "search",
		Placeholder: "Search...",
		Prefix:      ui.Text("⌕"),
		Suffix:      ui.Text("12 results"),
		Addon:       ui.Text("⌘K"),
	}))
	for _, want := range []string{
		`class="of-input-group"`,
		`class="of-input-group-input"`,
		`id="search"`,
		`type="search"`,
		`aria-label="Search"`,
		`placeholder="Search..."`,
		`class="of-input-group-prefix"`,
		"⌕",
		`class="of-input-group-suffix"`,
		"12 results",
		`class="of-input-group-addon"`,
		"⌘K",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("input group missing %q: %s", want, source)
		}
	}

	disabled := renderFormControlsMarkup(t, chadcn.InputGroup(chadcn.InputGroupOptions{ID: "search", Disabled: true}))
	for _, want := range []string{`disabled="disabled"`, `aria-disabled="true"`} {
		if !strings.Contains(disabled, want) {
			t.Fatalf("disabled input group missing %q: %s", want, disabled)
		}
	}
}

func TestInputOTPSlots(t *testing.T) {
	t.Parallel()
	source := renderFormControlsMarkup(t, chadcn.InputOTP(chadcn.InputOTPOptions{ID: "code", Length: 4, Value: "12"}))
	for _, want := range []string{
		`class="of-input-otp"`,
		`role="group"`,
		`aria-label="One-time password"`,
		`id="code"`,
		`id="code-0"`,
		`id="code-3"`,
		`value="1"`,
		`value="2"`,
		`maxlength="1"`,
		`autocomplete="one-time-code"`,
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("input otp missing %q: %s", want, source)
		}
	}
	if got := strings.Count(source, `class="of-input-otp-slot"`); got != 4 {
		t.Fatalf("otp slots=%d want 4: %s", got, source)
	}

	disabled := renderFormControlsMarkup(t, chadcn.InputOTP(chadcn.InputOTPOptions{ID: "code", Label: "Verification code", Disabled: true}))
	if got := strings.Count(disabled, `class="of-input-otp-slot"`); got != 6 {
		t.Fatalf("default otp slots=%d want 6: %s", got, disabled)
	}
	if got := strings.Count(disabled, `disabled="disabled"`); got != 6 {
		t.Fatalf("disabled otp slots=%d want 6: %s", got, disabled)
	}
	for _, want := range []string{`aria-label="Verification code"`, `aria-disabled="true"`} {
		if !strings.Contains(disabled, want) {
			t.Fatalf("disabled otp missing %q: %s", want, disabled)
		}
	}
}

func TestInputOTPWritesToForm(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	view, err := ui.Mount(ownframe.Config{Width: 480, Height: 240}, chadcn.InputOTP(chadcn.InputOTPOptions{ID: "code", Length: 4}))
	if err != nil {
		t.Fatal(err)
	}
	if err := view.Page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := view.Page.Form().Focus(ctx, "code-1"); err != nil {
		t.Fatal(err)
	}
	if err := view.Page.Edit().Type(ctx, "9"); err != nil {
		t.Fatal(err)
	}
	if got := view.Page.Form().FormValue("code-1"); got != "9" {
		t.Fatalf("otp form value=%q want 9", got)
	}
}

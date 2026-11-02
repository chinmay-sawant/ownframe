package chadcn

import (
	"context"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe/ui"
)

// ComboboxItem is one option in a Combobox list. An empty Label falls back to
// Value.
type ComboboxItem struct {
	Value string
	Label string
}

// ComboboxOptions describes a controlled searchable select. The host owns
// Open, Query, and Value, updates them in the callbacks, and rebuilds the view.
//
// ID names the root. The text input uses ID+"-input", the chevron button uses
// ID+"-trigger", the listbox uses ID+"-panel", and each option uses
// ID+"-option-<index>".
type ComboboxOptions struct {
	ID           string
	Open         bool
	Query        string
	Value        string
	Placeholder  string
	Items        []ComboboxItem
	OnOpenChange func(ctx context.Context, open bool) error
	OnQuery      func(ctx context.Context, query string) error
	OnSelect     func(ctx context.Context, value string) error
}

// Combobox renders a text input with a chevron button and, when Open, a
// filtered listbox. The query matches labels case-insensitively.
//
// The input is a plain text field, so the host reads its live value from the
// page form and passes it back as Query. OnQuery stays available for hosts
// that dispatch their own input callbacks.
func Combobox(options ComboboxOptions) *ui.Element {
	placeholder := options.Placeholder
	if placeholder == "" {
		placeholder = "Select an option"
	}
	display := options.Query
	if label, ok := comboboxLabel(options.Items, options.Value); ok && !options.Open && display == "" {
		display = label
	}
	input := ui.Input().
		Attr("type", "text").
		Attr("role", "combobox").
		Attr("aria-autocomplete", "list").
		Attr("aria-expanded", ariaBool(options.Open)).
		Attr("aria-label", placeholder).
		Attr("placeholder", placeholder).
		Attr("value", display).
		Class(classes("of-combobox-input", "of-transition"))
	if inputID := partID(options.ID, "-input"); inputID != "" {
		input.ID(inputID)
	}
	panelID := partID(options.ID, "-panel")
	if options.Open && panelID != "" {
		input.Attr("aria-controls", panelID)
	}
	toggleLabel := "Show options"
	if options.Open {
		toggleLabel = "Hide options"
	}
	toggle := ui.Button("").
		Attr("type", "button").
		Attr("aria-label", toggleLabel).
		Attr("aria-haspopup", "listbox").
		Attr("aria-expanded", ariaBool(options.Open)).
		Class(classes("of-combobox-trigger", "of-transition")).
		Children(ui.Span().Class("of-combobox-chevron").Attr("aria-hidden", "true").Text("⌄"))
	if triggerID := partID(options.ID, "-trigger"); triggerID != "" {
		toggle.ID(triggerID)
	}
	if options.OnOpenChange != nil {
		open := options.Open
		toggle.OnClick(func(ctx context.Context) error {
			return options.OnOpenChange(ctx, !open)
		})
	}
	root := ui.Div().Class(classes("of-scope", "of-combobox")).Children(
		ui.Div().Class("of-combobox-control").Children(input, toggle),
	)
	if options.ID != "" {
		root.ID(options.ID)
	}
	if options.Open {
		root.Children(comboboxPanel(options, panelID))
	}
	return root
}

func comboboxPanel(options ComboboxOptions, panelID string) *ui.Element {
	panel := ui.Div().Class("of-combobox-panel").Attr("role", "listbox")
	if panelID != "" {
		panel.ID(panelID)
	}
	matches := 0
	for index, item := range options.Items {
		if !comboboxMatches(item, options.Query) {
			continue
		}
		matches++
		panel.Children(comboboxOption(options, index, item))
	}
	if matches == 0 {
		panel.Children(ui.Div().Class("of-combobox-empty").Text("No results"))
	}
	return panel
}

func comboboxOption(options ComboboxOptions, index int, item ComboboxItem) *ui.Element {
	selected := options.Value != "" && item.Value == options.Value
	option := ui.Button("").
		Attr("type", "button").
		Attr("role", "option").
		Attr("aria-selected", ariaBool(selected)).
		Class(classes("of-combobox-item", "of-transition"))
	if optionID := partID(options.ID, "-option-"+strconv.Itoa(index)); optionID != "" {
		option.ID(optionID)
	}
	label := item.Label
	if label == "" {
		label = item.Value
	}
	option.Children(ui.Span().Class("of-combobox-item-label").Text(label))
	if selected {
		option.Children(ui.Span().Class("of-combobox-check").Attr("aria-hidden", "true").Text("✓"))
	}
	if options.OnSelect != nil {
		value := item.Value
		option.OnClick(func(ctx context.Context) error {
			return options.OnSelect(ctx, value)
		})
	}
	return option
}

func comboboxMatches(item ComboboxItem, query string) bool {
	query = strings.TrimSpace(query)
	if query == "" {
		return true
	}
	label := item.Label
	if label == "" {
		label = item.Value
	}
	return strings.Contains(strings.ToLower(label), strings.ToLower(query))
}

// comboboxLabel resolves the visible label for value. An empty value is the
// "no selection" state, so it never resolves against an item.
func comboboxLabel(items []ComboboxItem, value string) (string, bool) {
	if value == "" {
		return "", false
	}
	for _, item := range items {
		if item.Value != value {
			continue
		}
		if item.Label != "" {
			return item.Label, true
		}
		return item.Value, true
	}
	return "", false
}

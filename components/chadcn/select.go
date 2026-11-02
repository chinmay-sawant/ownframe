package chadcn

import (
	"context"
	"strconv"

	"github.com/chinmay-sawant/ownframe/ui"
)

// SelectItem is one option in a Select list. An empty Label falls back to Value.
type SelectItem struct {
	Value    string
	Label    string
	Disabled bool
}

// SelectOptions describes a controlled single-select. The host owns Open and
// Value, updates them in the callbacks, and rebuilds the view.
//
// ID names the root. The trigger uses ID+"-trigger", the listbox uses
// ID+"-panel", and each option uses ID+"-option-<index>". An empty ID leaves
// the parts unnamed.
type SelectOptions struct {
	ID           string
	Open         bool
	Value        string
	Placeholder  string
	Items        []SelectItem
	OnOpenChange func(ctx context.Context, open bool) error
	OnSelect     func(ctx context.Context, value string) error
}

// Select renders a trigger button and, when Open, its listbox panel. The
// trigger shows the selected label or the placeholder when Value is empty.
func Select(options SelectOptions) *ui.Element {
	label, selected := selectItemLabel(options.Items, options.Value)
	triggerLabel := "Select an option"
	labelClass := "of-select-placeholder"
	if selected {
		triggerLabel = label
		labelClass = "of-select-value"
	} else if options.Placeholder != "" {
		triggerLabel = options.Placeholder
	}
	trigger := ui.Button("").
		Attr("type", "button").
		Attr("role", "combobox").
		Attr("aria-haspopup", "listbox").
		Attr("aria-expanded", ariaBool(options.Open)).
		Class(classes("of-select-trigger", "of-transition")).
		Children(
			ui.Span().Class(classes("of-select-label", labelClass)).Text(triggerLabel),
			ui.Span().Class("of-select-chevron").Attr("aria-hidden", "true").Text("⌄"),
		)
	triggerID := partID(options.ID, "-trigger")
	if triggerID != "" {
		trigger.ID(triggerID)
	}
	panelID := partID(options.ID, "-panel")
	if options.Open && panelID != "" {
		trigger.Attr("aria-controls", panelID)
	}
	if options.OnOpenChange != nil {
		open := options.Open
		trigger.OnClick(func(ctx context.Context) error {
			return options.OnOpenChange(ctx, !open)
		})
	}
	root := ui.Div().Class(classes("of-scope", "of-select")).Children(trigger)
	if options.ID != "" {
		root.ID(options.ID)
	}
	if options.Open {
		root.Children(selectPanel(options, panelID, triggerID))
	}
	return root
}

func selectPanel(options SelectOptions, panelID, triggerID string) *ui.Element {
	panel := ui.Div().Class("of-select-panel").Attr("role", "listbox")
	if panelID != "" {
		panel.ID(panelID)
		if triggerID != "" {
			panel.Attr("aria-labelledby", triggerID)
		}
	}
	for index, item := range options.Items {
		panel.Children(selectOption(options, index, item))
	}
	return panel
}

func selectOption(options SelectOptions, index int, item SelectItem) *ui.Element {
	selected := options.Value != "" && item.Value == options.Value
	option := ui.Button("").
		Attr("type", "button").
		Attr("role", "option").
		Attr("aria-selected", ariaBool(selected)).
		Class(classes("of-select-item", "of-transition"))
	if optionID := partID(options.ID, "-option-"+strconv.Itoa(index)); optionID != "" {
		option.ID(optionID)
	}
	label := item.Label
	if label == "" {
		label = item.Value
	}
	option.Children(ui.Span().Class("of-select-item-label").Text(label))
	if selected {
		option.Children(ui.Span().Class("of-select-check").Attr("aria-hidden", "true").Text("✓"))
	}
	setDisabled(option, item.Disabled)
	if options.OnSelect != nil && !item.Disabled {
		value := item.Value
		option.OnClick(func(ctx context.Context) error {
			return options.OnSelect(ctx, value)
		})
	}
	return option
}

// selectItemLabel resolves the visible label for value. An empty value is the
// "no selection" state, so it never resolves against an item.
func selectItemLabel(items []SelectItem, value string) (string, bool) {
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

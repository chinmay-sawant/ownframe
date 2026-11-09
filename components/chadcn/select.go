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
// ID names the root. The trigger uses ID+"-trigger", the trigger label uses
// ID+"-label", the listbox uses ID+"-panel", each option uses
// ID+"-option-<index>", and each option's selection marker uses
// ID+"-option-<index>-check". An empty ID leaves the parts unnamed.
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
	panelID := partID(options.ID, "-panel")
	label, selected := selectItemLabel(options.Items, options.Value)
	triggerLabel := "Select an option"
	labelClass := "of-select-placeholder"
	if selected {
		triggerLabel = label
		labelClass = "of-select-value"
	} else if options.Placeholder != "" {
		triggerLabel = options.Placeholder
	}
	// The label is the mutable node on a closed value change. Its stable part
	// id lets that change land as one text edit instead of a full render.
	labelNode := ui.Span().Class(classes("of-select-label", labelClass)).Text(triggerLabel)
	if labelID := partID(options.ID, "-label"); labelID != "" {
		labelNode.ID(labelID)
	}
	trigger := ui.Button("").
		Attr("type", "button").
		Attr("role", "combobox").
		Attr("aria-haspopup", "listbox").
		Attr("aria-expanded", ariaBool(options.Open)).
		Class(classes("of-select-trigger", "of-transition")).
		Children(
			labelNode,
			ui.Span().Class("of-select-chevron").Attr("aria-hidden", "true").Text("⌄"),
		)
	triggerID := partID(options.ID, "-trigger")
	if triggerID != "" {
		trigger.ID(triggerID)
	}
	if options.Open && panelID != "" {
		trigger.Attr("aria-controls", panelID)
	}
	if options.OnOpenChange != nil {
		open := options.Open
		trigger.OnClick(func(ctx context.Context) error {
			return options.OnOpenChange(ctx, !open)
		})
	}
	// The open root is a stacking context. The panel keeps its absolute
	// placement, so it overlays later siblings without reflowing the row.
	openClass := ""
	if options.Open {
		openClass = "of-select--open"
	}
	root := ui.Div().Class(classes("of-scope", "of-select", openClass)).Children(trigger)
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
	// The marker always renders so a same-list selection move keeps every
	// option's child count stable and patches instead of rebuilding. An
	// unselected marker is an empty span, so no visible text changes.
	check := ui.Span().Class("of-select-check").Attr("aria-hidden", "true")
	if checkID := partID(options.ID, "-option-"+strconv.Itoa(index)+"-check"); checkID != "" {
		check.ID(checkID)
	}
	if selected {
		check.Text("✓")
	}
	option.Children(check)
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

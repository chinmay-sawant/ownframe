package chadcn

import (
	"context"
	"strconv"

	"github.com/chinmay-sawant/ownframe/ui"
)

// RadioItem is one option in a RadioGroup.
type RadioItem struct {
	// Value is the value reported when the item is selected.
	Value string
	// Label is the visible item text and its accessible name.
	Label string
	// Description is optional secondary text below the label.
	Description string
	// Disabled marks this item as non-interactive.
	Disabled bool
}

// RadioGroupOptions describes a controlled radio group.
type RadioGroupOptions struct {
	// ID is the group's stable host ID. Each item derives its own ID from it.
	ID string
	// Label is the group's accessible name.
	Label string
	// Value is the selected item's value. The host updates it in OnSelect.
	Value string
	// Items lists the options in display order.
	Items []RadioItem
	// Disabled disables every item in the group.
	Disabled bool
	// OnSelect receives the value of the item the host should select.
	OnSelect func(context.Context, string) error
}

// RadioGroup renders one radio button per item, with radiogroup semantics.
// The host owns the selected value and passes it back on the next build.
func RadioGroup(options RadioGroupOptions) *ui.Element {
	group := ui.Div().Class("of-radio-group").Attr("role", "radiogroup")
	if options.ID != "" {
		group.ID(options.ID)
	}
	if options.Label != "" {
		group.Attr("aria-label", options.Label)
	}
	for index, item := range options.Items {
		group.Children(radioButton(options, item, index))
	}
	return group
}

func radioButton(options RadioGroupOptions, item RadioItem, index int) *ui.Element {
	checked := item.Value == options.Value
	disabled := options.Disabled || item.Disabled
	indicator := ui.Span().Class("of-radio-group-indicator").Attr("aria-hidden", "true").
		Children(ui.Span().Class("of-radio-group-dot"))
	content := ui.Span().Class("of-radio-group-content").Children(
		ui.Span().Class("of-radio-group-label").Text(item.Label),
	)
	if item.Description != "" {
		content.Children(ui.Span().Class("of-radio-group-description").Text(item.Description))
	}
	checkedClass := ""
	if checked {
		checkedClass = "of-radio-group-item--checked"
	}
	button := ui.Button("").Attr("type", "button").Class(classes("of-radio-group-item", checkedClass)).
		Attr("role", "radio").Attr("aria-checked", ariaBool(checked)).
		Children(indicator, content)
	if options.ID != "" {
		button.ID(partID(options.ID, "-item-"+strconv.Itoa(index)))
	}
	if !checked && !disabled && options.OnSelect != nil {
		value := item.Value
		button.OnClick(func(ctx context.Context) error { return options.OnSelect(ctx, value) })
	}
	return setDisabled(button, disabled)
}

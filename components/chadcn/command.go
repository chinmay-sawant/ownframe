package chadcn

import (
	"context"

	"github.com/chinmay-sawant/ownframe/ui"
)

// CommandItem is one selectable command row. Shortcut is display text such as
// "⌘P". Icon is an optional leading node; keep it free of IDs and actions.
type CommandItem struct {
	Value    string
	Label    string
	Shortcut string
	Icon     *ui.Element
	Disabled bool
}

// CommandGroup is a labeled set of command rows.
type CommandGroup struct {
	Label string
	Items []CommandItem
}

// CommandOptions describes an inline command palette. The host owns Query,
// Active, and the filtered Groups.
//
// OnSelect receives the chosen value. OnQuery is available for hosts that
// dispatch their own input callbacks: live typing arrives through the host's
// change handler and Page.Form().Value(ID+"-input"), so pass the resulting
// query back as Query. The list carries role="listbox", each row role="option",
// and the active row aria-selected.
type CommandOptions struct {
	ID       string
	Query    string
	Groups   []CommandGroup
	Active   string
	OnQuery  func(ctx context.Context, query string) error
	OnSelect func(ctx context.Context, value string) error
}

// Command builds a command palette with a search field, labeled groups, one
// active row, and a "No results" state when every group is empty. The list is
// a bounded, vertically scrolling listbox that takes focus, so every group
// stays reachable when the palette is taller than its panel. Disabled rows
// keep their place in the list but dispatch no action.
func Command(options CommandOptions) *ui.Element {
	root := ui.Div().Class("of-command")
	if options.ID != "" {
		root.ID(options.ID)
	}

	input := ui.Input().Class("of-command-input").Attr("type", "text").
		Attr("value", options.Query).Attr("aria-label", "Search commands").
		Attr("placeholder", "Type a command or search...")
	if id := partID(options.ID, "-input"); id != "" {
		input.ID(id)
	}
	if id := partID(options.ID, "-list"); id != "" {
		input.Attr("aria-controls", id).Attr("role", "combobox").
			Attr("aria-expanded", "true").Attr("aria-autocomplete", "list")
	}

	// The list owns the bounded scroll region, so it takes focus for keyboard
	// scrolling while the input keeps the combobox role.
	list := ui.Div().Class("of-command-list").Attr("role", "listbox").Attr("aria-label", "Commands").
		Attr("aria-orientation", "vertical").Attr("tabindex", "0")
	if id := partID(options.ID, "-list"); id != "" {
		list.ID(id)
	}

	items := 0
	for _, group := range options.Groups {
		if len(group.Items) == 0 {
			continue
		}
		if items > 0 {
			list.Children(ui.Div().Class("of-command-separator").Attr("role", "presentation"))
		}
		section := ui.Div().Class("of-command-group").Attr("role", "group")
		if group.Label != "" {
			section.Attr("aria-label", group.Label)
			section.Children(ui.Div().Class("of-command-group-label").Text(group.Label))
		}
		for _, item := range group.Items {
			section.Children(commandItem(options, item))
			items++
		}
		list.Children(section)
	}
	if items == 0 {
		list.Children(ui.Div().Class("of-command-empty").Text("No results"))
	}

	return root.Children(
		ui.Div().Class("of-command-search").Children(
			ui.Span().Class("of-command-search-icon").Attr("aria-hidden", "true").Text("⌕"),
			input,
		),
		list,
	)
}

// commandItem builds one option row.
func commandItem(options CommandOptions, item CommandItem) *ui.Element {
	row := ui.Button("").Attr("type", "button").Attr("role", "option").
		Class("of-command-item").Attr("aria-selected", ariaBool(item.Value == options.Active))
	if item.Icon != nil {
		row.Children(ui.Span().Class("of-command-icon").Attr("aria-hidden", "true").Children(item.Icon))
	}
	row.Children(ui.Span().Class("of-command-item-label").Text(item.Label))
	if item.Shortcut != "" {
		row.Children(ui.Span().Class("of-command-shortcut").Text(item.Shortcut))
	}
	if item.Disabled {
		row.Attr("aria-disabled", "true")
		setDisabled(row, true)
		return row
	}
	if options.OnSelect != nil {
		value := item.Value
		row.OnClick(func(ctx context.Context) error {
			return options.OnSelect(ctx, value)
		})
	}
	return row
}

package chadcn

import (
	"context"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe/ui"
)

// SearchItem is one result row in a Search panel. An empty Label falls back
// to Value. Description is an optional muted second line.
type SearchItem struct {
	Value       string
	Label       string
	Description string
}

// SearchOptions describes a controlled search field with a backdrop card. The
// host owns Open and Query, updates them in the callbacks, and rebuilds the
// view.
//
// ID names the root. The field uses ID+"-input", the backdrop uses
// ID+"-backdrop", the card uses ID+"-panel", and each row uses
// ID+"-option-<index>". An empty ID leaves the parts unnamed.
type SearchOptions struct {
	ID           string
	Query        string
	Placeholder  string
	Open         bool
	Items        []SearchItem
	Value        string
	OnOpenChange func(ctx context.Context, open bool) error
	OnQuery      func(ctx context.Context, query string) error
	OnSelect     func(ctx context.Context, value string) error
}

// Search renders a search field that opens a backdrop card on click and
// filters its prefilled rows as the host's Query changes. Live typing arrives
// through the host's change handler and Page.Form().FormValue(ID+"-input"),
// so pass the resulting query back as Query. When no prefilled row matches,
// the host can run a background fetch from its change handler or OnQuery and
// append the remote rows to Items.
func Search(options SearchOptions) *ui.Element {
	placeholder := options.Placeholder
	if placeholder == "" {
		placeholder = "Search..."
	}
	input := ui.Input().
		Attr("type", "search").
		Attr("role", "combobox").
		Attr("aria-autocomplete", "list").
		Attr("aria-expanded", ariaBool(options.Open)).
		Attr("aria-label", placeholder).
		Attr("placeholder", placeholder).
		Attr("value", options.Query).
		Class(classes("of-scope", "of-input", "of-focus", "of-search-field", "of-search-input"))
	if inputID := partID(options.ID, "-input"); inputID != "" {
		input.ID(inputID)
	}
	if options.Open {
		if panelID := partID(options.ID, "-panel"); panelID != "" {
			input.Attr("aria-controls", panelID)
		}
	}
	if options.OnOpenChange != nil && !options.Open {
		open := options.Open
		input.OnClick(func(ctx context.Context) error {
			return options.OnOpenChange(ctx, !open)
		})
	}
	openClass := ""
	if options.Open {
		openClass = "of-search--open"
	}
	root := ui.Div().Class(classes("of-scope", "of-search", openClass))
	if options.ID != "" {
		root.ID(options.ID)
	}
	if !options.Open {
		return root.Children(input)
	}
	// Hit testing keeps the later box, and a fixed backdrop is painted after
	// in-flow content. The field and the panel follow the backdrop, so a
	// click on the field hits the field and a click on a row hits the panel.
	return root.Children(searchBackdrop(options), input, searchPanel(options))
}

func searchBackdrop(options SearchOptions) *ui.Element {
	backdrop := ui.Div().Class("of-search-backdrop")
	if backdropID := partID(options.ID, "-backdrop"); backdropID != "" {
		backdrop.ID(backdropID)
	}
	if options.OnOpenChange == nil {
		return backdrop
	}
	backdrop.OnClick(func(ctx context.Context) error {
		return options.OnOpenChange(ctx, false)
	})
	return backdrop
}

func searchPanel(options SearchOptions) *ui.Element {
	panel := ui.Div().Class("of-search-panel").Attr("role", "listbox").Attr("aria-label", "Search results")
	if panelID := partID(options.ID, "-panel"); panelID != "" {
		panel.ID(panelID)
	}
	matched := 0
	for index, item := range options.Items {
		if !searchMatches(item, options.Query) {
			continue
		}
		panel.Children(searchOption(options, index, item))
		matched++
	}
	if matched == 0 {
		empty := ui.Div().Class("of-search-empty")
		query := strings.TrimSpace(options.Query)
		if query == "" {
			empty.Text("No results")
		} else {
			empty.Text("No results for \"" + query + "\"")
		}
		panel.Children(empty)
	}
	return panel
}

func searchOption(options SearchOptions, index int, item SearchItem) *ui.Element {
	selected := options.Value != "" && item.Value == options.Value
	option := ui.Button("").
		Attr("type", "button").
		Attr("role", "option").
		Attr("aria-selected", ariaBool(selected)).
		Class(classes("of-search-item", "of-transition"))
	if optionID := partID(options.ID, "-option-"+strconv.Itoa(index)); optionID != "" {
		option.ID(optionID)
	}
	label := item.Label
	if label == "" {
		label = item.Value
	}
	option.Children(ui.Span().Class("of-search-item-label").Text(label))
	if item.Description != "" {
		option.Children(ui.Span().Class("of-search-item-description").Text(item.Description))
	}
	if selected {
		option.Children(ui.Span().Class("of-search-check").Attr("aria-hidden", "true").Text("✓"))
	}
	if options.OnSelect != nil {
		value := item.Value
		option.OnClick(func(ctx context.Context) error {
			return options.OnSelect(ctx, value)
		})
	}
	return option
}

func searchMatches(item SearchItem, query string) bool {
	query = strings.TrimSpace(query)
	if query == "" {
		return true
	}
	lower := strings.ToLower(query)
	label := item.Label
	if label == "" {
		label = item.Value
	}
	haystack := strings.ToLower(label + " " + item.Value + " " + item.Description)
	return strings.Contains(haystack, lower)
}

// FilterSearchItems keeps only the rows matching the query,
// case-insensitively on label, value, or description. An empty query keeps
// every row, so typing filters the card and clearing restores it. Hosts that
// fetch remotely can filter the merged list with this helper before passing
// it back as Items.
func FilterSearchItems(items []SearchItem, query string) []SearchItem {
	kept := make([]SearchItem, 0, len(items))
	for _, item := range items {
		if searchMatches(item, query) {
			kept = append(kept, item)
		}
	}
	return kept
}

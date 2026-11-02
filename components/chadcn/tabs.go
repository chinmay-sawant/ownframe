package chadcn

import (
	"context"

	"github.com/chinmay-sawant/ownframe/ui"
)

// TabItem is one tab trigger and the panel it selects.
type TabItem struct {
	Content *ui.Element
	ID      string
	Label   string
}

// TabsOptions describes a controlled tab set. Active is the selected item ID;
// when it is empty the first item is active. OnSelect receives the clicked tab
// ID, and keyboard traversal stays with the host.
type TabsOptions struct {
	OnSelect func(ctx context.Context, id string) error
	Active   string
	Items    []TabItem
}

// Tabs builds a tab list and the active panel. Tab and panel part IDs derive
// from each item ID, and the active tab is the only tab in the focus order.
func Tabs(options TabsOptions) *ui.Element {
	active := options.Active
	list := ui.Div().Class("of-tabs-list").Attr("role", "tablist")
	tabs := ui.Div().Class("of-tabs").Children(list)
	for index, item := range options.Items {
		selected := item.ID == active
		if active == "" {
			selected = index == 0
		}
		tabID, panelID := partID(item.ID, "-tab"), partID(item.ID, "-panel")
		tab := ui.Button("").Attr("type", "button").Class("of-tabs-trigger").
			Attr("role", "tab").Attr("aria-selected", ariaBool(selected))
		if tabID != "" {
			tab.ID(tabID)
		}
		if selected && panelID != "" {
			tab.Attr("aria-controls", panelID)
		}
		if selected {
			tab.Attr("tabindex", "0")
		} else {
			tab.Attr("tabindex", "-1")
		}
		tab.Text(item.Label)
		if options.OnSelect != nil {
			id := item.ID
			tab.OnClick(func(ctx context.Context) error { return options.OnSelect(ctx, id) })
		}
		list.Children(tab)
		if !selected {
			continue
		}
		panel := ui.Div().Class("of-tabs-panel").Attr("role", "tabpanel").Attr("tabindex", "0")
		if panelID != "" {
			panel.ID(panelID)
		}
		if tabID != "" {
			panel.Attr("aria-labelledby", tabID)
		} else if item.Label != "" {
			panel.Attr("aria-label", item.Label)
		}
		tabs.Children(panel.Children(item.Content))
	}
	return tabs
}

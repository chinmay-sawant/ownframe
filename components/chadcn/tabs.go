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
// ID. Interactions adds automatic left/right, Home, and End activation.
//
// Render builds one panel body on demand. When it is set, Tabs calls Render
// only for the active item and ignores TabItem.Content, so a host that changes
// tabs does not build every panel on each refresh. A nil result leaves the
// active panel empty. Callers that do not set Render keep the eager
// TabItem.Content behavior.
type TabsOptions struct {
	Interactions *Interactions
	OnSelect     func(ctx context.Context, id string) error
	Render       func(id string) *ui.Element
	Active       string
	Items        []TabItem
}

// Tabs builds a tab list and the active panel. Tab and panel part IDs derive
// from each item ID, and the active tab is the only tab in the focus order.
func Tabs(options TabsOptions) *ui.Element {
	if options.Interactions != nil {
		options.Interactions.addTabs(options)
	}
	active := activeTabIndex(options)
	list := ui.Div().Class("of-tabs-list").Attr("role", "tablist")
	tabs := ui.Div().Class("of-tabs").Children(list)
	for index, item := range options.Items {
		selected := index == active
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
		content := item.Content
		if options.Render != nil {
			content = options.Render(item.ID)
		}
		empty := ""
		if content == nil {
			empty = "of-tabs-panel--empty"
		}
		panel := ui.Div().Class(classes("of-tabs-panel", empty)).
			Attr("role", "tabpanel").Attr("tabindex", "0")
		if panelID != "" {
			panel.ID(panelID)
		}
		if tabID != "" {
			panel.Attr("aria-labelledby", tabID)
		} else if item.Label != "" {
			panel.Attr("aria-label", item.Label)
		}
		tabs.Children(panel.Children(content))
	}
	return tabs
}

// activeTabIndex resolves which item is selected. An empty Active selects the
// first item; an Active that names no item selects nothing.
func activeTabIndex(options TabsOptions) int {
	if options.Active == "" {
		if len(options.Items) == 0 {
			return -1
		}
		return 0
	}
	for index, item := range options.Items {
		if item.ID == options.Active {
			return index
		}
	}
	return -1
}

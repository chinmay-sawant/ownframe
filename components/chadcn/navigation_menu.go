package chadcn

import (
	"context"

	"github.com/chinmay-sawant/ownframe/ui"
)

// NavigationItem is one entry in a navigation menu. Children render inside the
// open panel as titled rows with optional descriptions.
type NavigationItem struct {
	ID          string
	Label       string
	Description string
	Children    []NavigationItem
}

// NavigationMenuOptions describes the items and the open item of a navigation
// menu. Label names the navigation landmark for assistive technology. The host
// owns Open and passes the chosen item ID back as Open after OnOpenChange.
type NavigationMenuOptions struct {
	ID           string
	Label        string
	Open         string
	Items        []NavigationItem
	OnOpenChange func(ctx context.Context, id string) error
}

// NavigationMenu builds a horizontal row of navigation triggers with one open
// content panel beneath it. Open holds the item ID whose panel is shown.
//
// Each trigger uses ID+"-trigger" and its panel uses ID+"-content". A trigger
// with children reports its ID through OnOpenChange and carries aria-expanded;
// the host decides whether that ID opens the panel or closes the open one. An
// item without children renders as a plain link button and reports its ID
// through the same callback, so the host can treat it as a selection.
func NavigationMenu(options NavigationMenuOptions) *ui.Element {
	root := ui.Nav().Attr("role", "navigation").Class("of-navigation-menu")
	if options.ID != "" {
		root.ID(options.ID)
	}
	if options.Label != "" {
		root.Attr("aria-label", options.Label)
	}

	list := ui.Div().Class("of-navigation-menu-list")
	for _, item := range options.Items {
		hasPanel := len(item.Children) > 0
		open := hasPanel && options.Open != "" && options.Open == item.ID

		triggerClass := "of-navigation-menu-trigger"
		if !hasPanel {
			triggerClass = "of-navigation-menu-link"
		}
		trigger := ui.Button("").Attr("type", "button").Class(triggerClass)
		triggerID, contentID := partID(item.ID, "-trigger"), partID(item.ID, "-content")
		if triggerID != "" {
			trigger.ID(triggerID)
		}
		if hasPanel {
			trigger.Attr("aria-expanded", ariaBool(open))
			if open && contentID != "" {
				trigger.Attr("aria-controls", contentID)
			}
		}
		if options.OnOpenChange != nil {
			id := item.ID
			trigger.OnClick(func(ctx context.Context) error {
				return options.OnOpenChange(ctx, id)
			})
		}
		trigger.Children(ui.Span().Class("of-navigation-menu-label").Text(item.Label))
		if hasPanel {
			trigger.Children(ui.Span().Class("of-navigation-menu-chevron").
				Attr("aria-hidden", "true").Text("▾"))
		}
		list.Children(trigger)
	}
	root.Children(list)

	if item, ok := openNavigationItem(options.Items, options.Open); ok {
		root.Children(navigationContent(item))
	}
	return root
}

// openNavigationItem finds the item whose panel is open. Items without
// children never open a panel.
func openNavigationItem(items []NavigationItem, open string) (NavigationItem, bool) {
	if open == "" {
		return NavigationItem{}, false
	}
	for _, item := range items {
		if item.ID == open && len(item.Children) > 0 {
			return item, true
		}
	}
	return NavigationItem{}, false
}

// navigationContent builds the panel for one open item.
func navigationContent(item NavigationItem) *ui.Element {
	content := ui.Div().Class("of-navigation-menu-content").Attr("role", "region")
	if item.ID != "" {
		content.ID(partID(item.ID, "-content")).Attr("aria-labelledby", partID(item.ID, "-trigger"))
	} else if item.Label != "" {
		content.Attr("aria-label", item.Label)
	}
	entries := ui.Ul().Class("of-navigation-menu-entries")
	for _, child := range item.Children {
		entry := ui.Li().Class("of-navigation-menu-entry").Children(
			ui.Span().Class("of-navigation-menu-entry-title").Text(child.Label),
		)
		if child.Description != "" {
			entry.Children(ui.Span().Class("of-navigation-menu-entry-description").Text(child.Description))
		}
		entries.Children(entry)
	}
	return content.Children(entries)
}

package chadcn

import (
	"github.com/chinmay-sawant/ownframe/ui"
)

// SidebarItem describes one navigation row. Label is the row text, and Active
// marks the current page. Action is dispatched as a data-action. Badge is
// optional trailing text, and Icon is an optional leading element.
type SidebarItem struct {
	Label  string
	Icon   *ui.Element
	Active bool
	Badge  string
	Action string
}

// SidebarOptions describes a vertical navigation surface. Open shows the
// surface; a closed sidebar renders hidden so the host can toggle it by
// rebuilding the view. Collapsed narrows an open sidebar to its icon rail.
// Header, Content, and Footer are optional stacked slots. Width sets the
// expanded rail width; the collapsed rail reads its own token.
type SidebarOptions struct {
	ID        string
	Label     string
	Open      bool
	Collapsed bool
	Header    *ui.Element
	Content   *ui.Element
	Footer    *ui.Element
	Width     ui.Length
}

// Sidebar builds the vertical navigation column. Items come from SidebarNavItem.
// The host owns open and collapsed state and passes it back in.
func Sidebar(options SidebarOptions) *ui.Element {
	state := "closed"
	if options.Open {
		state = "open"
	}
	names := []string{"of-scope", "of-sidebar", "of-sidebar--" + state}
	if options.Collapsed {
		names = append(names, "of-sidebar--collapsed")
	}
	label := options.Label
	if label == "" {
		label = "Sidebar"
	}
	root := ui.Nav().Class(classes(names...)).Attr("data-state", state).Attr("aria-label", label)
	if !options.Open {
		root.Attr("aria-hidden", "true")
	}
	if options.ID != "" {
		root.ID(options.ID)
	}
	if options.Width != (ui.Length{}) && !options.Collapsed {
		root.Width(options.Width)
	}
	if options.Header != nil {
		root.Children(ui.Div().Class("of-sidebar-header").Children(options.Header))
	}
	if options.Content != nil {
		root.Children(ui.Div().Class("of-sidebar-content").Children(options.Content))
	}
	if options.Footer != nil {
		root.Children(ui.Div().Class("of-sidebar-footer").Children(options.Footer))
	}
	return root
}

// SidebarNavItem builds one navigation row. The active row carries
// aria-current="page", and a row with an Action dispatches that action string.
func SidebarNavItem(item SidebarItem) *ui.Element {
	row := ui.Button("").Attr("type", "button").
		Class(classes("of-sidebar-item", sidebarActiveClass(item.Active)))
	if item.Active {
		row.Attr("aria-current", "page")
	}
	if item.Action != "" {
		row.Attr("data-action", item.Action)
	}
	if item.Icon != nil {
		row.Children(ui.Span().Class("of-sidebar-item-icon").Attr("aria-hidden", "true").
			Children(item.Icon))
	}
	row.Children(ui.Span().Class("of-sidebar-item-label").Text(item.Label))
	if item.Badge != "" {
		row.Children(ui.Span().Class("of-sidebar-item-badge").Text(item.Badge))
	}
	return row
}

func sidebarActiveClass(active bool) string {
	if active {
		return "of-sidebar-item--active"
	}
	return ""
}

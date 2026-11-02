package chadcn

import "github.com/chinmay-sawant/ownframe/ui"

// BreadcrumbItem is one trail entry. Action dispatches when a non-current
// entry is clicked; Current marks the page itself.
type BreadcrumbItem struct {
	Action  string
	Label   string
	Current bool
}

// Breadcrumb builds a navigational trail. The last item is always the current
// page, so a host does not have to repeat that flag on the final entry.
func Breadcrumb(items []BreadcrumbItem) *ui.Element {
	list := ui.Tag("ol").Class("of-breadcrumb-list")
	for index, item := range items {
		current := item.Current || index == len(items)-1
		var node *ui.Element
		switch {
		case current:
			node = ui.Span().Class("of-breadcrumb-page").Attr("aria-current", "page").Text(item.Label)
		case item.Action != "":
			node = ui.Button(item.Label).Attr("type", "button").Class("of-breadcrumb-link").
				Attr("data-action", item.Action)
		default:
			node = ui.Span().Class("of-breadcrumb-label").Text(item.Label)
		}
		entry := ui.Li().Class("of-breadcrumb-item").Children(node)
		if !current {
			entry.Children(ui.Span().Class("of-breadcrumb-separator").Attr("aria-hidden", "true").Text("›"))
		}
		list.Children(entry)
	}
	return ui.Nav().Class("of-breadcrumb").Attr("aria-label", "Breadcrumb").Children(list)
}

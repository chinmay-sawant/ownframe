package chadcn

import "github.com/chinmay-sawant/ownframe/ui"

// EmptyOptions describes an empty state. Every slot is optional.
type EmptyOptions struct {
	Icon        *ui.Element
	Title       string
	Description string
	Actions     []*ui.Element
	Footer      *ui.Element
}

// Empty fills a region with a centered icon, title, description, and actions.
// Extra children follow the description inside the text column.
func Empty(options EmptyOptions, children ...*ui.Element) *ui.Element {
	root := ui.Div().Class("of-empty")
	if options.Icon != nil {
		root.Children(ui.Div().Class("of-empty-icon").
			Attr("aria-hidden", ariaBool(true)).Children(options.Icon))
	}

	if options.Title != "" || options.Description != "" || len(children) > 0 {
		content := ui.Div().Class("of-empty-content")
		if options.Title != "" {
			content.Children(ui.Tag("h3").Class("of-empty-title").Text(options.Title))
		}
		if options.Description != "" {
			content.Children(ui.P().Class("of-empty-description").Text(options.Description))
		}
		content.Children(children...)
		root.Children(content)
	}

	if len(options.Actions) > 0 {
		root.Children(ui.Div().Class("of-empty-actions").Children(options.Actions...))
	}
	if options.Footer != nil {
		root.Children(ui.Div().Class("of-empty-footer").Children(options.Footer))
	}
	return root
}

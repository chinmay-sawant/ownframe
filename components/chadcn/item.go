package chadcn

import "github.com/chinmay-sawant/ownframe/ui"

// ItemOptions describes a list row. Media and Actions are optional slots.
// Content replaces the Title and Description convenience fields when set.
type ItemOptions struct {
	Media       *ui.Element
	Content     *ui.Element
	Actions     *ui.Element
	Title       string
	Description string
}

// Item lays out media, text content, and trailing actions on one row. Extra
// children are appended to the content column.
func Item(options ItemOptions, children ...*ui.Element) *ui.Element {
	root := ui.Div().Class("of-item")
	if options.Media != nil {
		root.Children(ui.Div().Class("of-item-media").Children(options.Media))
	}

	content := ui.Div().Class("of-item-content")
	if options.Content != nil {
		content.Children(options.Content)
	} else {
		if options.Title != "" {
			content.Children(ui.Tag("h3").Class("of-item-title").Text(options.Title))
		}
		if options.Description != "" {
			content.Children(ui.P().Class("of-item-description").Text(options.Description))
		}
	}
	content.Children(children...)
	root.Children(content)

	if options.Actions != nil {
		root.Children(ui.Div().Class("of-item-actions").Children(options.Actions))
	}
	return root
}

package chadcn

import "github.com/chinmay-sawant/ownframe/ui"

// H1 is the largest document heading.
func H1(text string) *ui.Element {
	return ui.Tag("h1").Class("of-typography of-typography--h1").Text(text)
}

// H2 is a major section heading.
func H2(text string) *ui.Element {
	return ui.Tag("h2").Class("of-typography of-typography--h2").Text(text)
}

// H3 is a subsection heading.
func H3(text string) *ui.Element {
	return ui.Tag("h3").Class("of-typography of-typography--h3").Text(text)
}

// H4 is a minor heading.
func H4(text string) *ui.Element {
	return ui.Tag("h4").Class("of-typography of-typography--h4").Text(text)
}

// P is a body paragraph.
func P(text string) *ui.Element {
	return ui.P().Class("of-typography of-typography--p").Text(text)
}

// Lead is the larger introductory paragraph.
func Lead(text string) *ui.Element {
	return ui.P().Class("of-typography of-typography--lead").Text(text)
}

// Muted is small secondary text.
func Muted(text string) *ui.Element {
	return ui.P().Class("of-typography of-typography--muted").Text(text)
}

// InlineCode marks a short code span inside prose.
func InlineCode(text string) *ui.Element {
	return ui.Tag("code").Class("of-typography-inline-code").Text(text)
}

// Blockquote marks quoted prose.
func Blockquote(text string) *ui.Element {
	return ui.Tag("blockquote").Class("of-typography-blockquote").Text(text)
}

// List is an unordered list with one item per string.
func List(items ...string) *ui.Element {
	list := ui.Ul().Class("of-typography-list")
	for _, item := range items {
		list.Children(ui.Li().Class("of-typography-list-item").Text(item))
	}
	return list
}

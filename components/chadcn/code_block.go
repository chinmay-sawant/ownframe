package chadcn

import (
	"github.com/chinmay-sawant/ownframe/ui"
)

// CodeBlockOptions describes a code frame. Filename and Language are optional
// labels; Code is rendered as plain text.
type CodeBlockOptions struct {
	Filename string
	Language string
	Code     string
}

// CodeBlock builds a code frame with a filename header and a copy label above a
// preformatted body. The copy label is decorative; a host that wants a live
// action can wrap the block with its own control. Syntax highlighting is not
// part of this component.
func CodeBlock(options CodeBlockOptions) *ui.Element {
	header := ui.Header().Class("of-code-block-header")
	if options.Filename != "" {
		header.Children(ui.Span().Class("of-code-block-filename").Text(options.Filename))
	}
	if options.Language != "" {
		header.Children(ui.Span().Class("of-code-block-language").Text(options.Language))
	}
	header.Children(ui.Span().Class("of-code-block-copy").Attr("aria-hidden", "true").Text("Copy"))

	body := ui.Tag("pre").Class("of-code-block-body").Children(
		ui.Tag("code").Class("of-code-block-code").Text(options.Code),
	)
	return ui.Div().Class("of-code-block").Children(header, body)
}

package chadcn

import (
	"context"

	"github.com/chinmay-sawant/ownframe/ui"
)

// CodeBlockOptions describes a code frame. Filename and Language are optional
// labels; Code is rendered as plain text. Action and OnCopy turn the copy
// chip into a button. OnCopy wins when both are set.
type CodeBlockOptions struct {
	ID       string
	Filename string
	Language string
	Code     string
	Action   string
	OnCopy   func(context.Context) error
}

// CodeBlock builds a code frame with a filename header and a copy label above a
// preformatted body. The copy label stays decorative until the host passes
// Action or OnCopy; that control is then a button whose id is derived from ID.
// The code body stays plain text. Syntax highlighting is not part of this
// component.
func CodeBlock(options CodeBlockOptions) *ui.Element {
	header := ui.Header().Class("of-code-block-header")
	if options.Filename != "" {
		header.Children(ui.Span().Class("of-code-block-filename").Text(options.Filename))
	}
	if options.Language != "" {
		header.Children(ui.Span().Class("of-code-block-language").Text(options.Language))
	}
	header.Children(codeBlockCopy(options))

	body := ui.Tag("pre").Class("of-code-block-body").Children(
		ui.Tag("code").Class("of-code-block-code").Text(options.Code),
	)
	root := ui.Div().Class("of-code-block").Children(header, body)
	if options.ID != "" {
		root.ID(options.ID)
	}
	return root
}

// codeBlockCopy is a decorative label until the host passes Action or OnCopy.
// The live control is a button so the copy hover rule can match its id.
func codeBlockCopy(options CodeBlockOptions) *ui.Element {
	if options.OnCopy == nil && options.Action == "" {
		return ui.Span().Class("of-code-block-copy").Attr("aria-hidden", "true").Text("Copy")
	}
	button := ui.Button("Copy").Attr("type", "button").Class("of-code-block-copy")
	if id := partID(options.ID, "-copy"); id != "" {
		button.ID(id)
	}
	if options.OnCopy != nil {
		return button.OnClick(options.OnCopy)
	}
	return button.Attr("data-action", options.Action)
}

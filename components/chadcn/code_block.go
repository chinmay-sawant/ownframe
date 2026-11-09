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
// The code body stays plain text and takes its own derived id when ID is set.
// Syntax highlighting is not part of this
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
		codeBlockCode(options),
	)
	root := ui.Div().Class("of-code-block").Children(header, body)
	if options.ID != "" {
		root.ID(options.ID)
	}
	return root
}

// codeBlockCode is the plain-text code body. It takes a stable id derived
// from ID so a code-text edit can patch, plus display:block so it owns a
// block box: the engine only patches an element that owns one. Without an
// ID it stays id-free with no inline style.
func codeBlockCode(options CodeBlockOptions) *ui.Element {
	code := ui.Tag("code").Class("of-code-block-code").Text(options.Code)
	if id := partID(options.ID, "-code"); id != "" {
		code.Display(ui.DisplayBlock).ID(id)
	}
	return code
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

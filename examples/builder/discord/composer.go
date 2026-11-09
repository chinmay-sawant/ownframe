// Composer: reply preview plus the message input and send button.
//
// Actions emitted:
//   - msg:send — send the composer draft for the active channel.
//   - msg:reply-clear — drop the pending reply started by msg:reply:<id>.
//
// The draft value comes from a.drafts (mirrored from Page.Form by
// actions.go); the chadcn input stays bound to the "dc-composer" form id so
// keystrokes keep flowing to the page form. Polish paint lives in
// composerCSS (composer_css.go) under the dc-cp- prefix; the older
// dc-composer* / dc-replybar* classes are kept alongside so the base sheet
// in styles.go still applies.
package main

import (
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// dcComposerFormID is the form id the composer input binds to. actions.go
// uses the same id when it mirrors Page.Form into a.drafts.
const dcComposerFormID = "dc-composer"

// replySnippetMax is the rune cap for the quoted text in the reply preview.
const replySnippetMax = 60

// Composer slice capacities: the body holds an optional reply block plus the
// input row, and a reply preview holds the label, an optional snippet, and
// the clear control.
const (
	composerRowsCap = 2
	replyPreviewCap = 3
)

// composer renders the message composer: a reply preview when a.replyTo
// names a message, then the input row with the send button.
func (a *app) composer() *ui.Element {
	draft := ""
	if a.drafts != nil {
		draft = a.drafts[a.channelID]
	}

	rows := make([]*ui.Element, 0, composerRowsCap)
	if a.replyTo != "" {
		author := ""
		snippet := ""
		for _, m := range a.currentMessages() {
			if m.ID != a.replyTo {
				continue
			}
			author = m.Author
			if runes := []rune(m.Text); len(runes) > replySnippetMax {
				snippet = string(runes[:replySnippetMax]) + "…"
			} else {
				snippet = m.Text
			}
			break
		}
		label := "Replying to message"
		if author != "" {
			label = "Replying to " + author
		}
		preview := make([]*ui.Element, 0, replyPreviewCap)
		preview = append(preview, ui.Span().Class("dc-cp-reply-label").Text(label))
		if snippet != "" {
			preview = append(preview, ui.Span().Class("dc-replybar-text dc-cp-reply-snippet").Text(snippet))
		}
		preview = append(preview, ui.Span().Class("dc-iconbtn dc-cp-reply-clear").ID("dc-reply-clear").Attr("data-action", "msg:reply-clear").Attr("title", "Discard reply").Text("✕"))
		rows = append(rows, ui.Div().Class("dc-replybar dc-cp-reply").Children(preview...))
	}

	input := chadcn.InputWithOptions(dcComposerFormID, draft, chadcn.InputOptions{
		Placeholder: "Message #" + a.currentChannel().Name,
	}).Flex(ui.Number(1)).MinWidth(ui.Px(0))
	rows = append(rows, ui.Div().Class("dc-composer-box dc-cp-box").Children(
		ui.Span().Class("dc-composer-plus dc-cp-plus").Text("+"),
		input,
		ui.Span().Class("dc-cp-glyphs").Attr("aria-hidden", "true").Children(
			ui.Span().Class("dc-cp-glyph").Attr("title", "Add emoji").Text("😊"),
			ui.Span().Class("dc-cp-gif").Attr("title", "Send a GIF").Text("GIF"),
		),
		ui.Span().Class("dc-cp-send").Children(
			chadcn.Button("Send", chadcn.ButtonOptions{Size: chadcn.ButtonSmall}).ID("dc-composer-send").Attr("data-action", "msg:send"),
		),
	))

	return ui.Div().Class("dc-composer").Children(rows...)
}

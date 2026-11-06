// Chat column: channel header, transcript, typing bar, and composer.
//
// Voice state lives in the sidebar voice panel and the user bar; joining
// or leaving reports a transient toast instead of a persistent strip, so
// the composer keeps its spacing.
//
// Actions emitted:
//   - members:toggle — show or hide the member list.
//   - msg:react:<id>:<emoji> — toggle one reaction on a message.
//   - msg:reply:<id> — start a reply to a message.
//   - msg:pin:<id> — pin or unpin a message.
//   - msg:menu:<id> — open the message hover menu.
//   - msg:reply-clear — drop the pending reply (see composer.go).
//   - msg:send — send the composer draft (see composer.go).
//
// Host surface (app.go, data.go): dmessage with ID, ChannelID, Author,
// AuthorColor, Time ("Today at 08:02"), Text, ReplyTo, Reactions ([]reaction
// with Emoji, Count, Reacted), and Edited, Pinned, Mention flags; dchannel
// with Name and Topic; a.currentChannel, a.currentMessages.
//
// The transcript uses chadcn.Message for rows and chadcn.Bubble for text,
// wrapped in .dc-msg divs for the Discord hover and mention paint. Rows sit
// in a chadcn.MessageScroller inside the .dc-scroll region; driving a scroll
// offset through MessageScroller State is follow-up work for the scroll
// owner and needs no change here. The header pin icon toggles the first
// pinned message, or renders muted when nothing is pinned.
package main

import (
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// mainCol renders the center chat column: the channel header, the scroll
// message list, the typing bar, and the composer.
func (a *app) mainCol() *ui.Element {
	channel := a.currentChannel()
	msgs := a.currentMessages()

	pin := ui.Span().Class("dc-iconbtn").Attr("title", "Pinned messages").Text("📌")
	for _, m := range msgs {
		if m.Pinned {
			pin.ID("dc-chan-pin").Attr("data-action", "msg:pin:"+m.ID)
			break
		}
	}
	header := ui.Div().Class("dc-chanhead").Children(
		ui.Span().Class("dc-topbar-hash").Text("#"),
		ui.Span().Text(channel.Name),
		ui.P().Class("dc-topic").Text(channel.Topic),
		ui.Div().Class("dc-topbar-icons").Children(
			pin,
			ui.Span().Class("dc-iconbtn").ID("dc-chan-members").Attr("data-action", "members:toggle").Attr("title", "Members").Text("👥"),
		),
	)

	offset := a.chatScroll.Offset()
	viewH := a.chatViewH
	if viewH <= 0 {
		viewH = 552
	}
	lo, hi, _, _ := visibleRange(msgs, offset, viewH)
	rows := make([]*ui.Element, 0, (hi-lo)*2)
	prevDay := ""
	for i := range msgs {
		day := messageDay(msgs[i].Time)
		needDiv := day != "" && day != prevDay
		if needDiv {
			prevDay = day
		}
		if i < lo || i >= hi {
			continue
		}
		if needDiv {
			rows = append(rows, dcDateDivider(day))
		}
		rows = append(rows, a.messageRow(msgs[i]))
	}
	if len(rows) == 0 {
		rows = append(rows, ui.Div().Class("dc-empty").Text("No messages yet. Say hello!"))
	}
	content := ui.Div().ID("dc-scroll-content").Children(rows...)
	scroll := ui.Div().ID("dc-scroll-viewport").Class("dc-scroll").Attr("role", "log").
		Attr("aria-live", "polite").Attr("aria-label", "Messages").
		Position(ui.PositionRelative).Overflow(ui.OverflowHidden).
		Children(content)
	scroll.Children(a.paneScrollbar())

	typing := ui.Div().Class("dc-typing").ID("dc-typing")
	if len(a.typing) > 0 {
		verb := " is typing"
		if len(a.typing) > 1 {
			verb = " are typing"
		}
		typing.Children(
			ui.Span().ID("dc-typing-text").Children(
				ui.Tag("strong").Text(strings.Join(a.typing, ", ")),
				ui.Text(verb),
			),
			ui.Span().Class("dc-typing-dots").Children(
				ui.Span(), ui.Span(), ui.Span(),
			),
		)
	}

	return ui.Div().Class("dc-main").Children(
		header,
		scroll,
		typing,
		a.composer(),
	)
}

// messageRow renders one transcript row: an avatar initial tinted with
// AuthorColor, the name+time meta, the text with @mentions in bold, a reply
// context line when ReplyTo names another message, reaction pills, and a
// hover row with reply, pin, and menu triggers.
func (a *app) messageRow(m dmessage) *ui.Element {
	initial := "?"
	if runes := []rune(strings.TrimSpace(m.Author)); len(runes) > 0 {
		initial = strings.ToUpper(string(runes[0]))
	}
	avatar := ui.Div().Class("dc-msg-avatar").Text(initial)
	if isHexColor(m.AuthorColor) {
		avatar.BackgroundColor(ui.Hex(m.AuthorColor))
	}

	content := []*ui.Element{
		ui.Div().Class("dc-msg-text").Children(renderMentionText(m.Text)...),
	}
	switch {
	case m.Edited && m.Pinned:
		content = append(content, ui.Span().Class("dc-msg-flag").Text("(edited, pinned)"))
	case m.Edited:
		content = append(content, ui.Span().Class("dc-msg-flag").Text("(edited)"))
	case m.Pinned:
		content = append(content, ui.Span().Class("dc-msg-flag").Text("(pinned)"))
	}
	if len(m.Reactions) > 0 {
		pills := make([]*ui.Element, 0, len(m.Reactions))
		for i, reaction := range m.Reactions {
			class := "dc-pill"
			if reaction.Reacted {
				class += " dc-pill--me"
			}
			pills = append(pills, ui.Span().Class(class).
				ID("dc-pill-"+m.ID+"-"+strconv.Itoa(i)).
				Attr("data-action", "msg:react:"+m.ID+":"+reaction.Emoji).
				Attr("title", reaction.Emoji).
				Children(
					ui.Text(reaction.Emoji),
					ui.Span().Class("dc-pill-count").Text(strconv.Itoa(reaction.Count)),
				))
		}
		content = append(content, ui.Div().Class("dc-reacts").Children(pills...))
	}
	content = append(content, ui.Div().Class("dc-msg-hover").Children(
		ui.Span().Class("dc-iconbtn").ID("dc-msg-reply-"+m.ID).Attr("data-action", "msg:reply:"+m.ID).Attr("title", "Reply").Text("↩"),
		ui.Span().Class("dc-iconbtn").ID("dc-msg-pin-"+m.ID).Attr("data-action", "msg:pin:"+m.ID).Attr("title", "Pin").Text("📌"),
		ui.Span().Class("dc-iconbtn").ID("dc-msg-menu-"+m.ID).Attr("data-action", "msg:menu:"+m.ID).Attr("title", "More").Text("⋯"),
	))

	kids := []*ui.Element{
		chadcn.Message(chadcn.MessageOptions{
			Name:     m.Author,
			Time:     m.Time,
			Avatar:   avatar,
			Children: content,
		}),
	}
	if m.ReplyTo != "" {
		context := "Replied message"
		for _, parent := range a.currentMessages() {
			if parent.ID == m.ReplyTo {
				context = "↩ " + parent.Author + ": " + parent.Text
				break
			}
		}
		const maxSnippet = 80
		if runes := []rune(context); len(runes) > maxSnippet {
			context = string(runes[:maxSnippet]) + "…"
		}
		kids = append([]*ui.Element{ui.Div().Class("dc-msg-reply").Text(context)}, kids...)
	}

	rowClass := "dc-msg"
	if m.Mention {
		rowClass += " dc-msg--mention"
	}
	return ui.Div().Class(rowClass).ID("dc-msg-" + m.ID).Children(kids...)
}

// dcDateDivider renders one transcript date separator.
func dcDateDivider(label string) *ui.Element {
	return ui.Div().Class("dc-divider").Text(label)
}

// messageDay recovers the divider label from a "Today at 08:02" timestamp:
// the part before " at ". Times without that shape group under no divider.
func messageDay(stamp string) string {
	if i := strings.Index(stamp, " at "); i > 0 {
		return stamp[:i]
	}
	return ""
}

// renderMentionText splits text so @name tokens render bold in the Discord
// mention paint while all other text stays plain.
func renderMentionText(text string) []*ui.Element {
	parts := make([]*ui.Element, 0, 4)
	rest := text
	for len(rest) > 0 {
		at := strings.Index(rest, "@")
		if at < 0 {
			return append(parts, ui.Text(rest))
		}
		end := at + 1
		for end < len(rest) && isMentionChar(rest[end]) {
			end++
		}
		if end == at+1 {
			parts = append(parts, ui.Text(rest[:at+1]))
			rest = rest[at+1:]
			continue
		}
		if at > 0 {
			parts = append(parts, ui.Text(rest[:at]))
		}
		parts = append(parts, ui.Tag("strong").Class("dc-msg-mention").Text(rest[at:end]))
		rest = rest[end:]
	}
	return parts
}

// isMentionChar reports whether b continues an @mention token.
func isMentionChar(b byte) bool {
	return b == '_' ||
		(b >= 'a' && b <= 'z') ||
		(b >= 'A' && b <= 'Z') ||
		(b >= '0' && b <= '9')
}

// isHexColor reports whether s is a #rgb or #rrggbb color.
func isHexColor(s string) bool {
	if len(s) != 4 && len(s) != 7 || s[0] != '#' {
		return false
	}
	for i := 1; i < len(s); i++ {
		c := s[i]
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F') {
			return false
		}
	}
	return true
}

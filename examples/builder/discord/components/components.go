// Package components holds discord-specific presentational helpers.
//
// Everything here is a pure function of its arguments: no app state, no
// callbacks. Panels in the parent package own state and wire data-actions;
// these helpers only paint Discord chrome (presence, pills, dividers) with
// the .dc-* classes from styles.go. Shared controls (Avatar, Badge, Message,
// Bubble, Sheet, Dialog, Toast) stay in components/chadcn.
package components

import (
	"strconv"

	"github.com/chinmay-sawant/ownframe/ui"
)

// StatusDot renders the presence dot for online/idle/dnd/offline/streaming.
func StatusDot(status string) *ui.Element {
	class := "dc-presence dc-presence--offline"
	switch status {
	case "online":
		class = "dc-presence dc-presence--online"
	case "idle":
		class = "dc-presence dc-presence--idle"
	case "dnd":
		class = "dc-presence dc-presence--dnd"
	case "streaming":
		class = "dc-presence dc-presence--streaming"
	}
	return ui.Span().Class(class).Attr("aria-label", status)
}

// MentionBadge renders the unread-mention count pill. Nil when n <= 0 so
// callers can append it unconditionally.
func MentionBadge(n int) *ui.Element {
	if n <= 0 {
		return nil
	}
	return ui.Span().Class("dc-chan-badge").Text(strconv.Itoa(n))
}

// ReactionPill renders one emoji reaction with its count. The --me modifier
// marks reactions the local user joined.
func ReactionPill(emoji string, count int, reacted bool) *ui.Element {
	class := "dc-pill"
	if reacted {
		class += " dc-pill--me"
	}
	return ui.Span().Class(class).Attr("title", emoji).Children(
		ui.Text(emoji),
		ui.Span().Class("dc-pill-count").Text(strconv.Itoa(count)),
	)
}

// DateDivider renders one transcript date separator.
func DateDivider(label string) *ui.Element {
	return ui.Div().Class("dc-divider").Text(label)
}

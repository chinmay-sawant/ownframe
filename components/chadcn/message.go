package chadcn

import (
	"strings"

	"github.com/chinmay-sawant/ownframe/ui"
)

// MessageAlign places a message in the thread. The zero value is a start-side
// message, and MessageEnd mirrors the row.
type MessageAlign uint8

const (
	// MessageStart is the received side.
	MessageStart MessageAlign = iota
	// MessageEnd is the sent side.
	MessageEnd
)

// messageInitialsLimit is the number of words that contribute an initial.
const messageInitialsLimit = 2

// MessageOptions describes one chat row. Avatar is an optional host element; a
// name without an avatar renders initials. Children are the message content,
// usually Bubbles.
type MessageOptions struct {
	Name     string
	Time     string
	Align    MessageAlign
	Avatar   *ui.Element
	Children []*ui.Element
}

// Message builds one chat row: an avatar, a name and time line, and the
// content column. The host owns the messages and rebuilds the row when it
// changes.
func Message(options MessageOptions) *ui.Element {
	row := ui.Tag("article").Class(classes("of-message", messageAlignClass(options.Align)))
	if options.Name != "" {
		row.Attr("aria-label", options.Name)
	}
	switch {
	case options.Avatar != nil:
		row.Children(ui.Div().Class("of-message-avatar").Children(options.Avatar))
	case options.Name != "":
		avatar := ui.Div().Class("of-message-avatar").Attr("aria-hidden", "true")
		row.Children(avatar.Text(messageInitials(options.Name)))
	}

	body := ui.Div().Class("of-message-body")
	if options.Name != "" || options.Time != "" {
		meta := ui.Div().Class("of-message-meta")
		if options.Name != "" {
			meta.Children(ui.Span().Class("of-message-name").Text(options.Name))
		}
		if options.Time != "" {
			meta.Children(ui.Span().Class("of-message-time").Text(options.Time))
		}
		body.Children(meta)
	}
	body.Children(ui.Div().Class("of-message-content").Children(options.Children...))
	return row.Children(body)
}

func messageAlignClass(align MessageAlign) string {
	if align == MessageEnd {
		return "of-message--end"
	}
	return "of-message--start"
}

// messageInitials keeps at most the first letter of the first two words.
func messageInitials(name string) string {
	fields := strings.Fields(name)
	if len(fields) == 0 {
		return ""
	}
	initials := make([]rune, 0, messageInitialsLimit)
	limit := min(len(fields), messageInitialsLimit)
	for index := 0; index < limit; index++ {
		first := []rune(fields[index])[0]
		initials = append(initials, first)
	}
	return strings.ToUpper(string(initials))
}

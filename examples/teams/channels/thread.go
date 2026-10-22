package channels

import (
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

// toggleExpand opens one post's thread and collapses every other post.
func toggleExpand(d *Data, id string) {
	for i := range d.Posts {
		if d.Posts[i].ID != id {
			continue
		}

		open := !d.Posts[i].Expanded
		for j := range d.Posts {
			d.Posts[j].Expanded = false
		}
		d.Posts[i].Expanded = open

		return
	}
}

// sendReply appends the typed text as an own reply and clears the input.
func sendReply(page *ownframe.Page, d *Data, id string) {
	text := strings.TrimSpace(page.Form().FormValue("channels-reply-" + id))
	if text == "" {
		return
	}

	for i := range d.Posts {
		if d.Posts[i].ID != id {
			continue
		}

		reply := Reply{
			ID:       id + "-own-" + strconv.Itoa(len(d.Posts[i].Replies)+1),
			Author:   "Robert Downey Jr.",
			Initials: "RDJ",
			Color:    "red",
			Time:     "now",
			Text:     text,
			Own:      true,
		}
		d.Posts[i].Replies = append(d.Posts[i].Replies, reply)
		keepReply(d, id, reply)
		page.Form().SetFormValue("channels-reply-"+id, "")

		return
	}
}

// keepReply mirrors an appended reply into the channel's stored post when
// that post does not already share d.Posts' backing array.
func keepReply(d *Data, id string, reply Reply) {
	for i := range d.allPosts[d.ActiveChannel] {
		p := &d.allPosts[d.ActiveChannel][i]
		if p.ID != id {
			continue
		}

		if n := len(p.Replies); n > 0 && p.Replies[n-1].ID == reply.ID {
			return
		}

		p.Replies = append(p.Replies, reply)

		return
	}
}

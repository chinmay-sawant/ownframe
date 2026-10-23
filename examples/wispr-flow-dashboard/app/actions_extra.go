package app

import "github.com/chinmay-sawant/ownframe"

// inviteURL is the shareable link the invite page copies.
const inviteURL = "https://wisprflow.ai/invite/chinmay"

// extraNotes maps a page control to the toast it shows.
var extraNotes = map[string]string{
	"dict-start":   "Style setup opened",
	"dict-play":    "Playing dictation",
	"dict-copy":    "Dictation copied",
	"note-new":     "New note created",
	"note-play":    "Playing note",
	"word-add":     "Word added to your dictionary",
	"snip-add":     "Snippet added",
	"style-save":   "Style preference saved",
	"tf-run":       "Transform applied",
	"scratch-new":  "New scratchpad opened",
	"set-save":     "Settings saved",
	"inv-send":     "Invites sent",
	"free-claim":   "Reward claimed",
	"help-article": "Help article opened",
	"voice-record": "Voice recording is not wired in this demo",
}

// dispatchExtra runs page controls that do not switch pages or tabs.
func (a *App) dispatchExtra(action string) {
	if action == "inv-copy" {
		ownframe.Write(inviteURL)
		a.view.Note = "Invite link copied"

		return
	}

	if note, ok := extraNotes[action]; ok {
		a.view.Note = note
	}
}

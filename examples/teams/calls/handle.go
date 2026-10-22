package calls

import (
	"context"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

// keyText maps a dial key action suffix to the character it types.
var keyText = map[string]string{
	"1": "1", "2": "2", "3": "3", "4": "4", "5": "5", "6": "6",
	"7": "7", "8": "8", "9": "9", "0": "0", "star": "*", "hash": "#",
}

// Handle applies one click action; it reports whether the action was ours.
func Handle(_ context.Context, page *ownframe.Page, d *Data, action string) bool {
	switch {
	case action == "calls-query":
		d.Query = page.Form().FormValue("calls-search")
		d.rebuild()
	case action == "calls-tab-history":
		d.Tab = "history"
	case action == "calls-tab-voicemail":
		d.Tab = "voicemail"
	case strings.HasPrefix(action, "calls-open-"):
		d.open(strings.TrimPrefix(action, "calls-open-"))
	case strings.HasPrefix(action, "calls-key-"):
		if key, ok := keyText[strings.TrimPrefix(action, "calls-key-")]; ok {
			d.Number += key
		}
	case action == "calls-backspace":
		if d.Number != "" {
			d.Number = d.Number[:len(d.Number)-1]
		}
	case action == "calls-clear":
		d.Number = ""
	case action == "calls-call":
		d.call()
	case action == "calls-hangup":
		d.Status = ""
	default:
		return false
	}

	return true
}

// open selects a history or voicemail row by id.
func (d *Data) open(id string) {
	d.Active = id

	for _, c := range d.History {
		if c.ID == id {
			d.StatusName = c.Name

			return
		}
	}

	for _, v := range d.Voicemails {
		if v.ID == id {
			d.StatusName = v.Name

			return
		}
	}
}

// call starts a call to the active name, the dialed number, or Unknown.
func (d *Data) call() {
	d.Status = "calling"

	switch {
	case d.StatusName != "":
	case d.Number != "":
		d.StatusName = d.Number
	default:
		d.StatusName = "Unknown"
	}
}

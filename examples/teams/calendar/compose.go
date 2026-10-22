package calendar

import (
	"sort"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

// create appends the meeting chosen in the modal, keeps the list sorted, and
// closes the modal.
func (d *Data) create(page *ownframe.Page) {
	title := strings.TrimSpace(page.Form().FormValue("cal-title"))
	if title == "" {
		title = "Untitled meeting"
	}

	e := Event{
		ID:       nextID(d.AllEvents()),
		Title:    title,
		Time:     d.startTime(),
		Dur:      d.duration(),
		Color:    "violet",
		Location: "Teams meeting",
		Day:      d.dayColumn(),
		Week:     d.Week,
	}

	d.allEvents = append(d.allEvents, e)
	d.Events = append(d.Events, e)
	d.sortEvents()
	d.Compose = false
	page.Form().SetFormValue("cal-title", "")
}

// selectEvent marks only id as selected.
func (d *Data) selectEvent(id string) {
	d.Selected = id
	for i := range d.Events {
		d.Events[i].Selected = d.Events[i].ID == id
	}
}

// sortEvents orders events by day, then start time.
func (d *Data) sortEvents() {
	sort.SliceStable(d.Events, func(i, j int) bool {
		if d.Events[i].Day != d.Events[j].Day {
			return d.Events[i].Day < d.Events[j].Day
		}

		return minutes(d.Events[i].Time) < minutes(d.Events[j].Time)
	})
}

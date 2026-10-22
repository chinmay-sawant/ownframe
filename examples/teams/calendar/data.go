package calendar

import "time"

// weekStart is the Monday of week 0, September 14, 2026.
var weekStart = time.Date(2026, time.September, 14, 0, 0, 0, 0, time.UTC)

// defaultWeek is the week that contains today, Monday October 5, 2026.
const defaultWeek = 3

// weekCount is how many weeks the calendar can show.
const weekCount = 7

// weekMeta computes the day headers, the "September 14 – September 18"
// label, and the today column for week i. Nothing is stored.
func weekMeta(i int) ([]Day, string, int) {
	mon := weekStart.AddDate(0, 0, weekDays*i)
	fri := mon.AddDate(0, 0, workWeekEnd)
	today := -1

	days := make([]Day, 0, workDays)

	for d := 0; d < workDays; d++ {
		days = append(days, Day{Label: mon.AddDate(0, 0, d).Format("Mon 1/2")})
	}

	if i == defaultWeek {
		days[0].Today = true
		today = 0
	}

	label := mon.Format("January 2") + " – " + fri.Format("January 2")

	return days, label, today
}

// FromDB builds the state for one stored week from every stored event and
// the selected id.
func FromDB(week int, events []Event, selected string) Data {
	i := clamp(week)
	days, label, today := weekMeta(i)

	d := Data{
		Week:       i,
		WeekLabel:  label,
		Today:      today,
		SelDur:     1,
		Selected:   selected,
		Days:       days,
		StartTimes: startTimes,
		Durs:       durations,
		allEvents:  events,
	}

	for _, e := range events {
		if e.Week == i {
			d.Events = append(d.Events, e)
		}
	}

	for j := range d.Events {
		d.Events[j].Selected = d.Events[j].ID == selected
	}

	d.sortEvents()

	return d
}

// AllEvents returns every stored event.
func (d Data) AllEvents() []Event { return d.allEvents }

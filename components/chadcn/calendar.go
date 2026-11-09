package chadcn

import (
	"context"
	"strconv"
	"time"

	"github.com/chinmay-sawant/ownframe/ui"
)

const (
	calendarWeekCount   = 6
	calendarDaysPerWeek = 7
)

// CalendarOptions describes a controlled month grid. Month selects the visible
// month and Selected marks the chosen day. A zero Month uses Selected's month,
// then the current month. A zero Selected renders no selection.
//
// AllowOutsideSelection enables clicking days from adjacent months. When
// clicked, the callback receives that day.
//
// ID names the grid. Each day button uses ID+"-day-<cell>", where <cell> is
// the zero-based position in the fixed six-by-seven grid (0..41), so the ids
// stay stable across selection and month changes and those changes land as
// text and attribute edits instead of a rebuild. The title uses ID+"-title"
// and the grid uses ID+"-grid" for the same reason. An empty ID leaves the
// grid, title, and days unnamed.
type CalendarOptions struct {
	ID                    string
	Month                 time.Time
	Selected              time.Time
	AllowOutsideSelection bool
	OnSelect              func(ctx context.Context, day time.Time) error
	OnMonthChange         func(ctx context.Context, month time.Time) error
}

// Calendar renders a month header, a weekday row, and a fixed six-by-seven day
// grid. Days outside Month stay visible but muted. The host owns Month and
// Selected and passes them back on the next view.
func Calendar(options CalendarOptions) *ui.Element {
	month := calendarMonth(options)
	start := month.AddDate(0, 0, -int(month.Weekday()))

	prev := ui.Button("‹").
		Attr("type", "button").
		Attr("aria-label", "Previous month").
		Class(classes("of-calendar-nav", "of-transition"))
	next := ui.Button("›").
		Attr("type", "button").
		Attr("aria-label", "Next month").
		Class(classes("of-calendar-nav", "of-transition"))
	if options.OnMonthChange != nil {
		prev.OnClick(func(ctx context.Context) error {
			return options.OnMonthChange(ctx, month.AddDate(0, -1, 0))
		})
		next.OnClick(func(ctx context.Context) error {
			return options.OnMonthChange(ctx, month.AddDate(0, 1, 0))
		})
	}
	label := calendarMonthLabel(month)
	title := ui.Div().Class("of-calendar-title").Attr("aria-live", "polite").Text(label)
	if titleID := partID(options.ID, "-title"); titleID != "" {
		title.ID(titleID)
	}
	header := ui.Div().Class("of-calendar-header").Children(
		prev,
		title,
		next,
	)

	grid := ui.Div().Class("of-calendar-grid").Attr("role", "grid").Attr("aria-label", label)
	if gridID := partID(options.ID, "-grid"); gridID != "" {
		grid.ID(gridID)
	}
	grid.Children(calendarWeekdays())
	for week := 0; week < calendarWeekCount; week++ {
		row := ui.Div().Class("of-calendar-week").Attr("role", "row")
		for dayIndex := 0; dayIndex < calendarDaysPerWeek; dayIndex++ {
			cell := week*calendarDaysPerWeek + dayIndex
			day := start.AddDate(0, 0, cell)
			row.Children(calendarDay(cell, day, month, options))
		}
		grid.Children(row)
	}
	root := ui.Div().Class(classes("of-scope", "of-calendar")).Children(header, grid)
	if options.ID != "" {
		root.ID(options.ID)
	}
	return root
}

func calendarWeekdays() *ui.Element {
	short := [calendarDaysPerWeek]string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}
	full := [calendarDaysPerWeek]string{
		"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday",
	}
	row := ui.Div().Class("of-calendar-weekdays").Attr("role", "row")
	for index, label := range short {
		row.Children(ui.Div().
			Class("of-calendar-weekday").
			Attr("role", "columnheader").
			Attr("aria-label", full[index]).
			Text(label))
	}
	return row
}

func calendarDay(cell int, day, month time.Time, options CalendarOptions) *ui.Element {
	selected := calendarSameDay(day, options.Selected)
	outside := day.Year() != month.Year() || day.Month() != month.Month()
	cellClass := ""
	if outside {
		cellClass = "of-calendar-day--outside"
	}
	selectedClass := ""
	if selected {
		selectedClass = "of-calendar-day--selected"
	}
	button := ui.Button(strconv.Itoa(day.Day())).
		Attr("type", "button").
		Attr("role", "gridcell").
		Attr("aria-selected", ariaBool(selected)).
		Attr("aria-label", day.Format("January 2, 2006")).
		Class(classes("of-calendar-day", "of-transition", cellClass, selectedClass))
	// The cell position keeps the id stable across selection and month
	// changes, so those changes land as text and attribute edits instead of
	// a rebuild. The click handler still closes over the cell's date, so the
	// callback receives the day shown at that position.
	if dayID := partID(options.ID, "-day-"+strconv.Itoa(cell)); dayID != "" {
		button.ID(dayID)
	}
	// Days from the adjacent month stay visible, and become selectable
	// when AllowOutsideSelection is enabled.
	if options.OnSelect != nil && (options.AllowOutsideSelection || !outside) {
		chosen := day
		button.OnClick(func(ctx context.Context) error {
			return options.OnSelect(ctx, chosen)
		})
	}
	return button
}

// calendarMonth normalizes Month to the first day of its month. A zero Month
// falls back to Selected's month, then to the current month.
func calendarMonth(options CalendarOptions) time.Time {
	month := options.Month
	if month.IsZero() {
		month = options.Selected
	}
	if month.IsZero() {
		month = time.Now()
	}
	return time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, month.Location())
}

func calendarSameDay(left, right time.Time) bool {
	if right.IsZero() {
		return false
	}
	return left.Year() == right.Year() && left.Month() == right.Month() && left.Day() == right.Day()
}

func calendarMonthLabel(month time.Time) string {
	return month.Format("January 2006")
}

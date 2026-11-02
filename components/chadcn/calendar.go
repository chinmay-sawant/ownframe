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
type CalendarOptions struct {
	Month         time.Time
	Selected      time.Time
	OnSelect      func(ctx context.Context, day time.Time) error
	OnMonthChange func(ctx context.Context, month time.Time) error
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
	header := ui.Div().Class("of-calendar-header").Children(
		prev,
		ui.Div().Class("of-calendar-title").Attr("aria-live", "polite").Text(label),
		next,
	)

	grid := ui.Div().Class("of-calendar-grid").Attr("role", "grid").Attr("aria-label", label)
	grid.Children(calendarWeekdays())
	for week := 0; week < calendarWeekCount; week++ {
		row := ui.Div().Class("of-calendar-week").Attr("role", "row")
		for dayIndex := 0; dayIndex < calendarDaysPerWeek; dayIndex++ {
			day := start.AddDate(0, 0, week*calendarDaysPerWeek+dayIndex)
			row.Children(calendarDay(day, month, options))
		}
		grid.Children(row)
	}
	return ui.Div().Class(classes("of-scope", "of-calendar")).Children(header, grid)
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

func calendarDay(day, month time.Time, options CalendarOptions) *ui.Element {
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
	cell := ui.Button(strconv.Itoa(day.Day())).
		Attr("type", "button").
		Attr("role", "gridcell").
		Attr("aria-selected", ariaBool(selected)).
		Attr("aria-label", day.Format("January 2, 2006")).
		Class(classes("of-calendar-day", "of-transition", cellClass, selectedClass))
	if options.OnSelect != nil {
		chosen := day
		cell.OnClick(func(ctx context.Context) error {
			return options.OnSelect(ctx, chosen)
		})
	}
	return cell
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

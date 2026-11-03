package chadcn

import (
	"context"
	"time"

	"github.com/chinmay-sawant/ownframe/ui"
)

// DatePickerOptions describes a controlled date picker. The host owns Open,
// Month, and Selected, updates them in the callbacks, and rebuilds the view.
//
// ID names the root. The trigger uses ID+"-trigger" and the calendar panel
// uses ID+"-panel". An empty Selected shows the placeholder in the trigger.
type DatePickerOptions struct {
	ID            string
	Open          bool
	Month         time.Time
	Selected      time.Time
	OnOpenChange  func(ctx context.Context, open bool) error
	OnSelect      func(ctx context.Context, day time.Time) error
	OnMonthChange func(ctx context.Context, month time.Time) error
}

// DatePicker renders a trigger that shows the selected date or a placeholder,
// and, when Open, a calendar panel. The picker does not close itself; the
// host closes it from OnSelect when that is the wanted behavior.
func DatePicker(options DatePickerOptions) *ui.Element {
	panelID := partID(options.ID, "-panel")
	label := "Pick a date"
	labelClass := "of-date-picker-placeholder"
	if !options.Selected.IsZero() {
		label = options.Selected.Format("January 2, 2006")
		labelClass = "of-date-picker-value"
	}
	trigger := ui.Button("").
		Attr("type", "button").
		Attr("aria-haspopup", "dialog").
		Attr("aria-expanded", ariaBool(options.Open)).
		Class(classes("of-date-picker-trigger", "of-transition")).
		Children(
			ui.Span().Class(classes("of-date-picker-label", labelClass)).Text(label),
			ui.Span().Class("of-date-picker-chevron").Attr("aria-hidden", "true").Text("⌄"),
		)
	triggerID := partID(options.ID, "-trigger")
	if triggerID != "" {
		trigger.ID(triggerID)
	}
	if options.Open && panelID != "" {
		trigger.Attr("aria-controls", panelID)
	}
	if options.OnOpenChange != nil {
		open := options.Open
		trigger.OnClick(func(ctx context.Context) error {
			return options.OnOpenChange(ctx, !open)
		})
	}
	openClass := ""
	if options.Open {
		openClass = "of-date-picker--open"
	}
	root := ui.Div().Class(classes("of-scope", "of-date-picker", openClass)).Children(trigger)
	if options.ID != "" {
		root.ID(options.ID)
	}
	if options.Open {
		root.Children(datePickerPanel(options, panelID))
	}
	return root
}

func datePickerPanel(options DatePickerOptions, panelID string) *ui.Element {
	panel := ui.Div().
		Class("of-date-picker-panel").
		Attr("role", "dialog").
		Attr("aria-label", "Choose a date").
		Children(Calendar(CalendarOptions{
			Month:         options.Month,
			Selected:      options.Selected,
			OnSelect:      options.OnSelect,
			OnMonthChange: options.OnMonthChange,
		}))
	if panelID != "" {
		panel.ID(panelID)
	}
	return panel
}

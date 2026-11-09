package main

import (
	"context"
	"strconv"
	"time"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

const (
	sampleRateMin  = 10
	sampleRateMax  = 100
	sampleRateStep = 10
)

// logFilters is the filter bar above the log table. Each control writes shared
// app state and rebuilds the view; Apply also commits the text fields from the
// page form and Reset returns the bar to the default 24 hour view.
func (a *app) logFilters() *ui.Element {
	rangeItems := make([]chadcn.SelectItem, 0, len(timeRangeOptions))
	for _, option := range timeRangeOptions {
		rangeItems = append(rangeItems, chadcn.SelectItem{Value: option, Label: option})
	}
	serviceItems := make([]chadcn.ComboboxItem, 0, len(serviceOptions))
	for _, option := range serviceOptions {
		serviceItems = append(serviceItems, chadcn.ComboboxItem{Value: option, Label: option})
	}
	sortItems := make([]chadcn.SelectValue, 0, len(sortOptions))
	for _, option := range sortOptions {
		sortItems = append(sortItems, chadcn.SelectValue{Value: option, Label: option})
	}
	severityItems := make([]chadcn.ToggleItem, 0, len(severityOptions))
	for _, level := range severityOptions {
		severityItems = append(severityItems, chadcn.ToggleItem{
			ID:    "log-severity-" + level,
			Value: level,
			Label: level,
		})
	}

	apply := func(ctx context.Context) error {
		query := a.formValue("log-query")
		if search := a.formValue("log-search"); search != a.query && query == a.query {
			query = search
		}
		sortValue := a.formValue("log-sort")
		if sortValue == "" {
			sortValue = a.sortValue
		}
		a.query = query
		a.sortValue = sortValue
		a.pageIndex = 0
		a.setFormValue("log-query", query)
		a.setFormValue("log-search", query)
		return a.showToast(ctx, "Filters applied", a.timeRange+" · "+a.service+" · "+a.sortValue)
	}

	reset := func(ctx context.Context) error {
		a.timeRange = "24 hours"
		a.service = "All services"
		a.query = ""
		a.sortValue = "Newest first"
		a.severity = []string{"error", "warn"}
		a.liveTail = false
		a.autoRefresh = false
		a.nextDataRefresh = time.Time{}
		a.showContext = false
		a.sampleRate = 100
		a.fromDate = time.Now()
		a.calendarMonth = time.Now()
		a.pageIndex = 0
		a.selectOpen = false
		a.comboboxOpen = false
		a.dateOpen = false
		a.setFormValue("log-query", "")
		a.setFormValue("log-search", "")
		a.setFormValue("log-sort", a.sortValue)
		return a.showToast(ctx, "Filters reset", "Back to the last 24 hours, newest first.")
	}

	return ui.Div().Class("app-stack").Children(
		ui.Div().Class("app-filter-row").Children(
			ui.Div().Class("app-filter").Children(
				chadcn.Label("Time range", chadcn.LabelOptions{For: "log-range-trigger"}),
				chadcn.Select(chadcn.SelectOptions{
					ID:    "log-range",
					Open:  a.selectOpen,
					Value: a.timeRange,
					Items: rangeItems,
					OnOpenChange: func(ctx context.Context, open bool) error {
						a.selectOpen = open
						return a.refresh(ctx)
					},
					OnSelect: func(ctx context.Context, value string) error {
						a.timeRange = value
						a.selectOpen = false
						a.pageIndex = 0
						return a.refresh(ctx)
					},
				}).Width(ui.Px(150)),
			),
			ui.Div().Class("app-filter").Children(
				chadcn.Label("Service", chadcn.LabelOptions{For: "log-service-input"}),
				chadcn.Combobox(chadcn.ComboboxOptions{
					ID:    "log-service",
					Open:  a.comboboxOpen,
					Value: a.service,
					Items: serviceItems,
					OnOpenChange: func(ctx context.Context, open bool) error {
						a.comboboxOpen = open
						return a.refresh(ctx)
					},
					OnSelect: func(ctx context.Context, value string) error {
						a.service = value
						a.comboboxOpen = false
						a.pageIndex = 0
						return a.refresh(ctx)
					},
				}).Width(ui.Px(170)),
			),
			ui.Div().Class("app-filter").Children(
				chadcn.Label("From date", chadcn.LabelOptions{For: "log-date-trigger"}),
				chadcn.DatePicker(chadcn.DatePickerOptions{
					ID:                    "log-date",
					Open:                  a.dateOpen,
					Month:                 a.calendarMonth,
					Selected:              a.fromDate,
					AllowOutsideSelection: true,
					OnOpenChange: func(ctx context.Context, open bool) error {
						a.dateOpen = open
						return a.refresh(ctx)
					},
					OnSelect: func(ctx context.Context, day time.Time) error {
						a.fromDate = day
						a.calendarMonth = day
						a.dateOpen = false
						a.pageIndex = 0
						return a.refresh(ctx)
					},
					OnMonthChange: func(ctx context.Context, month time.Time) error {
						a.calendarMonth = month
						return a.refresh(ctx)
					},
				}).Width(ui.Px(175)),
			),
			ui.Div().Class("app-filter app-filter--wide").Children(
				chadcn.Field(chadcn.FieldOptions{
					Label: "Query",
					For:   "log-query",
					Control: chadcn.InputWithOptions("log-query", a.query, chadcn.InputOptions{
						Placeholder: "level:error service:auth",
					}),
				}),
			),
			ui.Div().Class("app-filter app-filter--wide").Children(
				chadcn.Label("Search", chadcn.LabelOptions{For: "log-search"}),
				chadcn.InputGroup(chadcn.InputGroupOptions{
					ID:          "log-search",
					Label:       "Search logs",
					Value:       a.query,
					Placeholder: "Message, service, or trace",
					Prefix:      ui.Text("⌕"),
					Suffix:      ui.Text(strconv.Itoa(len(a.visibleLogs())) + " rows"),
				}),
			),
		),
		ui.Div().Class("app-filter-row").Children(
			ui.Div().Class("app-filter").Children(
				chadcn.Label("Sort", chadcn.LabelOptions{For: "log-sort"}),
				chadcn.NativeSelect("log-sort", sortItems, a.sortValue).Width(ui.Px(145)),
			),
			chadcn.Separator(chadcn.SeparatorOptions{Orientation: chadcn.SeparatorVertical}).
				Height(ui.Px(28)).
				AlignSelf(ui.AlignSelfCenter),
			ui.Div().Class("app-filter app-filter--severity").Children(
				chadcn.Label("Severity"),
				chadcn.ToggleGroup(chadcn.ToggleGroupOptions{
					Label:    "Severity",
					Multiple: true,
					Size:     chadcn.ButtonSmall,
					Selected: a.severity,
					Items:    severityItems,
					OnSelect: func(ctx context.Context, value string) error {
						next := make([]string, 0, len(a.severity)+1)
						found := false
						for _, level := range a.severity {
							if level == value {
								found = true
								continue
							}
							next = append(next, level)
						}
						if !found {
							next = append(next, value)
						}
						a.severity = next
						a.pageIndex = 0
						return a.refresh(ctx)
					},
				}),
			),
			chadcn.Toggle("Live tail", a.liveTail, chadcn.ToggleOptions{
				Size: chadcn.ButtonSmall,
			}).ID("log-live-tail").
				OnClick(func(ctx context.Context) error {
					a.liveTail = !a.liveTail
					return a.patchToggle(ctx, "log-live-tail", a.liveTail)
				}),
			chadcn.SwitchWithOptions("Auto refresh", a.autoRefresh, chadcn.SwitchOptions{
				ID: "log-auto-refresh",
			}).OnClick(func(ctx context.Context) error {
				a.autoRefresh = !a.autoRefresh
				if a.autoRefresh {
					a.nextDataRefresh = time.Now().Add(time.Duration(dataRefreshSeconds) * time.Second)
				} else {
					a.nextDataRefresh = time.Time{}
				}
				a.configureLiveTick()
				return a.patchSwitch(ctx, "log-auto-refresh", a.autoRefresh)
			}),
			chadcn.CheckboxWithOptions("Show context", a.showContext, chadcn.CheckboxOptions{
				ID: "log-context",
			}).OnClick(func(ctx context.Context) error {
				a.showContext = !a.showContext
				return a.patchCheckbox(ctx, "log-context", a.showContext)
			}),
			ui.Div().Class("app-filter").Width(ui.Px(140)).Children(
				ui.Div().Class("app-row app-row--between").Children(
					chadcn.Label("Sample rate"),
					ui.Text(strconv.Itoa(a.sampleRate)+"%").Class("app-muted").ID("log-sample-rate-text"),
				),
				chadcn.Slider(chadcn.SliderOptions{
					ID:    "log-sample-rate",
					Label: "Sample rate",
					Value: float64(a.sampleRate),
					Min:   sampleRateMin,
					Max:   sampleRateMax,
					Step:  sampleRateStep,
				}).Attr("title", "Click to cycle the sample rate").
					OnClick(func(ctx context.Context) error {
						a.sampleRate += 10
						if a.sampleRate > sampleRateMax {
							a.sampleRate = 10
						}
						return a.patchSlider(ctx, "log-sample-rate", float64(a.sampleRate), sampleRateMin, sampleRateMax, strconv.Itoa(a.sampleRate)+"%")
					}),
			),
			chadcn.ButtonGroup(chadcn.ButtonGroupOptions{Label: "Filter actions"},
				chadcn.Button("Apply", chadcn.ButtonOptions{
					Variant: chadcn.ButtonFilled,
					Size:    chadcn.ButtonSmall,
				}).ID("log-apply").OnClick(apply),
				chadcn.Button("Reset", chadcn.ButtonOptions{
					Variant: chadcn.ButtonOutlined,
					Size:    chadcn.ButtonSmall,
				}).ID("log-reset").OnClick(reset),
			),
		),
	)
}

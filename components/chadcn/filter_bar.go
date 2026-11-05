package chadcn

import "github.com/chinmay-sawant/ownframe/ui"

// FilterBarOption is one pill in a FilterBar. Active marks the selected pill.
// Action dispatches as a data-action string.
type FilterBarOption struct {
	Label  string
	Active bool
	Action string
}

// FilterBarLegend is one static legend dot with a label, such as the Opened
// and Resolved keys above a trend plot.
type FilterBarLegend struct {
	Label string
	Tone  int
}

// FilterBarOptions describes a dashboard filter row: a static legend plus a
// pill group the host controls. It is the tab bar above a plot, without the
// tab panels that Tabs owns.
type FilterBarOptions struct {
	ID      string
	Label   string
	Legend  []FilterBarLegend
	Options []FilterBarOption
}

// FilterBar builds a legend and pill row. Pills use aria-pressed and stay
// host-controlled: the host passes Active back in after a click.
func FilterBar(options FilterBarOptions) *ui.Element {
	label := options.Label
	if label == "" {
		label = "Filters"
	}
	root := ui.Div().Class("of-filter-bar").Attr("role", "group").Attr("aria-label", label)
	if options.ID != "" {
		root.ID(options.ID)
	}
	if len(options.Legend) > 0 {
		legend := ui.Div().Class("of-filter-bar-legend").Attr("aria-hidden", "true")
		for _, item := range options.Legend {
			dot := ui.Span().Class("of-filter-bar-dot").Attr("data-tone", filterToneSlot(item.Tone))
			legend.Children(ui.Span().Class("of-filter-bar-legend-item").Children(dot,
				ui.Span().Class("of-filter-bar-legend-label").Text(item.Label)))
		}
		root.Children(legend)
	}
	group := ui.Div().Class("of-filter-bar-group")
	for _, item := range options.Options {
		pill := ui.Button("").Attr("type", "button").
			Class(classes("of-filter-bar-pill", filterActiveClass(item.Active))).
			Attr("aria-pressed", ariaBool(item.Active)).Text(item.Label)
		if item.Action != "" {
			pill.Attr("data-action", item.Action)
		}
		group.Children(pill)
	}
	return root.Children(group)
}

func filterActiveClass(active bool) string {
	if active {
		return "of-filter-bar-pill--active"
	}
	return ""
}

func filterToneSlot(tone int) string {
	if tone < 1 {
		return "1"
	}
	return chartSeriesSlot(tone - 1)
}

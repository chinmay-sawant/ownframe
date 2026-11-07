package chadcn

import (
	"strconv"

	"github.com/chinmay-sawant/ownframe/ui"
)

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
// tab panels that Tabs owns. When ID is set, each pill id is that value plus
// the pill index and each legend label takes a stable suffixed id, so a
// pill-active change patches instead of taking the full-render path. Added or
// removed pills stay FULL by design.
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
		for index, item := range options.Legend {
			dot := ui.Span().Class("of-filter-bar-dot").Attr("data-tone", filterToneSlot(item.Tone))
			label := ui.Span().Class("of-filter-bar-legend-label").Text(item.Label)
			if id := partID(options.ID, "-legend-"+strconv.Itoa(index)+"-label"); id != "" {
				label.ID(id)
			}
			legend.Children(ui.Span().Class("of-filter-bar-legend-item").Children(dot, label))
		}
		root.Children(legend)
	}
	group := ui.Div().Class("of-filter-bar-group")
	for index, item := range options.Options {
		pill := ui.Button("").Attr("type", "button").
			Class(classes("of-filter-bar-pill", filterActiveClass(item.Active))).
			Attr("aria-pressed", ariaBool(item.Active)).Text(item.Label)
		if id := partID(options.ID, "-pill-"+strconv.Itoa(index)); id != "" {
			pill.ID(id)
		}
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

package echarts

import (
	"math"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe/ui"
)

const (
	extraBlue   = "#2563eb"
	extraGreen  = "#16a34a"
	extraAmber  = "#d97706"
	extraRed    = "#dc2626"
	extraViolet = "#7c3aed"

	extraPaletteSize    = 5
	extraFull           = 100
	extraPictorialMin   = 1
	extraPictorialMax   = 8
	extraRoundTolerance = 0.05
	extraFloatDigits    = 1
	extraFloatBits      = 64
	extraDepthStep      = 4
	extraEmptyText      = "No data"
	extraMark           = "▮"
)

// extraColor is the palette for this chart group. The bar group keeps its own helper.
func extraColor(index int) string {
	colors := [extraPaletteSize]string{
		extraBlue,
		extraGreen,
		extraAmber,
		extraRed,
		extraViolet,
	}
	if index < 0 {
		index = 0
	}
	return colors[index%extraPaletteSize]
}

func extraPaint(color string, index int) string {
	cleaned := strings.TrimSpace(color)
	if extraSafeColor(cleaned) {
		return cleaned
	}
	return extraColor(index)
}

func extraSafeColor(color string) bool {
	if color == "" {
		return false
	}
	for _, r := range color {
		if !extraColorRune(r) {
			return false
		}
	}
	return true
}

func extraColorRune(r rune) bool {
	letter := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
	if letter {
		return true
	}
	switch r {
	case '#', '(', ')', ',', '.', '%', ' ', '+', '-', '/':
		return true
	default:
		return false
	}
}

func extraFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func extraShare(value, total float64) float64 {
	if !extraFinite(value) || !extraFinite(total) || total <= 0 || value <= 0 {
		return 0
	}
	share := value / total * extraFull
	if share > extraFull {
		return extraFull
	}
	return share
}

// extraPart is value's share of total. A non-positive total splits the row evenly.
func extraPart(value, total float64, count int) float64 {
	if count <= 0 {
		return 0
	}
	if total <= 0 {
		return extraFull / float64(count)
	}
	return extraShare(value, total)
}

func extraText(value float64) string {
	if !extraFinite(value) {
		return "0"
	}
	rounded := math.Round(value)
	if math.Abs(value-rounded) < extraRoundTolerance {
		return strconv.FormatFloat(rounded, 'f', 0, extraFloatBits)
	}
	return strconv.FormatFloat(value, 'f', extraFloatDigits, extraFloatBits)
}

func extraAria(kind, title, description string) string {
	label := kind
	if title != "" {
		label += ": " + title
	}
	if description != "" {
		label += ". " + description
	}
	return label
}

func extraRoot(chartClass, kind, title, description string, body *ui.Element) *ui.Element {
	root := ui.Section().
		Class("of-ech "+chartClass).
		Attr("role", "group").
		Attr("aria-label", extraAria(kind, title, description))
	if header := extraHeader(title, description); header != nil {
		root.Children(header)
	}
	if body == nil {
		return root.Children(ui.P().Class("of-ech-empty").Text(extraEmptyText))
	}
	return root.Children(body)
}

func extraHeader(title, description string) *ui.Element {
	if title == "" && description == "" {
		return nil
	}
	header := ui.Header().Class("of-ech-header")
	if title != "" {
		header.Children(ui.Tag("h2").Class("of-ech-title").Text(title))
	}
	if description != "" {
		header.Children(ui.P().Class("of-ech-description").Text(description))
	}
	return header
}

func extraBody(plot *ui.Element, series []Series) *ui.Element {
	if plot == nil {
		return nil
	}
	body := ui.Div().Class("of-ech-body").Children(plot)
	if legend := extraLegend(series); legend != nil {
		body.Children(legend)
	}
	return body
}

func extraLegend(series []Series) *ui.Element {
	legend := ui.Div().Class("of-ech-legend")
	shown := 0
	for index, item := range series {
		if item.Name == "" {
			continue
		}
		shown++
		legend.Children(ui.Span().Class("of-ech-legend-item").Children(
			extraSwatch(item.Color, index),
			ui.Span().Class("of-ech-legend-label").Text(item.Name),
		))
	}
	if shown == 0 {
		return nil
	}
	return legend
}

func extraSwatch(color string, index int) *ui.Element {
	return ui.Span().Class("of-ech-swatch").Attr("aria-hidden", "true").
		CustomProperty("--of-ech-color", extraPaint(color, index))
}

func extraColumnCount(series []Series) int {
	count := 0
	for _, item := range series {
		if len(item.Values) > count {
			count = len(item.Values)
		}
	}
	return count
}

func extraSeriesMax(series []Series) float64 {
	maximum := 0.0
	for _, item := range series {
		for _, value := range item.Values {
			if extraFinite(value) && value > maximum {
				maximum = value
			}
		}
	}
	return maximum
}

func extraLabelAt(labels []string, index int) string {
	if index < 0 || index >= len(labels) {
		return ""
	}
	return labels[index]
}

// extraDepthPad is the indent for a sibling row. Rows are not nested, so the
// percent is the absolute depth and does not compound.
func extraDepthPad(depth int) float64 {
	if depth <= 0 {
		return 0
	}
	pad := float64(depth) * extraDepthStep
	if pad > extraFull {
		return extraFull
	}
	return pad
}

func extraNodeWeight(node Node) float64 {
	if extraFinite(node.Value) && node.Value > 0 {
		return node.Value
	}
	return extraLevelTotal(node.Children)
}

func extraLevelTotal(nodes []Node) float64 {
	total := 0.0
	for _, node := range nodes {
		weight := extraNodeWeight(node)
		if weight > 0 {
			total += weight
		}
	}
	return total
}

func extraLinkRows(rowClass string, links []Link) *ui.Element {
	if len(links) == 0 {
		return nil
	}
	list := ui.Div().Class("of-ech-links")
	for index, link := range links {
		list.Children(ui.Div().Class(rowClass).Children(
			extraSwatch("", index),
			ui.Span().Class("of-ech-link-source").Text(link.Source),
			ui.Span().Class("of-ech-link-target").Text(link.Target),
			ui.Span().Class("of-ech-link-value").Text(extraText(link.Value)),
		))
	}
	return list
}

func extraLinkMax(links []Link) float64 {
	maximum := 0.0
	for _, link := range links {
		if extraFinite(link.Value) && link.Value > maximum {
			maximum = link.Value
		}
	}
	return maximum
}

func extraFlowBars(rowClass, barClass string, links []Link) *ui.Element {
	if len(links) == 0 {
		return nil
	}
	maximum := extraLinkMax(links)
	list := ui.Div().Class("of-ech-links")
	for index, link := range links {
		bar := ui.Div().Class(barClass).
			Width(ui.Percent(extraShare(link.Value, maximum))).
			CustomProperty("--of-ech-color", extraColor(index))
		list.Children(ui.Div().Class(rowClass).Children(
			ui.Span().Class("of-ech-link-source").Text(link.Source),
			ui.Span().Class("of-ech-link-target").Text(link.Target),
			ui.Div().Class("of-ech-link-track").Children(bar),
			ui.Span().Class("of-ech-link-value").Text(extraText(link.Value)),
		))
	}
	return list
}

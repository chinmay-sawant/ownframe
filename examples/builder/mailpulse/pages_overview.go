//nolint:mnd // Sample overview fixtures are literal display content: KPI tiles and segment shares are fixed demo values.
package main

import (
	"context"
	"math"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// overviewPage renders the Overview tab: list health, growth and send
// performance at a glance. It mirrors the engagement tab structure so the
// shell can swap it in without layout changes.
func (a *app) overviewPage() *ui.Element {
	return ui.Div().Class("mp-card mp-ov").Children(
		a.overviewHead(),
		a.overviewKPIRow(),
		ui.Div().Class("mp-grid-mid").Children(
			a.overviewGrowthCard(),
			a.overviewSegmentsCard(),
		),
		ui.Div().Class("mp-grid-bot").Children(
			a.overviewSendsCard(),
			a.overviewActionsCard(),
		),
	)
}

// overviewHead renders the Overview header. It is separate from pageHead so
// the engagement tab keeps its own copy.
func (a *app) overviewHead() *ui.Element {
	exportBtn := chadcn.Button("↓ Export", chadcn.ButtonOptions{
		Variant: chadcn.ButtonOutlined,
		Size:    chadcn.ButtonSmall,
	}).Attr("data-action", "export")
	rangeBtn := chadcn.Button(a.rangeLabel+" ▾", chadcn.ButtonOptions{
		Variant: chadcn.ButtonOutlined,
		Size:    chadcn.ButtonSmall,
	}).Attr("data-action", "range:cycle")
	left := ui.Div().Class("mp-pagehead-text").Children(
		chadcn.H1("Overview"),
		chadcn.Muted("List health, growth and send performance at a glance."),
	)
	actions := ui.Div().Class("mp-pagehead-actions").Children(
		rangeBtn,
		exportBtn,
	)
	return ui.Div().Class("mp-pagehead").Children(left, actions)
}

type overviewStat struct {
	label    string
	value    string
	delta    string
	footnote string
	tint     string
	glyph    string
}

// overviewStatsFor returns the Overview KPI tiles for a range window.
// Unknown day counts fall back to the default 30-day window.
func overviewStatsFor(days int) []overviewStat {
	switch days {
	case 7:
		return []overviewStat{
			{label: "Total subscribers", value: "18,442", delta: "+2.1%", footnote: "+96 this week", tint: "purple", glyph: "◉"},
			{label: "Avg. open rate", value: "48.9%", delta: "+2.4%", footnote: "Unique opens ÷ delivered", tint: "orange", glyph: "✉"},
			{label: "Avg. click rate", value: "19.1%", delta: "+1.2%", footnote: "Unique clicks ÷ delivered", tint: "blue", glyph: "➤"},
			{label: "Revenue/Mo", value: "$2,940", delta: "+6.4%", footnote: "+$178 this week", tint: "green", glyph: "↩"},
		}
	case 90:
		return []overviewStat{
			{label: "Total subscribers", value: "18,442", delta: "+7.1%", footnote: "+1,204 this quarter", tint: "purple", glyph: "◉"},
			{label: "Avg. open rate", value: "46.8%", delta: "+2.8%", footnote: "Unique opens ÷ delivered", tint: "orange", glyph: "✉"},
			{label: "Avg. click rate", value: "17.7%", delta: "+1.5%", footnote: "Unique clicks ÷ delivered", tint: "blue", glyph: "➤"},
			{label: "Revenue/Mo", value: "$36,900", delta: "+9.6%", footnote: "+$3,240 this quarter", tint: "green", glyph: "↩"},
		}
	default:
		return []overviewStat{
			{label: "Total subscribers", value: "18,442", delta: "+2.4%", footnote: "+428 this month", tint: "purple", glyph: "◉"},
			{label: "Avg. open rate", value: "47.8%", delta: "+3.1%", footnote: "Unique opens ÷ delivered", tint: "orange", glyph: "✉"},
			{label: "Avg. click rate", value: "18.6%", delta: "+1.9%", footnote: "Unique clicks ÷ delivered", tint: "blue", glyph: "➤"},
			{label: "Revenue/Mo", value: "$12,480", delta: "+8.2%", footnote: "Paid conversions this month", tint: "green", glyph: "↩"},
		}
	}
}

// overviewKPICard renders one Overview KPI tile inline with ui only,
// following the components.KPICard pattern without importing it.
func overviewKPICard(stat overviewStat) *ui.Element {
	top := ui.Div().Class("mp-kpi-top").Children(
		ui.Span().Class("mp-kpi-dot mp-kpi-dot--"+stat.tint).Text(stat.glyph),
		ui.Span().Class("mp-kpi-label").Text(stat.label),
		ui.Span().Class("mp-kpi-more").Text("..."),
	)
	middle := ui.Div().Class("mp-kpi-value-row").Children(
		ui.Span().Class("mp-kpi-value").Text(stat.value),
		ui.Span().Class("mp-delta mp-delta--up").Text("↗ "+stat.delta),
	)
	foot := ui.Div().Class("mp-kpi-foot").Children(
		ui.Span().Class("mp-kpi-leader"),
		ui.Span().Class("mp-muted").Text(stat.footnote),
		ui.Span().Class("mp-kpi-leader"),
	)
	return ui.Div().Class("mp-kpi").Children(top, middle, foot)
}

func (a *app) overviewKPIRow() *ui.Element {
	stats := overviewStatsFor(a.rangeDays())
	cards := make([]*ui.Element, 0, len(stats))
	for _, stat := range stats {
		cards = append(cards, overviewKPICard(stat))
	}
	return ui.Div().Class("mp-kpis").Children(cards...)
}

const (
	// overviewBarTrackHeight is the 0-90 percent track the tallest List
	// growth bar fills; every other bar scales against it.
	overviewBarTrackHeight = 90
	// overviewWeekWindowDays selects the 7-day List growth and send windows.
	overviewWeekWindowDays = 7
	// overviewDefaultWindowDays is the default 30-day window the range
	// helpers fall back to.
	overviewDefaultWindowDays = 30
)

// overviewBarHeights returns the hand-tuned 30-day List growth track: bar
// fill heights as percentages of the track. It returns a fresh slice per
// call so callers never share backing state.
func overviewBarHeights() []float64 {
	return []float64{34, 48, 42, 55, 61, 58, 90, 66, 72, 64, 78, 84}
}

// overviewBarValues holds the net new subscribers per week behind each bar.
// The 12 values sum to the +428 total in the card footnote. It returns a
// fresh slice per call so callers never share backing state.
func overviewBarValues() []int {
	return []int{20, 27, 24, 31, 35, 33, 51, 38, 41, 36, 44, 48}
}

// overviewDailyValues holds net new subscribers per day for the 7-day
// window. The 7 values sum to the +96 weekly total in the KPI footnote. It
// returns a fresh slice per call so callers never share backing state.
func overviewDailyValues() []int {
	return []int{11, 14, 12, 16, 15, 13, 15}
}

// overviewQuarterValues holds net new subscribers per week for the 90-day
// window. The 12 values sum to the +1,204 quarterly total. It returns a
// fresh slice per call so callers never share backing state.
func overviewQuarterValues() []int {
	return []int{56, 76, 68, 87, 98, 93, 143, 107, 115, 101, 124, 136}
}

// barValues returns the List growth values for a range window: 7 daily
// bars for 7 days, 12 weekly bars otherwise. Unknown day counts fall back
// to the default 30-day window.
func barValues(days int) []int {
	switch days {
	case 7:
		return overviewDailyValues()
	case 90:
		return overviewQuarterValues()
	default:
		return overviewBarValues()
	}
}

// barHeightsFor scales the window's values onto a 0-90 percent track so
// the tallest bar fills the track in every window. The 30-day window keeps
// its hand-tuned heights; other windows scale from their peak.
func barHeightsFor(days int) []float64 {
	if days != 7 && days != 90 {
		return overviewBarHeights()
	}
	values := barValues(days)
	peak := 1
	for _, v := range values {
		if v > peak {
			peak = v
		}
	}
	heights := make([]float64, len(values))
	for i, v := range values {
		heights[i] = math.Round(float64(v) / float64(peak) * overviewBarTrackHeight)
	}
	return heights
}

// scaleStr formats a 0-90 track height as a scaleY factor.
func scaleStr(height float64) string {
	return strconv.FormatFloat(height/100, 'f', 3, 64)
}

// growthBarID names one List growth bar so the hover handler can map it
// back to its week index.
func growthBarID(i int) string {
	return "mp-growth-" + strconv.Itoa(i)
}

// parseGrowthBarID recovers the week index from a List growth bar id.
func parseGrowthBarID(id string) (int, bool) {
	rest, ok := strings.CutPrefix(id, "mp-growth-")
	if !ok {
		return 0, false
	}
	i, err := strconv.Atoi(rest)
	if err != nil || i < 0 || i >= len(overviewBarHeights()) {
		return 0, false
	}
	return i, true
}

// growthBarLabel builds the readout line for one bar in the default 30-day
// window, e.g. "W7 · +51 net new".
func growthBarLabel(i int) string {
	heights := overviewBarHeights()
	values := overviewBarValues()
	if i < 0 || i >= len(heights) || i >= len(values) {
		return ""
	}
	return "W" + strconv.Itoa(i+1) + " · +" + strconv.Itoa(values[i]) + " net new"
}

// growthBarLabelFor builds the readout line for the active range window:
// "D4 · +16 net new" for the 7-day window, "W7 · +51 net new" otherwise.
func (a *app) growthBarLabelFor(i int) string {
	values := barValues(a.rangeDays())
	if i < 0 || i >= len(values) {
		return ""
	}
	prefix := "W"
	if a.rangeDays() == overviewWeekWindowDays {
		prefix = "D"
	}
	return prefix + strconv.Itoa(i+1) + " · +" + strconv.Itoa(values[i]) + " net new"
}

func growthReadoutDefault() string {
	return "+428 net new subscribers over the last 12 weeks."
}

// growthReadoutDefaultFor returns the growth total line for the active
// range window. Bar values in each window sum to its total.
func (a *app) growthReadoutDefaultFor() string {
	switch a.rangeDays() {
	case 7:
		return "+96 net new subscribers over the last 7 days."
	case 90:
		return "+1,204 net new subscribers over the last 90 days."
	default:
		return growthReadoutDefault()
	}
}

func (a *app) growthReadout() string {
	if i := a.growthHover; i >= 0 && i < len(barValues(a.rangeDays())) {
		if label := a.growthBarLabelFor(i); label != "" {
			return label
		}
	}
	return a.growthReadoutDefaultFor()
}

func (a *app) overviewGrowthCard() *ui.Element {
	values := barValues(a.rangeDays())
	heights := barHeightsFor(a.rangeDays())
	peakIdx := 0
	for i, v := range values {
		if v > values[peakIdx] {
			peakIdx = i
		}
	}
	desc := "Net new subscribers per week, " + a.rangeWindow() + "."
	if a.rangeDays() == overviewWeekWindowDays {
		desc = "Net new subscribers per day, last 7 days."
	}
	bars := make([]*ui.Element, 0, len(values))
	for i, height := range heights {
		color := "#c9bcfb"
		if i == peakIdx {
			color = "#7c3aed"
		}
		// The fill spans the full track and scales from the baseline, so
		// a range switch moves on the transform path (one composited
		// matrix per bar) instead of the height path (one relayout per
		// bar per frame). At rest scaleY(h/100) of a full-height fill
		// paints the same pixels as a height:h% bar.
		scale := height
		if a.enterBars {
			scale = 0
		}
		class := "mp-bar"
		if a.enterSlow {
			class += " mp-bar-slow"
		}
		bars = append(bars, ui.Div().ID(growthBarID(i)).Class(class).
			Height(ui.Percent(100)).
			Transform(ui.RawValue("scaleY("+scaleStr(scale)+")")).
			Flex(ui.Number(1)).
			MinWidth(ui.Px(0)).
			MaxWidth(ui.Px(40)).
			BackgroundColor(ui.Hex(color)).
			BorderRadius(ui.Px(4)).
			Attr("title", a.growthBarLabelFor(i)))
	}
	return chadcn.Card(
		chadcn.CardHeader(
			chadcn.CardTitle("List growth"),
			chadcn.CardDescription(desc),
		),
		chadcn.CardContent(
			ui.Div().Class("mp-bars").
				Display(ui.DisplayFlex).
				AlignItems(ui.AlignItemsEnd).
				Gap(ui.Px(8)).
				MinHeight(ui.Px(200)).
				FlexGrow(ui.Number(1)).
				Width(ui.Percent(100)).
				Children(bars...),
			ui.Div().ID("mp-growth-readout").Class("mp-muted").Text(a.growthReadout()),
		).Display(ui.DisplayFlex).
			FlexDirection(ui.FlexDirectionColumn).
			Flex(ui.Number(1)).
			Gap(ui.Px(12)),
	).Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn)
}

type overviewSegment struct {
	label  string
	share  string
	filled int
}

// overviewSegmentsFor returns the segment shares for a range window.
// Unknown day counts fall back to the default 30-day window.
func overviewSegmentsFor(days int) []overviewSegment {
	switch days {
	case 7:
		return []overviewSegment{
			{label: "Power readers", share: "34%", filled: 8},
			{label: "New subscribers", share: "26%", filled: 7},
			{label: "Paid members", share: "14%", filled: 6},
			{label: "Dormant win-back", share: "15%", filled: 5},
			{label: "Referral friends", share: "11%", filled: 4},
		}
	case 90:
		return []overviewSegment{
			{label: "Power readers", share: "30%", filled: 8},
			{label: "New subscribers", share: "28%", filled: 7},
			{label: "Paid members", share: "16%", filled: 6},
			{label: "Dormant win-back", share: "15%", filled: 5},
			{label: "Referral friends", share: "11%", filled: 4},
		}
	default:
		return []overviewSegment{
			{label: "Power readers", share: "32%", filled: 8},
			{label: "New subscribers", share: "24%", filled: 7},
			{label: "Paid members", share: "15%", filled: 6},
			{label: "Dormant win-back", share: "18%", filled: 5},
			{label: "Referral friends", share: "11%", filled: 4},
		}
	}
}

// overviewSegmentDotTotal sizes the dot share bars on the Top segments card.
const overviewSegmentDotTotal = 10

func overviewSegmentRow(segment overviewSegment) *ui.Element {
	dots := make([]*ui.Element, 0, overviewSegmentDotTotal)
	for i := 0; i < overviewSegmentDotTotal; i++ {
		class := "mp-dot"
		if i < segment.filled {
			class += " mp-dot--on"
		}
		dots = append(dots, ui.Span().Class(class))
	}
	return ui.Div().Class("mp-linkrow mp-enter").Children(
		ui.Div().Class("mp-link-meta").Children(
			ui.Span().Class("mp-engage-label").Text(segment.label),
			ui.Div().Class("mp-dots").Children(dots...),
		),
		ui.Div().Class("mp-engage-stats").Children(
			ui.Span().Class("mp-engage-open").Text(segment.share).FontWeight(ui.FontWeightBold),
		),
	)
}

func (a *app) overviewSegmentsCard() *ui.Element {
	segments := overviewSegmentsFor(a.rangeDays())
	rows := make([]*ui.Element, 0, len(segments))
	for _, segment := range segments {
		rows = append(rows, overviewSegmentRow(segment))
	}
	return chadcn.Card(
		chadcn.CardHeader(
			chadcn.CardTitle("Top segments"),
			chadcn.CardDescription("Share of engaged readers by segment, "+a.rangeWindow()+"."),
		),
		chadcn.CardContent(rows...),
	)
}

type overviewSend struct {
	id    string
	name  string
	date  string
	open  string
	click string
}

// overviewSendsFor returns the Recent sends rows for a range window: the
// last 2 sends for 7 days, the last 4 for 30 days, and those 4 plus 2
// extra historical rows for 90 days. Unknown day counts fall back to the
// default 30-day window.
func overviewSendsFor(days int) []overviewSend {
	switch days {
	case 7:
		return overviewSendsFor(overviewDefaultWindowDays)[:2]
	case 90:
		return append(overviewSendsFor(overviewDefaultWindowDays),
			overviewSend{id: "send-5", name: "Weekly Creator Brief #51", date: "Feb 12", open: "45.3% opens", click: "16.1% clicks"},
			overviewSend{id: "send-6", name: "Sponsor Spotlight: Acme", date: "Feb 5", open: "37.2% opens", click: "8.4% clicks"},
		)
	default:
		return []overviewSend{
			{id: "send-1", name: "Weekly Creator Brief #52", date: "Mar 12", open: "52% opens", click: "21% clicks"},
			{id: "send-2", name: "AI Tools Worth Trying", date: "Mar 5", open: "49% opens", click: "19% clicks"},
			{id: "send-3", name: "The Creator Economy Report", date: "Feb 26", open: "47% opens", click: "17% clicks"},
			{id: "send-4", name: "Sponsor Spotlight: Notion", date: "Feb 19", open: "38% opens", click: "9% clicks"},
		}
	}
}

// recentSendsDesc names the send count behind the Recent sends card for
// the active range window.
func (a *app) recentSendsDesc() string {
	switch a.rangeDays() {
	case 7:
		return "Last 2 campaigns and their results."
	case 90:
		return "Last 6 campaigns and their results."
	default:
		return "Last 4 campaigns and their results."
	}
}

func (a *app) overviewSendsCard() *ui.Element {
	sends := overviewSendsFor(a.rangeDays())
	rows := make([]*ui.Element, 0, len(sends))
	for _, send := range sends {
		rows = append(rows, ui.Div().Class("mp-linkrow mp-selectable mp-enter").
			Attr("data-action", "select:send:"+send.id+":"+send.name).
			Children(
				ui.Div().Class("mp-linkmain").Children(
					ui.Text(send.name).FontWeight(ui.FontWeightBold).FontSize(ui.Px(13)),
					ui.Text(send.date).FontSize(ui.Px(12)).Color(ui.Hex("#64748b")),
				),
				ui.Div().Class("mp-linkstats").Children(
					ui.Text(send.open).FontWeight(ui.FontWeightBold),
					ui.Text(send.click).Color(ui.Hex("#64748b")),
				),
			))
	}
	return chadcn.Card(
		chadcn.CardHeader(
			chadcn.CardTitle("Recent sends"),
			chadcn.CardDescription(a.recentSendsDesc()),
		),
		chadcn.CardContent(rows...),
	)
}

func (a *app) overviewActionsCard() *ui.Element {
	campaignBtn := chadcn.Button("+ New campaign", chadcn.ButtonOptions{
		Variant: chadcn.ButtonOutlined,
		Size:    chadcn.ButtonSmall,
	}).Attr("data-action", "nav:campaigns")
	automationBtn := chadcn.Button("↻ New automation", chadcn.ButtonOptions{
		Variant: chadcn.ButtonOutlined,
		Size:    chadcn.ButtonSmall,
	}).Attr("data-action", "nav:automations")
	inviteBtn := chadcn.Button("⦿ Invite teammate", chadcn.ButtonOptions{
		Variant: chadcn.ButtonOutlined,
		Size:    chadcn.ButtonSmall,
	}).OnClick(func(ctx context.Context) error {
		a.toast = "Invite sent — your teammate will join shortly."
		a.toastOpen = true
		return a.refresh(ctx)
	})
	return chadcn.Card(
		chadcn.CardHeader(
			chadcn.CardTitle("Quick actions"),
			chadcn.CardDescription("Jump into your next send."),
		),
		chadcn.CardContent(
			ui.Div().Class("mp-actions").
				Display(ui.DisplayFlex).
				FlexDirection(ui.FlexDirectionColumn).
				Gap(ui.Px(8)).
				Children(campaignBtn, automationBtn, inviteBtn),
		),
	)
}

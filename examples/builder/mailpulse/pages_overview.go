package main

import (
	"context"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// overviewPage renders the Overview tab: list health, growth and send
// performance at a glance. It mirrors the engagement tab structure so the
// shell can swap it in without layout changes.
func (a *app) overviewPage() *ui.Element {
	return ui.Div().Class("mp-card").Children(
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
	exportBtn := chadcn.Button("⤓ Export", chadcn.ButtonOptions{
		Variant: chadcn.ButtonOutlined,
		Size:    chadcn.ButtonSmall,
	}).Attr("data-action", "export")
	left := ui.Div().Class("mp-pagehead-text").Children(
		chadcn.H1("Overview"),
		chadcn.Muted("List health, growth and send performance at a glance."),
	)
	actions := ui.Div().Class("mp-pagehead-actions").Children(
		chadcn.Badge("📅 "+a.rangeLabel),
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

func overviewStats() []overviewStat {
	return []overviewStat{
		{label: "Total subscribers", value: "18,442", delta: "+2.4%", footnote: "+428 this month", tint: "purple", glyph: "◉"},
		{label: "Avg. open rate", value: "47.8%", delta: "+3.1%", footnote: "Unique opens ÷ delivered", tint: "orange", glyph: "✉"},
		{label: "Avg. click rate", value: "18.6%", delta: "+1.9%", footnote: "Unique clicks ÷ delivered", tint: "blue", glyph: "➤"},
		{label: "Revenue/Mo", value: "$12,480", delta: "+8.2%", footnote: "Paid conversions this month", tint: "green", glyph: "↩"},
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
	stats := overviewStats()
	cards := make([]*ui.Element, 0, len(stats))
	for _, stat := range stats {
		cards = append(cards, overviewKPICard(stat))
	}
	return ui.Div().Class("mp-kpis").Children(cards...)
}

var overviewBarHeights = []float64{34, 48, 42, 55, 61, 58, 90, 66, 72, 64, 78, 84}

func (a *app) overviewGrowthCard() *ui.Element {
	bars := make([]*ui.Element, 0, len(overviewBarHeights))
	for i, height := range overviewBarHeights {
		color := "#c9bcfb"
		if i == 6 {
			color = "#7c3aed"
		}
		bars = append(bars, ui.Div().Class("mp-bar").
			Height(ui.Percent(height)).
			Width(ui.Px(14)).
			BackgroundColor(ui.Hex(color)).
			BorderRadius(ui.Px(4)))
	}
	return chadcn.Card(
		chadcn.CardHeader(
			chadcn.CardTitle("List growth"),
			chadcn.CardDescription("Net new subscribers per week, last 12 weeks."),
		),
		chadcn.CardContent(
			ui.Div().Class("mp-bars").
				Display(ui.DisplayFlex).
				AlignItems(ui.AlignItemsEnd).
				Gap(ui.Px(6)).
				Height(ui.Px(120)).
				Children(bars...),
			ui.Div().Class("mp-muted").Text("+428 net new subscribers over the last 12 weeks."),
		),
	)
}

type overviewSegment struct {
	label  string
	share  string
	filled int
}

func overviewSegments() []overviewSegment {
	return []overviewSegment{
		{label: "Power readers", share: "32%", filled: 8},
		{label: "New subscribers", share: "24%", filled: 7},
		{label: "Paid members", share: "15%", filled: 6},
		{label: "Dormant win-back", share: "18%", filled: 5},
		{label: "Referral friends", share: "11%", filled: 4},
	}
}

func overviewSegmentRow(segment overviewSegment) *ui.Element {
	dots := make([]*ui.Element, 0, 10)
	for i := 0; i < 10; i++ {
		class := "mp-dot"
		if i < segment.filled {
			class += " mp-dot--on"
		}
		dots = append(dots, ui.Span().Class(class))
	}
	return ui.Div().Class("mp-linkrow").Children(
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
	segments := overviewSegments()
	rows := make([]*ui.Element, 0, len(segments))
	for _, segment := range segments {
		rows = append(rows, overviewSegmentRow(segment))
	}
	return chadcn.Card(
		chadcn.CardHeader(
			chadcn.CardTitle("Top segments"),
			chadcn.CardDescription("Share of engaged readers by segment."),
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

func overviewSends() []overviewSend {
	return []overviewSend{
		{id: "send-1", name: "Weekly Creator Brief #52", date: "Mar 12", open: "52% opens", click: "21% clicks"},
		{id: "send-2", name: "AI Tools Worth Trying", date: "Mar 5", open: "49% opens", click: "19% clicks"},
		{id: "send-3", name: "The Creator Economy Report", date: "Feb 26", open: "47% opens", click: "17% clicks"},
		{id: "send-4", name: "Sponsor Spotlight: Notion", date: "Feb 19", open: "38% opens", click: "9% clicks"},
	}
}

func (a *app) overviewSendsCard() *ui.Element {
	sends := overviewSends()
	rows := make([]*ui.Element, 0, len(sends))
	for _, send := range sends {
		rows = append(rows, ui.Div().Class("mp-linkrow mp-selectable").
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
			chadcn.CardDescription("Last 4 campaigns and their results."),
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

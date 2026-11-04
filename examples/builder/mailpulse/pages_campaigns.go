package main

import (
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

type campaign struct {
	id     string
	name   string
	status string
	sent   string
	open   string
	click  string
}

func campaignRows() []campaign {
	return []campaign{
		{id: "brief-52", name: "Weekly Creator Brief #52", status: "Sent", sent: "May 28", open: "52.4%", click: "21.2%"},
		{id: "ai-tools", name: "AI Tools Worth Trying", status: "Sent", sent: "May 21", open: "48.9%", click: "19.4%"},
		{id: "economy-report", name: "Creator Economy Report", status: "Scheduled", sent: "Jun 4", open: "—", click: "—"},
		{id: "referral-push", name: "Referral Push", status: "Draft", sent: "—", open: "—", click: "—"},
		{id: "sponsor-special", name: "Sponsor Special", status: "Sent", sent: "May 14", open: "44.1%", click: "12.8%"},
		{id: "reengagement", name: "Re-engagement", status: "Paused", sent: "May 9", open: "31.2%", click: "6.4%"},
	}
}

func campaignStatusBadge(status string) *ui.Element {
	switch status {
	case "Sent":
		return chadcn.Badge(status)
	case "Scheduled":
		return chadcn.Badge(status, chadcn.BadgeOptions{Variant: chadcn.BadgeSecondary})
	default:
		return chadcn.Badge(status, chadcn.BadgeOptions{Variant: chadcn.BadgeOutline})
	}
}

func (a *app) campaignsPage() *ui.Element {
	rows := make([]*ui.Element, 0, len(campaignRows()))
	for _, row := range campaignRows() {
		rows = append(rows, ui.Div().Class("mp-linkrow "+selectableRow(false)).
			Attr("data-action", "select:campaign:"+row.id+":"+row.name).
			Children(
				ui.Div().Class("mp-linkmain").Children(
					ui.Text(row.name).FontWeight(ui.FontWeightBold).FontSize(ui.Px(13)),
					ui.Text(row.sent).FontSize(ui.Px(12)).Color(ui.Hex("#64748b")),
				),
				campaignStatusBadge(row.status),
				ui.Div().Class("mp-linkstats").Children(
					ui.Text(row.open).FontWeight(ui.FontWeightBold),
					ui.Text(row.click).Color(ui.Hex("#64748b")),
				),
			))
	}
	return ui.Div().Class("mp-card").Children(
		ui.Div().Class("mp-pagehead").Children(
			ui.Div().Class("mp-pagehead-text").Children(
				chadcn.H1("Campaigns"),
				chadcn.Muted("Sends, drafts and schedules."),
			),
		),
		ui.Div().Class("mp-heat-badges").Children(
			chadcn.Badge("Sent 24"),
			chadcn.Badge("Scheduled 3", chadcn.BadgeOptions{Variant: chadcn.BadgeSecondary}),
			chadcn.Badge("Draft 5", chadcn.BadgeOptions{Variant: chadcn.BadgeOutline}),
			chadcn.Badge("Paused 1", chadcn.BadgeOptions{Variant: chadcn.BadgeOutline}),
		),
		chadcn.Card(
			chadcn.CardHeader(
				chadcn.CardTitle("All campaigns"),
				chadcn.CardDescription("Every send, draft and scheduled issue."),
			),
			chadcn.CardContent(rows...),
		),
	)
}

type contentIssue struct {
	title string
	kind  string
	date  string
	open  string
}

func contentIssues() []contentIssue {
	return []contentIssue{
		{title: "Weekly Creator Brief #52", kind: "Newsletter", date: "May 28", open: "52.4%"},
		{title: "AI Tools Worth Trying", kind: "Roundup", date: "May 21", open: "48.9%"},
		{title: "Sponsor Special", kind: "Sponsored", date: "May 14", open: "44.1%"},
		{title: "Re-engagement: We Miss You", kind: "Winback", date: "May 9", open: "31.2%"},
		{title: "The Creator Playbook", kind: "Guide", date: "May 2", open: "39.7%"},
	}
}

type contentBlock struct {
	name string
	kind string
}

func contentBlocks() []contentBlock {
	return []contentBlock{
		{name: "Hero header", kind: "Banner"},
		{name: "CTA button", kind: "Action"},
		{name: "Quote", kind: "Text"},
		{name: "Footer", kind: "Layout"},
	}
}

func (a *app) contentPage() *ui.Element {
	issues := contentIssues()
	issueRows := make([]*ui.Element, 0, len(issues))
	for i, issue := range issues {
		badge := chadcn.Badge(issue.kind, chadcn.BadgeOptions{Variant: chadcn.BadgeOutline})
		if i == 0 {
			badge = chadcn.Badge(issue.kind)
		} else if i == 1 {
			badge = chadcn.Badge(issue.kind, chadcn.BadgeOptions{Variant: chadcn.BadgeSecondary})
		}
		issueRows = append(issueRows, ui.Div().Class("mp-linkrow").Children(
			ui.Div().Class("mp-linkmain").Children(
				ui.Text(issue.title).FontWeight(ui.FontWeightBold).FontSize(ui.Px(13)),
				ui.Text(issue.date).FontSize(ui.Px(12)).Color(ui.Hex("#64748b")),
			),
			badge,
			ui.Div().Class("mp-linkstats").Children(
				ui.Text(issue.open).FontWeight(ui.FontWeightBold),
			),
		))
	}
	blocks := contentBlocks()
	blockRows := make([]*ui.Element, 0, len(blocks))
	for _, block := range blocks {
		use := chadcn.Button("Use", chadcn.ButtonOptions{
			Variant: chadcn.ButtonOutlined,
			Size:    chadcn.ButtonSmall,
		}).Attr("data-action", "export")
		blockRows = append(blockRows, ui.Div().Class("mp-linkrow").Children(
			ui.Div().Class("mp-linkmain").Children(
				ui.Text(block.name).FontWeight(ui.FontWeightBold).FontSize(ui.Px(13)),
				ui.Text(block.kind).FontSize(ui.Px(12)).Color(ui.Hex("#64748b")),
			),
			use,
		))
	}
	return ui.Div().Class("mp-card").Children(
		ui.Div().Class("mp-pagehead").Children(
			ui.Div().Class("mp-pagehead-text").Children(
				chadcn.H1("Content"),
				chadcn.Muted("Issues, templates and saved blocks."),
			),
		),
		ui.Div().Class("mp-grid-bot").Children(
			chadcn.Card(
				chadcn.CardHeader(
					chadcn.CardTitle("Recent issues"),
					chadcn.CardDescription("Latest published issues and their open rates."),
				),
				chadcn.CardContent(issueRows...),
			),
			chadcn.Card(
				chadcn.CardHeader(
					chadcn.CardTitle("Saved blocks"),
					chadcn.CardDescription("Reusable sections for new issues."),
				),
				chadcn.CardContent(blockRows...),
			),
		),
	)
}

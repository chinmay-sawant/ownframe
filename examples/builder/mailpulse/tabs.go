package main

import (
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// integrator: replace the engagement card stack in shell.go (pageHead,
// kpiRow, heatmap/funnel/links/engage grids inside .mp-card) with
// a.contentBody(). contentBody returns the same engagement stack when
// a.tab == "engagement" and a placeholder page per tab otherwise, so the
// shell stays additive until the per-tab owners land their bodies.

// contentBody switches the card content on the active tab. Engagement keeps
// the existing dashboard stack; every other tab renders its head plus a
// placeholder card its owner will replace.
func (a *app) contentBody() *ui.Element {
	switch a.tab {
	case "overview":
		return a.overviewPage()
	case "audience":
		return a.audiencePage()
	case "campaigns":
		return a.campaignsPage()
	case "content":
		return a.contentPage()
	case "revenue":
		return a.revenuePage()
	case "subscribers":
		return a.subscribersPage()
	case "automations":
		return a.automationsPage()
	case "segments":
		return a.segmentsPage()
	case "integrations":
		return a.integrationsPage()
	case "analytics":
		return a.analyticsPage()
	case "reports":
		return a.reportsPage()
	case "learn":
		return a.learnPage()
	case "help":
		return a.helpPage()
	case "settings":
		return a.settingsPage()
	case "engagement":
		return a.engagementPage()
	default:
		return a.engagementPage()
	}
}

// pageTitle returns the breadcrumb/head title and subtitle for the active
// tab. The topbar breadcrumb and the per-tab head both read from here.
func (a *app) pageTitleOr(fallback string) string {
	title, _ := a.pageTitle()
	if title == "" {
		return fallback
	}
	return title
}

func (a *app) pageTitle() (string, string) {
	switch a.tab {
	case "overview":
		return "Overview", "Workspace health at a glance."
	case "audience":
		return "Audience", "Who your readers are and how they grow."
	case "campaigns":
		return "Campaigns", "Every send, scheduled or shipped."
	case "content":
		return "Content", "Issues, templates, and drafts."
	case "engagement":
		return "Engagement", "How readers open, click and reply to every send."
	case "revenue":
		return "Revenue", "Paid conversion from every send."
	case "subscribers":
		return "Subscribers", "Growth, churn, and list hygiene."
	case "automations":
		return "Automations", "Welcome, win-back, and drip flows."
	case "segments":
		return "Segments", "Slice the list by behavior and plan."
	case "integrations":
		return "Integrations", "Connect the rest of the stack."
	case "analytics":
		return "Analytics", "Deeper cuts beyond the dashboard."
	case "reports":
		return "Reports", "Scheduled digests and exports."
	case "learn":
		return "Learn", "Guides and deliverability playbooks."
	case "help":
		return "Help Center", "Answers and support."
	case "settings":
		return "System Settings", "Workspace, billing, and defaults."
	default:
		return "Engagement", "How readers open, click and reply to every send."
	}
}

// engagementPage is the existing dashboard stack, extracted so contentBody
// can return it without shell.go changes. It keeps its own .mp-card wrapper
// so the tab renders exactly as before; other tabs own their card structure.
func (a *app) engagementPage() *ui.Element {
	return ui.Div().Class("mp-card mp-stack").Children(
		a.pageHead(),
		a.kpiRow(),
		ui.Div().Class("mp-grid-mid").Children(
			a.heatmapCard(),
			a.funnelCard(),
		),
		ui.Div().Class("mp-grid-bot").Children(
			a.linksCard(),
			a.engageCard(),
		),
	)
}

// tabHead mirrors pageHead with the per-tab title from pageTitle, so every
// placeholder tab keeps the range/export controls until its owner replaces
// the body.
func (a *app) tabHead() *ui.Element {
	title, sub := a.pageTitle()
	rangeBtn := chadcn.Button(a.rangeLabel+" ▾", chadcn.ButtonOptions{
		Variant: chadcn.ButtonOutlined,
		Size:    chadcn.ButtonSmall,
	}).Attr("data-action", "range:cycle")
	exportBtn := chadcn.Button("↓ Export", chadcn.ButtonOptions{
		Variant: chadcn.ButtonOutlined,
		Size:    chadcn.ButtonSmall,
	}).Attr("data-action", "export")
	left := ui.Div().Class("mp-pagehead-text").Children(
		chadcn.H1(title),
		chadcn.Muted(sub),
	)
	actions := ui.Div().Class("mp-pagehead-actions").Children(rangeBtn, exportBtn)
	return ui.Div().Class("mp-pagehead").Children(left, actions)
}

// placeholderPage renders the tab head plus a stub card. Per-tab owners
// replace the stub body with real content.
func (a *app) placeholderPage(note string) *ui.Element {
	if note == "" {
		title, _ := a.pageTitle()
		note = title + " content lands here."
	}
	return ui.Div().Children(
		a.tabHead(),
		chadcn.Card(
			chadcn.CardHeader(
				chadcn.CardTitle("Coming soon"),
				chadcn.CardDescription(note),
			),
			chadcn.CardContent(
				chadcn.Muted("This tab is owned by another workstream."),
			),
		),
	)
}

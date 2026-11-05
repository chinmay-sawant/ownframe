//nolint:mnd // Sample management fixtures are literal display content: segment slices and dot counts are fixed demo values.
package main

import (
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

type autoFlow struct {
	label  string
	status string
	detail string
}

func automationFlows() []autoFlow {
	return []autoFlow{
		{label: "Welcome series", status: "Active", detail: "3 emails · 62.4% open"},
		{label: "Onboarding", status: "Active", detail: "5 emails · 54.1% open"},
		{label: "Win-back", status: "Paused", detail: "2 emails · 28.4% open"},
		{label: "Upsell", status: "Active", detail: "2 emails · 41.9% open"},
	}
}

func autoFlowID(label string) string {
	switch label {
	case "Welcome series":
		return "welcome"
	case "Onboarding":
		return "onboarding"
	case "Win-back":
		return "winback"
	case "Upsell":
		return "upsell"
	default:
		return "flow"
	}
}

func (a *app) automationsPage() *ui.Element {
	rows := make([]*ui.Element, 0, len(automationFlows()))
	for _, flow := range automationFlows() {
		flow := flow
		toggle := "Pause"
		if flow.status != "Active" {
			toggle = "Enable"
		}
		badge := chadcn.Badge(flow.status)
		if flow.status != "Active" {
			badge = chadcn.Badge(flow.status, chadcn.BadgeOptions{Variant: chadcn.BadgeSecondary})
		}
		rows = append(rows, ui.Div().Class("mp-linkrow mp-selectable").Children(
			ui.Div().Class("mp-linkmain").Children(
				ui.Text(flow.label).FontWeight(ui.FontWeightBold).FontSize(ui.Px(13)),
				ui.Text(flow.detail).FontSize(ui.Px(12)).Color(ui.Hex("#64748b")),
			),
			badge,
			chadcn.Button(toggle, chadcn.ButtonOptions{
				Variant: chadcn.ButtonOutlined,
				Size:    chadcn.ButtonSmall,
			}).Attr("data-action", "select:auto:"+autoFlowID(flow.label)+":"+flow.label),
		))
	}
	return ui.Div().Class("mp-card mp-manage mp-stack").Children(
		ui.Div().Class("mp-pagehead").Children(
			ui.Div().Class("mp-pagehead-text").Children(
				chadcn.H1("Automations"),
				chadcn.Muted("Welcome, win-back and upsell flows."),
			),
		),
		chadcn.Card(chadcn.CardContent(rows...)),
	)
}

type segmentSlice struct {
	label  string
	count  string
	filled int
}

func segmentSlices() []segmentSlice {
	return []segmentSlice{
		{label: "Power readers", count: "3,204", filled: 6},
		{label: "New subs", count: "1,204", filled: 3},
		{label: "Dormant", count: "2,010", filled: 4},
		{label: "Paid", count: "842", filled: 2},
		{label: "Free", count: "17,600", filled: 10},
	}
}

// manageDotTotal sizes the 10-dot share bars on the Segments rows.
const manageDotTotal = 10

func manageDots(filled, total int) *ui.Element {
	dots := make([]*ui.Element, 0, total)
	for i := 0; i < total; i++ {
		class := "mp-dot"
		if i < filled {
			class += " mp-dot--on"
		}
		dots = append(dots, ui.Span().Class(class))
	}
	return ui.Div().Class("mp-dots").Children(dots...)
}

func (a *app) segmentsPage() *ui.Element {
	slices := segmentSlices()
	rows := make([]*ui.Element, 0, len(slices))
	for _, slice := range slices {
		slice := slice
		rows = append(rows, ui.Div().Class("mp-linkrow mp-selectable").Children(
			ui.Div().Class("mp-linkmain").Children(
				ui.Text(slice.label).FontWeight(ui.FontWeightBold).FontSize(ui.Px(13)),
				manageDots(slice.filled, manageDotTotal),
			),
			ui.Div().Class("mp-linkstats").Children(
				ui.Text(slice.count).FontWeight(ui.FontWeightBold),
				chadcn.Button("View", chadcn.ButtonOptions{
					Variant: chadcn.ButtonOutlined,
					Size:    chadcn.ButtonSmall,
				}).Attr("data-action", "nav:subscribers"),
			),
		))
	}
	return ui.Div().Class("mp-card mp-manage mp-stack").Children(
		ui.Div().Class("mp-pagehead").Children(
			ui.Div().Class("mp-pagehead-text").Children(
				chadcn.H1("Segments"),
				chadcn.Muted("Targeted slices of the list."),
			),
		),
		chadcn.Card(chadcn.CardContent(rows...)),
	)
}

type integrationTool struct {
	id     string
	label  string
	status string
}

func integrationTools() []integrationTool {
	return []integrationTool{
		{id: "stripe", label: "Stripe", status: "Connected"},
		{id: "zapier", label: "Zapier", status: "Connected"},
		{id: "webhooks", label: "Webhooks", status: "Active"},
		{id: "api-keys", label: "API Keys", status: "2 keys"},
	}
}

func (a *app) integrationsPage() *ui.Element {
	tools := integrationTools()
	rows := make([]*ui.Element, 0, len(tools))
	for _, tool := range tools {
		tool := tool
		badge := chadcn.Badge(tool.status)
		if tool.id == "api-keys" {
			badge = chadcn.Badge(tool.status, chadcn.BadgeOptions{Variant: chadcn.BadgeOutline})
		}
		rows = append(rows, ui.Div().Class("mp-linkrow mp-selectable").Children(
			ui.Div().Class("mp-linkmain").Children(
				ui.Text(tool.label).FontWeight(ui.FontWeightBold).FontSize(ui.Px(13)),
				ui.Text(tool.status).FontSize(ui.Px(12)).Color(ui.Hex("#64748b")),
			),
			badge,
			chadcn.Button("Configure", chadcn.ButtonOptions{
				Variant: chadcn.ButtonOutlined,
				Size:    chadcn.ButtonSmall,
			}).Attr("data-action", "select:integration:"+tool.id+":"+tool.label),
		))
	}
	return ui.Div().Class("mp-card mp-manage mp-stack").Children(
		ui.Div().Class("mp-pagehead").Children(
			ui.Div().Class("mp-pagehead-text").Children(
				chadcn.H1("Integrations"),
				chadcn.Muted("Connected tools."),
			),
		),
		chadcn.Card(chadcn.CardContent(rows...)),
	)
}

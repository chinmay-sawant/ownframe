package main

import (
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

type assistGuide struct {
	id       string
	title    string
	duration string
}

func assistGuides() []assistGuide {
	return []assistGuide{
		{id: "open-rate", title: "Double your open rate", duration: "12 min"},
		{id: "segmentation", title: "Segmentation 101", duration: "8 min"},
		{id: "sponsor", title: "Sponsor playbook", duration: "15 min"},
		{id: "deliverability", title: "Deliverability checklist", duration: "6 min"},
	}
}

type assistFAQ struct {
	id    string
	label string
}

func assistFAQs() []assistFAQ {
	return []assistFAQ{
		{id: "import", label: "How do I import?"},
		{id: "opens", label: "Why did opens drop?"},
		{id: "payouts", label: "How do payouts work?"},
		{id: "pause", label: "Can I pause?"},
	}
}

func (a *app) learnPage() *ui.Element {
	_ = a
	guides := assistGuides()
	rows := make([]*ui.Element, 0, len(guides))
	for _, guide := range guides {
		rows = append(rows, ui.Div().Class("mp-linkrow mp-learn-row").Children(
			ui.Div().Class("mp-linkmain").
				FlexGrow(ui.Number(1)).
				MinWidth(ui.Px(0)).
				Children(
					ui.Text(guide.title).FontWeight(ui.FontWeightBold).FontSize(ui.Px(13)),
					ui.Text(guide.duration+" · Guide").FontSize(ui.Px(12)).Color(ui.Hex("#6b7280")),
				),
			chadcn.Button("Start", chadcn.ButtonOptions{
				Variant: chadcn.ButtonOutlined,
				Size:    chadcn.ButtonSmall,
			}).Attr("data-action", "select:guide:"+guide.id+":"+guide.title).
				FlexShrink(ui.Number(0)),
		))
	}
	return ui.Div().Class("mp-card mp-page mp-stack").Children(
		ui.Div().Class("mp-pagehead").Children(
			ui.Div().Class("mp-pagehead-text").Children(
				chadcn.H1("Learn"),
				chadcn.Muted("Guides to grow the list."),
			),
		),
		chadcn.Card(
			chadcn.CardHeader(
				chadcn.CardTitle("Guides"),
				chadcn.CardDescription("Short lessons to grow the list."),
			),
			chadcn.CardContent(rows...),
		),
	)
}

func (a *app) helpPage() *ui.Element {
	_ = a
	faqs := assistFAQs()
	rows := make([]*ui.Element, 0, len(faqs))
	for _, faq := range faqs {
		rows = append(rows, ui.Div().Class("mp-linkrow mp-selectable").
			Attr("data-action", "select:faq:"+faq.id+":"+faq.label).
			Children(
				ui.Text(faq.label).FontWeight(ui.FontWeightBold).FontSize(ui.Px(13)),
				ui.Text("›").FontSize(ui.Px(16)).Color(ui.Hex("#6b7280")),
			))
	}
	contact := chadcn.Button("Contact", chadcn.ButtonOptions{
		Variant: chadcn.ButtonOutlined,
		Size:    chadcn.ButtonSmall,
	}).Attr("data-action", "export")
	return ui.Div().Class("mp-card mp-page mp-stack").Children(
		ui.Div().Class("mp-pagehead").Children(
			ui.Div().Class("mp-pagehead-text").Children(
				chadcn.H1("Help Center"),
				chadcn.Muted("Answers and contact."),
			),
			ui.Div().Class("mp-pagehead-actions").Children(contact),
		),
		chadcn.Card(
			chadcn.CardHeader(
				chadcn.CardTitle("Frequently asked"),
				chadcn.CardDescription("Quick answers to common questions."),
			),
			chadcn.CardContent(rows...),
		),
	)
}

func (a *app) settingsPage() *ui.Element {
	_ = a
	profile := chadcn.Card(
		chadcn.CardHeader(
			chadcn.CardTitle("Profile"),
			chadcn.CardDescription("How your workspace appears to readers."),
		),
		chadcn.CardContent(
			chadcn.InputWithOptions("set-name", "Jordan Doe", chadcn.InputOptions{
				Placeholder: "Name",
			}),
			chadcn.InputWithOptions("set-email", "jordan@mailpulse.co", chadcn.InputOptions{
				Type:        "email",
				Placeholder: "Email",
			}),
		).Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(12)),
	)
	sending := chadcn.Card(
		chadcn.CardHeader(
			chadcn.CardTitle("Sending"),
			chadcn.CardDescription("Daily sending limit for the workspace."),
		),
		chadcn.CardContent(
			ui.Div().Class("mp-linkrow").Children(
				ui.Text("Daily limit").FontWeight(ui.FontWeightBold).FontSize(ui.Px(13)),
				chadcn.Badge("50k", chadcn.BadgeOptions{Variant: chadcn.BadgeSecondary}),
				chadcn.Button("Pause", chadcn.ButtonOptions{
					Variant: chadcn.ButtonOutlined,
					Size:    chadcn.ButtonSmall,
				}).Attr("data-action", "select:setting:pause:Pause"),
			),
		),
	)
	billing := chadcn.Card(
		chadcn.CardHeader(
			chadcn.CardTitle("Billing"),
			chadcn.CardDescription("Plan Pro $49/mo, billed monthly."),
		),
		chadcn.CardContent(
			ui.Div().Class("mp-linkrow").Children(
				ui.Text("Current plan").FontWeight(ui.FontWeightBold).FontSize(ui.Px(13)),
				chadcn.Badge("Pro $49/mo"),
				chadcn.Button("Manage", chadcn.ButtonOptions{
					Variant: chadcn.ButtonOutlined,
					Size:    chadcn.ButtonSmall,
				}).Attr("data-action", "export"),
			),
		),
	)
	return ui.Div().Class("mp-card mp-page mp-stack").Children(
		ui.Div().Class("mp-pagehead").Children(
			ui.Div().Class("mp-pagehead-text").Children(
				chadcn.H1("System Settings"),
				chadcn.Muted("Workspace, sending and billing."),
			),
		),
		profile,
		sending,
		billing,
	)
}

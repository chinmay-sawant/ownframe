package main

import (
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

func (a *app) sidebar() *ui.Element {
	brand := ui.Div().Class("mp-brand").Children(
		ui.Div().Class("mp-brand-mark").Text("M"),
		ui.Text("MailPulse").FontWeight(ui.FontWeightBold).FontSize(ui.Px(15)),
	)

	sections := navSections()
	groups := make([]*ui.Element, 0, len(sections))
	for _, section := range sections {
		rows := make([]*ui.Element, 0, len(section.items))
		for _, item := range section.items {
			label := item.label
			id := navID(label)
			if item.action != "" {
				if t := tabFromAction(item.action); t != "" {
					id = t
				}
			}
			rows = append(rows, chadcn.SidebarNavItem(chadcn.SidebarItem{
				Label:  label,
				Icon:   ui.Text(item.icon),
				Active: a.tab == id,
				Badge:  item.badge,
				Action: "nav:" + id,
			}))
		}
		groups = append(groups, chadcn.SidebarSection(section.title, rows...))
	}
	content := ui.Div().Class("mp-nav").Children(groups...)

	footer := ui.Div().Class("mp-user").Children(
		chadcn.Avatar(chadcn.AvatarOptions{Initials: "JD"}),
		ui.Div().Class("mp-user-text").Children(
			ui.Text("Jordan Doe").FontWeight(ui.FontWeightBold).FontSize(ui.Px(12)),
			ui.Text("Admin").FontSize(ui.Px(11)).Color(ui.ColorVar("--of-muted", ui.Hex("#6b7280"))),
		),
	)

	return chadcn.Sidebar(chadcn.SidebarOptions{
		ID:        "mp-sidebar",
		Label:     "MailPulse navigation",
		Open:      true,
		Collapsed: false,
		Width:     ui.Px(248),
		Header:    brand,
		Content:   content,
		Footer:    footer,
	})
}

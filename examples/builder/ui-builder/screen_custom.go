package main

import (
	"context"
	"fmt"

	"github.com/chinmay-sawant/ownframe/examples/builder/ui-builder/custom"
	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/components"
	"github.com/chinmay-sawant/ownframe/ui/theme"
)

type customState struct {
	pinned [2]bool
	opened string
}

func (a *builderApp) customPanel() *ui.Element {
	return ui.Div().Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(16)).Children(
		a.panelHeading("Build your own component"),
		ui.P().Text("Compose library controls in your own package. Pass data, actions, and optional local properties.").
			Margin(ui.Px(0)).Color(theme.Muted()).FontSize(ui.Px(13)),
		ui.Div().Display(ui.DisplayFlex).Gap(ui.Px(16)).Children(
			a.projectExample(0, "Workspace", "Inherits the global application palette.", nil).Flex(ui.Number(1)).MinWidth(ui.Px(0)),
			a.projectExample(1, "Release kit", "Keeps its app-owned accent and local action color.", projectProperties()).Flex(ui.Number(1)).MinWidth(ui.Px(0)),
		),
		components.Notice("Override order", "OwnFrame defaults → application theme → component properties. Switch the global palette above; the second card keeps its local values."),
		components.Card(
			ui.Text("custom.ProjectCard(options)").FontFamily(ui.FontFamilyMonospace).FontWeight(ui.FontWeightBold),
			ui.P().Text("ProjectOptions supplies Name, Description, Done, Total, Pinned, OnPin, OnOpen, and Overrides.").
				Color(theme.Muted()).FontSize(ui.Px(12)),
			ui.Text("examples/builder/ui-builder/custom/project.go").FontFamily(ui.FontFamilyMonospace).FontSize(ui.Px(12)),
		),
	)
}

func (a *builderApp) projectExample(index int, name, description string, overrides theme.Overrides) *ui.Element {
	const initialDone = 3
	const totalTasks = 8
	return custom.ProjectCard(custom.ProjectOptions{
		ID: fmt.Sprintf("custom-project-%d", index), Name: name, Description: description, Done: initialDone + index, Total: totalTasks,
		Pinned: a.custom.pinned[index], Overrides: overrides,
		OnPin: func(ctx context.Context) error {
			a.custom.pinned[index] = !a.custom.pinned[index]
			a.notice = name + " pin updated."
			return a.refresh(ctx)
		},
		OnOpen: func(ctx context.Context) error {
			a.custom.opened = name
			a.notice = "Opened " + name + " in this demo."
			return a.refresh(ctx)
		},
	})
}

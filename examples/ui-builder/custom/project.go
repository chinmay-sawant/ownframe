// Package custom demonstrates app-owned controls composed from OwnFrame's
// component library. It has no dependency on the gallery application.
package custom

import (
	"context"
	"fmt"

	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/components"
	"github.com/chinmay-sawant/ownframe/ui/theme"
)

// ProjectOptions is all the state and behavior a project card consumes.
// The host owns state; each call constructs a fresh tree.
type ProjectOptions struct {
	ID          string
	Name        string
	Description string
	Done        int
	Total       int
	Pinned      bool
	OnPin       func(context.Context) error
	OnOpen      func(context.Context) error
	Overrides   theme.Overrides
}

// ProjectCard composes a card, badge, progress bar, and action buttons.
// --project-accent is an app-owned property with the global primary as fallback.
func ProjectCard(options ProjectOptions) *ui.Element {
	total := max(options.Total, 0)
	done := min(max(options.Done, 0), total)
	const percentScale = 100
	percent := 0
	if total > 0 {
		percent = int(float64(done) / float64(total) * percentScale)
	}
	label := "Pin project"
	if options.Pinned {
		label = "Unpin project"
	}
	accent := ui.ColorVar("--project-accent", theme.Primary())
	return options.Overrides.Apply(components.Card().ID(options.ID).
		BorderLeft(ui.SolidBorder(ui.Px(3), accent)).Display(ui.DisplayFlex).
		FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(14)).Children(
		ui.Div().Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).JustifyContent(ui.JustifyContentSpaceBetween).Children(
			ui.Text(options.Name).FontSize(ui.Px(20)).FontWeight(ui.FontWeightBold),
			components.Badge("Custom control").Color(theme.Text()),
		),
		ui.P().Text(options.Description).Margin(ui.Px(0)).Color(theme.Muted()).FontSize(ui.Px(13)),
		components.Progress(options.Name+" completion", percent),
		ui.Text(fmt.Sprintf("%d of %d tasks complete", done, total)).Color(theme.Muted()).FontSize(ui.Px(12)),
		ui.Div().Display(ui.DisplayFlex).Gap(ui.Px(10)).Children(
			components.Button("Open project", components.ButtonOptions{Variant: components.ButtonFilled, Disabled: options.OnOpen == nil}).ID(options.ID+"-open").OnClick(options.OnOpen),
			components.Button(label, components.ButtonOptions{Variant: components.ButtonOutlined, Disabled: options.OnPin == nil}).ID(options.ID+"-pin").OnClick(options.OnPin),
		),
	))
}

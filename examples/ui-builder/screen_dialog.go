package main

import (
	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/components"
)

// dialogState owns the backdrop dialog demo. builderApp carries one as
// a.dialog; the overlay mirrors examples/jmap-mail settingsCard.
type dialogState struct {
	open      bool
	confirmed bool
}

func defaultDialogState() dialogState {
	return dialogState{}
}

func (a *builderApp) dialogPanel() *ui.Element {
	return ui.Section().Flex(ui.RawValue("1")).MinWidth(ui.Px(0)).MinHeight(ui.Px(0)).
		Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(14)).
		Children(
			ui.Div().MaxWidth(ui.Px(600)).Width(ui.Percent(100)).
				Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(14)).
				Children(
					ui.Div().Children(
						a.panelHeading("Backdrop dialog"),
						ui.P().Text("An overlay with a translucent backdrop over the app, like the mail settings card. Open it, confirm, cancel, or click the backdrop to dismiss.").
							Margin(ui.RawValue("7px 0 0")).Color(ui.RawValue("var(--of-muted,#a9bbd1)")).FontSize(ui.Px(12)),
					),
					ui.Div().Padding(ui.Px(16)).BorderRadius(ui.Px(11)).BackgroundColor(ui.RawValue("var(--of-surface,#202d42)")).
						Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(10)).
						Children(
							ui.Button("Open dialog").ID("dialog-open").Attr("type", "button").Attr("data-action", "dialog:open").
								BackgroundColor(ui.RawValue("var(--of-primary,#4677ed)")).Color(ui.RawValue("var(--of-on-primary,#ffffff)")).Border(ui.RawValue("1px solid var(--of-border,#4677ed)")).
								Padding(ui.Edges(ui.Px(9), ui.Px(14))).BorderRadius(ui.Px(8)).FontSize(ui.Px(12)).FontWeight(ui.FontWeightBold),
							ui.Text(a.dialog.status()).FontSize(ui.Px(12)).Color(ui.RawValue("var(--of-muted,#b5c7df)")),
						),
					ui.Div().Padding(ui.Edges(ui.Px(12), ui.Px(15))).BackgroundColor(ui.RawValue("var(--of-surface,#202d42)")).BorderRadius(ui.Px(10)).
						Children(
							ui.Text("How it works").FontWeight(ui.FontWeightBold),
							ui.P().Text("The overlay is position:absolute inset 0 with z-index 50 and a translucent background, holding a centered 420px panel (id dialog-panel, no action) so interior clicks land harmlessly while backdrop clicks carry dialog:close. Same shape as jmap-mail settings-card.").
								Margin(ui.RawValue("8px 0 0")).FontSize(ui.Px(12)).Color(ui.RawValue("var(--of-muted,#abbcd2)")),
						),
				),
		)
}

func (s *dialogState) status() string {
	if s.confirmed {
		return "Last result: confirmed."
	}
	if s.open {
		return "Dialog is open."
	}
	return "Dialog is closed."
}

// dialogOverlay renders the backdrop + centered panel, or nil when closed.
// It is mounted at the screen root so the backdrop covers the whole window.
// A nil child is skipped by ui.Children, so the open flag fully owns it.
func (a *builderApp) dialogOverlay() *ui.Element {
	if a.tab != "dialog" || !a.dialog.open {
		return nil
	}
	return components.Dialog(components.DialogOptions{
		BackdropID: "dialog-backdrop", PanelID: "dialog-panel", Label: "Confirm action", CloseAction: "dialog:close", Width: ui.Px(420),
	},
		ui.Div().Display(ui.DisplayFlex).JustifyContent(ui.JustifyContentSpaceBetween).
			AlignItems(ui.AlignItemsCenter).Padding(ui.Edges(ui.Px(16), ui.Px(20))).
			BorderBottom(ui.RawValue("1px solid var(--of-border,#39465a)")).
			Children(
				ui.Text("Confirm action").FontSize(ui.Px(16)).FontWeight(ui.FontWeightBold).Color(ui.RawValue("var(--of-text,#f4f7ff)")),
				ui.Button("×").ID("dialog-x").Attr("type", "button").Attr("data-action", "dialog:close").
					BackgroundColor(ui.RawValue("var(--of-control,#2b3950)")).Color(ui.Hex("#dce6f6")).Border(ui.RawValue("1px solid var(--of-border,#42536b)")).
					Padding(ui.Edges(ui.Px(6), ui.Px(12))).BorderRadius(ui.Px(8)).FontWeight(ui.FontWeightBold),
			),
		ui.Div().Padding(ui.Edges(ui.Px(18), ui.Px(20), ui.Px(20))).
			Children(
				ui.P().Text("This dialog floats above the gallery on a backdrop. Confirming records the choice in Go state; clicking the dimmed backdrop dismisses without choosing.").
					Margin(ui.Px(0)).FontSize(ui.Px(13)).Color(ui.RawValue("var(--of-muted,#b5c7df)")),
				ui.Div().Display(ui.DisplayFlex).Gap(ui.Px(8)).MarginTop(ui.Px(16)).
					Children(
						ui.Button("Confirm").ID("dialog-confirm").Attr("type", "button").Attr("data-action", "dialog:confirm").
							BackgroundColor(ui.RawValue("var(--of-primary,#4677ed)")).Color(ui.RawValue("var(--of-on-primary,#ffffff)")).Border(ui.RawValue("1px solid var(--of-border,#4677ed)")).
							Padding(ui.Edges(ui.Px(9), ui.Px(14))).BorderRadius(ui.Px(8)).FontSize(ui.Px(12)).FontWeight(ui.FontWeightBold),
						ui.Button("Cancel").ID("dialog-cancel").Attr("type", "button").Attr("data-action", "dialog:close").
							BackgroundColor(ui.RawValue("var(--of-control,#2b3950)")).Color(ui.Hex("#dce6f6")).Border(ui.RawValue("1px solid var(--of-border,#42536b)")).
							Padding(ui.Edges(ui.Px(9), ui.Px(14))).BorderRadius(ui.Px(8)).FontSize(ui.Px(12)).FontWeight(ui.FontWeightBold),
					),
			),
	)
}

func (a *builderApp) handleDialog(action string) bool {
	switch action {
	case "dialog:open":
		a.dialog.open = true
		a.dialog.confirmed = false
		a.notice = "Dialog opened over a backdrop."
		return true
	case "dialog:close":
		if !a.dialog.open {
			return false
		}
		a.dialog.open = false
		a.notice = "Dialog dismissed."
		return true
	case "dialog:confirm":
		if !a.dialog.open {
			return false
		}
		a.dialog.open = false
		a.dialog.confirmed = true
		a.notice = "Dialog confirmed."
		return true
	}
	return false
}

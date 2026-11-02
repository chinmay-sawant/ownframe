package main

import (
	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/components"
	"github.com/chinmay-sawant/ownframe/ui/theme"
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
	return ui.Section().Flex(ui.Number(1)).MinWidth(ui.Px(0)).MinHeight(ui.Px(0)).
		Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(14)).
		Children(
			ui.Div().MaxWidth(ui.Px(600)).Width(ui.Percent(100)).
				Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(14)).
				Children(
					ui.Div().Children(
						a.panelHeading("Backdrop dialog"),
						ui.P().Text("An overlay with a translucent backdrop over the app, like the mail settings card. Open it, confirm, cancel, or click the backdrop to dismiss.").
							Margin(ui.Edges(ui.Px(7), ui.Px(0), ui.Px(0))).Color(theme.Muted()).FontSize(ui.Px(12)),
					),
					ui.Div().Padding(ui.Px(16)).BorderRadius(ui.Px(11)).BackgroundColor(theme.Surface()).
						Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(10)).
						Children(
							components.Button("Open dialog", components.ButtonOptions{Variant: components.ButtonFilled}).ID("dialog-open").Attr("data-action", "dialog:open"),
							ui.Text(a.dialog.status()).FontSize(ui.Px(12)).Color(theme.Muted()),
						),
					ui.Div().Padding(ui.Edges(ui.Px(12), ui.Px(15))).BackgroundColor(theme.Surface()).BorderRadius(ui.Px(10)).
						Children(
							ui.Text("How it works").FontWeight(ui.FontWeightBold),
							ui.P().Text("The overlay is position:absolute inset 0 with z-index 50 and a translucent background, holding a centered 420px panel (id dialog-panel, no action) so interior clicks land harmlessly while backdrop clicks carry dialog:close. Same shape as jmap-mail settings-card.").
								Margin(ui.Edges(ui.Px(8), ui.Px(0), ui.Px(0))).FontSize(ui.Px(12)).Color(theme.Muted()),
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
			BorderBottom(theme.Border()).
			Children(
				ui.Text("Confirm action").FontSize(ui.Px(16)).FontWeight(ui.FontWeightBold).Color(theme.Text()),
				components.Button("×", components.ButtonOptions{Variant: components.ButtonTonal}).ID("dialog-x").Attr("data-action", "dialog:close"),
			),
		ui.Div().Padding(ui.Edges(ui.Px(18), ui.Px(20), ui.Px(20))).
			Children(
				ui.P().Text("This dialog floats above the gallery on a backdrop. Confirming records the choice in Go state; clicking the dimmed backdrop dismisses without choosing.").
					Margin(ui.Px(0)).FontSize(ui.Px(13)).Color(theme.Muted()),
				ui.Div().Display(ui.DisplayFlex).Gap(ui.Px(8)).MarginTop(ui.Px(16)).
					Children(
						components.Button("Confirm", components.ButtonOptions{Variant: components.ButtonFilled}).ID("dialog-confirm").Attr("data-action", "dialog:confirm"),
						components.Button("Cancel", components.ButtonOptions{Variant: components.ButtonTonal}).ID("dialog-cancel").Attr("data-action", "dialog:close"),
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

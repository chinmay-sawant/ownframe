package main

import (
	"github.com/chinmay-sawant/ownframe/ui"
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
	return ui.Section().CSS("flex", "1").CSS("min-width", "0px").CSS("min-height", "0px").
		Display("flex").FlexDirection("column").Gap("14px").
		Children(
			ui.Div().CSS("max-width", "600px").CSS("width", "100%").
				Display("flex").FlexDirection("column").Gap("14px").
				Children(
					ui.Div().Children(
						ui.Text("Backdrop dialog").FontSize("24px").FontWeight("bold").Color("#f4f7ff"),
						ui.P().Text("An overlay with a translucent backdrop over the app, like the mail settings card. Open it, confirm, cancel, or click the backdrop to dismiss.").
							Margin("7px 0 0").Color("#a9bbd1").FontSize("12px"),
					),
					ui.Div().Padding("16px").BorderRadius("11px").BackgroundColor("#202d42").
						Display("flex").AlignItems("center").Gap("10px").
						Children(
							ui.Button("Open dialog").ID("dialog-open").Attr("type", "button").Attr("data-action", "dialog:open").
								BackgroundColor("#4677ed").Color("#ffffff").CSS("border", "1px solid #4677ed").
								Padding("9px 14px").BorderRadius("8px").FontSize("12px").FontWeight("bold"),
							ui.Text(a.dialog.status()).FontSize("12px").Color("#b5c7df"),
						),
					ui.Div().Padding("12px 15px").BackgroundColor("#202d42").BorderRadius("10px").
						Children(
							ui.Text("How it works").FontWeight("bold"),
							ui.P().Text("The overlay is position:absolute inset 0 with z-index 50 and a translucent background, holding a centered 420px panel (id dialog-panel, no action) so interior clicks land harmlessly while backdrop clicks carry dialog:close. Same shape as jmap-mail settings-card.").
								Margin("8px 0 0").FontSize("12px").Color("#abbcd2"),
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
	return ui.Div().ID("dialog-backdrop").Attr("data-action", "dialog:close").
		CSS("position", "absolute").CSS("top", "0").CSS("left", "0").
		CSS("right", "0").CSS("bottom", "0").CSS("z-index", "50").
		Display("flex").CSS("align-items", "center").CSS("justify-content", "center").
		BackgroundColor("rgba(0, 0, 0, 0.55)").
		Children(
			ui.Div().ID("dialog-panel").
				Width("420px").
				BackgroundColor("#1c2738").
				CSS("border", "1px solid #39465a").
				BorderRadius("12px").CSS("overflow", "hidden").
				Children(
					ui.Div().Display("flex").CSS("justify-content", "space-between").
						AlignItems("center").Padding("16px 20px").
						CSS("border-bottom", "1px solid #39465a").
						Children(
							ui.Text("Confirm action").FontSize("16px").FontWeight("bold").Color("#f4f7ff"),
							ui.Button("×").ID("dialog-x").Attr("type", "button").Attr("data-action", "dialog:close").
								BackgroundColor("#2b3950").Color("#dce6f6").CSS("border", "1px solid #42536b").
								Padding("6px 12px").BorderRadius("8px").FontWeight("bold"),
						),
					ui.Div().Padding("18px 20px 20px").
						Children(
							ui.P().Text("This dialog floats above the gallery on a backdrop. Confirming records the choice in Go state; clicking the dimmed backdrop dismisses without choosing.").
								Margin("0").FontSize("13px").Color("#b5c7df"),
							ui.Div().Display("flex").Gap("8px").CSS("margin-top", "16px").
								Children(
									ui.Button("Confirm").ID("dialog-confirm").Attr("type", "button").Attr("data-action", "dialog:confirm").
										BackgroundColor("#4677ed").Color("#ffffff").CSS("border", "1px solid #4677ed").
										Padding("9px 14px").BorderRadius("8px").FontSize("12px").FontWeight("bold"),
									ui.Button("Cancel").ID("dialog-cancel").Attr("type", "button").Attr("data-action", "dialog:close").
										BackgroundColor("#2b3950").Color("#dce6f6").CSS("border", "1px solid #42536b").
										Padding("9px 14px").BorderRadius("8px").FontSize("12px").FontWeight("bold"),
								),
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

package main

import (
	"strings"

	"github.com/chinmay-sawant/ownframe/ui"
)

type compareState struct {
	Mode    string
	Preview bool
}

func defaultCompareState() compareState {
	return compareState{Mode: "split"}
}

func builderCardSnippet() string {
	return `ui.Div().Display("flex").FlexDirection("column").Gap("10px").Padding("16px").BackgroundColor("#202d42").BorderRadius("11px").Children(
    ui.Text("Deploy preview").FontSize("13px").Color("#b5c7df"),
    ui.P().Text("1,248 rows").Margin("0").FontSize("29px").FontWeight("bold"),
    ui.Button("Open preview").Attr("data-action", "compare:open").CSS("align-self", "flex-start").
        BackgroundColor("#4677ed").Color("#ffffff").Padding("9px 12px").BorderRadius("8px"),
)`
}

func (a *builderApp) comparePanel() *ui.Element {
	mode := a.compare.Mode
	switch mode {
	case "split", "builder", "html":
	default:
		mode = "split"
	}
	modes := []string{"split", "builder", "html"}
	toggles := make([]*ui.Element, 0, len(modes))
	for _, m := range modes {
		bg := "#2b3950"
		if m == mode {
			bg = "#355596"
		}
		toggles = append(toggles, ui.Button(m).Attr("data-action", "compare:mode:"+m).
			BackgroundColor(bg).Color("#f1f5fd").
			CSS("border", "1px solid #40546f").
			Padding("9px 12px").BorderRadius("8px").FontWeight("bold"))
	}
	content := ui.Div().Display("flex").CSS("min-width", "0px").Gap("12px")
	switch mode {
	case "builder":
		content.Children(a.compareLivePane())
	case "html":
		content.Children(a.compareCodePane())
	default:
		content.Children(a.compareLivePane(), a.compareCodePane())
	}
	return ui.Section().CSS("flex", "1").CSS("min-width", "0px").CSS("min-height", "0px").
		Display("flex").FlexDirection("column").Gap("14px").
		Children(
			ui.Div().Children(
				ui.Text("Builder vs HTML").FontSize("24px").FontWeight("bold").Color("#f4f7ff"),
				ui.P().Text("The same three-field card, built with Go and shown as escaped source.").
					Margin("7px 0 0").Color("#a9bbd1").FontSize("12px"),
			),
			ui.Div().Display("flex").Gap("8px").Children(toggles...),
			content,
			ui.P().Text("Render emits plain HTML to Blinkless, no JS/DOM.").
				Margin("0").FontSize("12px").Color("#abbcd2"),
		)
}

func (a *builderApp) compareLivePane() *ui.Element {
	return ui.Div().CSS("flex", "1").CSS("min-width", "0px").
		Padding("16px").BorderRadius("11px").BackgroundColor("#202d42").
		Children(
			ui.Text("Live builder card").FontWeight("bold").Color("#f4f7ff"),
			ui.Div().CSS("margin-top", "10px").Children(a.compareCard()),
		)
}

func (a *builderApp) compareCodePane() *ui.Element {
	return ui.Div().CSS("flex", "1").CSS("min-width", "0px").
		Padding("16px").BorderRadius("11px").BackgroundColor("#202d42").
		Children(
			ui.Text("Same card as HTML source").FontWeight("bold").Color("#f4f7ff"),
			ui.Tag("pre").CSS("margin-top", "10px").
				Padding("12px").BorderRadius("8px").
				BackgroundColor("#111827").Color("#cbd8eb").FontSize("12px").
				CSS("font-family", "monospace").
				CSS("white-space", "pre-wrap").CSS("overflow-wrap", "break-word").
				CSS("border", "1px solid #3a4b62").
				Children(
					ui.Tag("code").Text(builderCardSnippet()),
				),
		)
}

func (a *builderApp) compareCard() *ui.Element {
	return ui.Div().Padding("16px").BackgroundColor("#202d42").BorderRadius("11px").
		CSS("border", "1px solid #3e526e").
		Display("flex").FlexDirection("column").Gap("10px").
		Children(
			ui.Text("Deploy preview").FontSize("13px").Color("#b5c7df"),
			ui.P().Text("1,248 rows").Margin("0").
				FontSize("29px").FontWeight("bold").Color("#f4f7ff"),
			ui.Button("Open preview").ID("compare-open").Attr("data-action", "compare:open").
				CSS("align-self", "flex-start").
				BackgroundColor("#4677ed").Color("#ffffff").
				Padding("9px 12px").BorderRadius("8px").FontWeight("bold"),
		)
}

// comparePreviewOverlay floats the live card above the tab on a backdrop,
// or nil when closed. Mounted at the screen root like the dialog overlay.
func (a *builderApp) comparePreviewOverlay() *ui.Element {
	if a.tab != "compare" || !a.compare.Preview {
		return nil
	}
	return ui.Div().ID("compare-preview-backdrop").Attr("data-action", "compare:preview-close").
		CSS("position", "absolute").CSS("top", "0").CSS("left", "0").
		CSS("right", "0").CSS("bottom", "0").CSS("z-index", "50").
		Display("flex").CSS("align-items", "center").CSS("justify-content", "center").
		BackgroundColor("rgba(0, 0, 0, 0.55)").
		Children(
			ui.Div().ID("compare-preview-panel").
				Width("480px").
				BackgroundColor("#1c2738").
				CSS("border", "1px solid #39465a").
				BorderRadius("12px").CSS("overflow", "hidden").
				Children(
					ui.Div().Display("flex").CSS("justify-content", "space-between").
						AlignItems("center").Padding("16px 20px").
						CSS("border-bottom", "1px solid #39465a").
						Children(
							ui.Text("Card preview").FontSize("16px").FontWeight("bold").Color("#f4f7ff"),
							ui.Button("×").ID("compare-preview-x").Attr("type", "button").Attr("data-action", "compare:preview-close").
								BackgroundColor("#2b3950").Color("#dce6f6").CSS("border", "1px solid #42536b").
								Padding("6px 12px").BorderRadius("8px").FontWeight("bold"),
						),
					ui.Div().Padding("18px 20px 20px").
						Children(
							a.compareCard(),
							ui.P().Text("The live card, floating above the tab. Backdrop click or × closes it.").
								Margin("12px 0 0").FontSize("12px").Color("#abbcd2"),
						),
				),
		)
}

func (a *builderApp) handleCompare(action string) bool {
	switch action {
	case "compare:open":
		a.compare.Preview = true
		a.notice = "Preview opened above the tab."
		return true
	case "compare:preview-close":
		if !a.compare.Preview {
			return false
		}
		a.compare.Preview = false
		a.notice = "Preview closed."
		return true
	}
	name := action
	if rest, ok := strings.CutPrefix(name, "compare:"); ok {
		name = rest
	} else if action != "open" && !strings.HasPrefix(action, "mode:") {
		return false
	}
	if mode, ok := strings.CutPrefix(name, "mode:"); ok {
		switch mode {
		case "split", "builder", "html":
			a.compare.Mode = mode
			return true
		default:
			return false
		}
	}
	if name == "open" {
		return true
	}
	return false
}

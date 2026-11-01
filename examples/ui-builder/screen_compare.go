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
	return `ui.Div().
    Display(ui.DisplayFlex).
    FlexDirection(ui.FlexDirectionColumn).
    Gap(ui.Px(10)).
    Padding(ui.Px(16)).
    BackgroundColor(ui.Hex("#202d42")).
    BorderRadius(ui.Px(11)).
    Children(
        ui.Text("Deploy preview").FontSize(ui.Px(13)).Color(ui.Hex("#b5c7df")),
        ui.P().Text("1,248 rows").Margin(ui.Px(0)).FontSize(ui.Px(29)).FontWeight(ui.FontWeightBold),
        ui.Button("Open preview").Attr("data-action", "compare:open").
            AlignSelf(ui.RawValue("flex-start")).
            BackgroundColor(ui.Hex("#4677ed")).Color(ui.Hex("#ffffff")).
            Padding(ui.Edges(ui.Px(9), ui.Px(12))).BorderRadius(ui.Px(8)),
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
			BackgroundColor(ui.RawValue(bg)).Color(ui.Hex("#f1f5fd")).
			Border(ui.RawValue("1px solid #40546f")).
			Padding(ui.Edges(ui.Px(9), ui.Px(12))).BorderRadius(ui.Px(8)).FontWeight(ui.FontWeightBold))
	}
	content := ui.Div().Display(ui.DisplayFlex).MinWidth(ui.Px(0)).Gap(ui.Px(12))
	switch mode {
	case "builder":
		content.Children(a.compareLivePane())
	case "html":
		content.Children(a.compareCodePane())
	default:
		content.Children(a.compareLivePane(), a.compareCodePane())
	}
	return ui.Section().Flex(ui.RawValue("1")).MinWidth(ui.Px(0)).MinHeight(ui.Px(0)).
		Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(14)).
		Children(
			ui.Div().Children(
				a.panelHeading("Builder vs HTML"),
				ui.P().Text("The same three-field card, built with Go and shown as escaped source.").
					Margin(ui.RawValue("7px 0 0")).Color(ui.Hex("#a9bbd1")).FontSize(ui.Px(12)),
			),
			ui.Div().Display(ui.DisplayFlex).Gap(ui.Px(8)).Children(toggles...),
			content,
			ui.P().Text("Render emits plain HTML to Blinkless, no JS/DOM.").
				Margin(ui.Px(0)).FontSize(ui.Px(12)).Color(ui.Hex("#abbcd2")),
		)
}

func (a *builderApp) compareLivePane() *ui.Element {
	return ui.Div().Flex(ui.RawValue("1")).MinWidth(ui.Px(0)).
		Padding(ui.Px(16)).BorderRadius(ui.Px(11)).BackgroundColor(ui.Hex("#202d42")).
		Children(
			ui.Text("Live builder card").FontWeight(ui.FontWeightBold).Color(ui.Hex("#f4f7ff")),
			ui.Div().MarginTop(ui.Px(10)).Children(a.compareCard()),
		)
}

func (a *builderApp) compareCodePane() *ui.Element {
	return ui.Div().Flex(ui.RawValue("1")).MinWidth(ui.Px(0)).
		Padding(ui.Px(16)).BorderRadius(ui.Px(11)).BackgroundColor(ui.Hex("#202d42")).
		Children(
			ui.Text("Same card as HTML source").FontWeight(ui.FontWeightBold).Color(ui.Hex("#f4f7ff")),
			ui.Tag("pre").MarginTop(ui.Px(10)).
				Padding(ui.Px(12)).BorderRadius(ui.Px(8)).
				BackgroundColor(ui.Hex("#111827")).Color(ui.Hex("#cbd8eb")).FontSize(ui.Px(12)).
				FontFamily(ui.RawValue("monospace")).
				WhiteSpace(ui.WhiteSpacePreWrap).OverflowWrap(ui.OverflowWrapBreakWord).
				Border(ui.RawValue("1px solid #3a4b62")).
				Children(
					ui.Tag("code").Text(builderCardSnippet()),
				),
		)
}

func (a *builderApp) compareCard() *ui.Element {
	return ui.Div().Padding(ui.Px(16)).BackgroundColor(ui.Hex("#202d42")).BorderRadius(ui.Px(11)).
		Border(ui.RawValue("1px solid #3e526e")).
		Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(10)).
		Children(
			ui.Text("Deploy preview").FontSize(ui.Px(13)).Color(ui.Hex("#b5c7df")),
			ui.P().Text("1,248 rows").Margin(ui.Px(0)).
				FontSize(ui.Px(29)).FontWeight(ui.FontWeightBold).Color(ui.Hex("#f4f7ff")),
			ui.Button("Open preview").ID("compare-open").Attr("data-action", "compare:open").
				AlignSelf(ui.RawValue("flex-start")).
				BackgroundColor(ui.Hex("#4677ed")).Color(ui.Hex("#ffffff")).
				Padding(ui.Edges(ui.Px(9), ui.Px(12))).BorderRadius(ui.Px(8)).FontWeight(ui.FontWeightBold),
		)
}

// comparePreviewOverlay floats the live card above the tab on a backdrop,
// or nil when closed. Mounted at the screen root like the dialog overlay.
func (a *builderApp) comparePreviewOverlay() *ui.Element {
	if a.tab != "compare" || !a.compare.Preview {
		return nil
	}
	return ui.Div().ID("compare-preview-backdrop").Attr("data-action", "compare:preview-close").
		Position(ui.PositionAbsolute).Top(ui.Px(0)).Left(ui.Px(0)).
		Right(ui.Px(0)).Bottom(ui.Px(0)).ZIndex(ui.Number(50)).
		Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).JustifyContent(ui.JustifyContentCenter).
		BackgroundColor(ui.RawValue("rgba(0, 0, 0, 0.55)")).
		Children(
			ui.Div().ID("compare-preview-panel").
				Width(ui.Px(480)).
				BackgroundColor(ui.Hex("#1c2738")).
				Border(ui.RawValue("1px solid #39465a")).
				BorderRadius(ui.Px(12)).Overflow(ui.OverflowHidden).
				Children(
					ui.Div().Display(ui.DisplayFlex).JustifyContent(ui.JustifyContentSpaceBetween).
						AlignItems(ui.AlignItemsCenter).Padding(ui.Edges(ui.Px(16), ui.Px(20))).
						BorderBottom(ui.RawValue("1px solid #39465a")).
						Children(
							ui.Text("Card preview").FontSize(ui.Px(16)).FontWeight(ui.FontWeightBold).Color(ui.Hex("#f4f7ff")),
							ui.Button("×").ID("compare-preview-x").Attr("type", "button").Attr("data-action", "compare:preview-close").
								BackgroundColor(ui.Hex("#2b3950")).Color(ui.Hex("#dce6f6")).Border(ui.RawValue("1px solid #42536b")).
								Padding(ui.Edges(ui.Px(6), ui.Px(12))).BorderRadius(ui.Px(8)).FontWeight(ui.FontWeightBold),
						),
					ui.Div().Padding(ui.Edges(ui.Px(18), ui.Px(20), ui.Px(20))).
						Children(
							a.compareCard(),
							ui.P().Text("The live card, floating above the tab. Backdrop click or × closes it.").
								Margin(ui.RawValue("12px 0 0")).FontSize(ui.Px(12)).Color(ui.Hex("#abbcd2")),
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

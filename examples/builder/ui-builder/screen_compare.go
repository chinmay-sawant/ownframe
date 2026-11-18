package main

import (
	"strings"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/theme"
)

type compareState struct {
	Mode    string
	Preview bool
}

func defaultCompareState() compareState { return compareState{Mode: "split"} }

func builderCardSnippet() string {
	return `chadcn.Card(
    ui.Text("Deploy preview").Color(theme.Muted()),
    ui.P().Text("1,248 rows").FontSize(ui.Px(29)),
    chadcn.Button("Open preview", chadcn.ButtonOptions{
        Variant: chadcn.ButtonFilled,
    }).ID("compare-open").Attr("data-action", "compare:open"),
)`
}

func (a *builderApp) comparePanel() *ui.Element {
	modes := []*ui.Element{}
	for _, mode := range []string{"split", "builder", "html"} {
		modes = append(modes, widgetsBtn(mode, "compare:mode:"+mode, a.compare.Mode == mode))
	}
	content := ui.Div().Display(ui.DisplayFlex).Gap(ui.Px(12)).MinWidth(ui.Px(0))
	switch a.compare.Mode {
	case "builder":
		content.Children(a.compareLivePane())
	case "html":
		content.Children(a.compareCodePane())
	default:
		content.Children(a.compareLivePane(), a.compareCodePane())
	}
	return ui.Div().Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(16)).Children(
		a.panelHeading("Builder vs HTML"),
		ui.P().Text("A shared component, its Go composition, and the HTML it generates.").Margin(ui.Px(0)).Color(theme.Muted()).FontSize(ui.Px(13)),
		ui.Div().Display(ui.DisplayFlex).Gap(ui.Px(8)).Children(modes...), content,
	)
}

func (a *builderApp) compareLivePane() *ui.Element {
	return chadcn.Card(ui.Text("Live component").FontWeight(ui.FontWeightBold), a.compareCard("compare-open").MarginTop(ui.Px(12))).
		Flex(ui.Number(1)).MinWidth(ui.Px(0))
}

func (a *builderApp) compareCodePane() *ui.Element {
	source, title := builderCardSnippet(), "Go composition"
	if a.compare.Mode == "html" {
		generated, err := ui.Render(a.compareCard("compare-source"))
		if err != nil {
			return chadcn.Notice("Source error", err.Error())
		}
		source, title = generated, "Generated HTML"
	}
	return chadcn.Card(ui.Text(title).FontWeight(ui.FontWeightBold),
		ui.Tag("pre").Text(source).FontFamily(ui.FontFamilyMonospace).FontSize(ui.Px(12)).
			WhiteSpace(ui.WhiteSpacePreWrap).OverflowWrap(ui.OverflowWrapBreakWord).MarginTop(ui.Px(12)).Color(theme.Text()),
	).Flex(ui.Number(1)).MinWidth(ui.Px(0))
}

func (a *builderApp) compareCard(id string) *ui.Element {
	return chadcn.Card(
		ui.Text("Deploy preview").FontSize(ui.Px(13)).Color(theme.Muted()),
		ui.P().Text("1,248 rows").Margin(ui.Px(0)).FontSize(ui.Px(29)).FontWeight(ui.FontWeightBold),
		chadcn.Button("Open preview", chadcn.ButtonOptions{Variant: chadcn.ButtonFilled}).
			ID(id).Attr("data-action", "compare:open").AlignSelf(ui.AlignSelfFlexStart),
	).Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(10))
}

func (a *builderApp) comparePreviewOverlay() *ui.Element {
	if a.tab != "compare" || !a.compare.Preview {
		return nil
	}
	return chadcn.Dialog(chadcn.DialogOptions{
		Interactions: a.interactions,
		BackdropID:   "compare-preview-backdrop", PanelID: "compare-preview-panel", Label: "Card preview", CloseAction: "compare:preview-close", Width: ui.Px(480),
	},
		ui.Div().Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).JustifyContent(ui.JustifyContentSpaceBetween).
			Padding(theme.Spacing()).BorderBottom(theme.Border()).Children(ui.Text("Card preview").FontWeight(ui.FontWeightBold),
			chadcn.Button("×", chadcn.ButtonOptions{}).ID("compare-preview-x").Attr("aria-label", "Close card preview").Attr("data-action", "compare:preview-close")),
		a.compareCard("compare-preview-open").Margin(theme.Spacing()),
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

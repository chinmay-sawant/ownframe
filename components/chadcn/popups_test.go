package chadcn_test

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

func renderPopupTree(t *testing.T, root *ui.Element) string {
	t.Helper()
	source, err := ui.Render(root)
	if err != nil {
		t.Fatalf("render popup: %v", err)
	}
	return source
}

func TestPopoverOpenRendersAnchoredPanel(t *testing.T) {
	t.Parallel()
	source := renderPopupTree(t, chadcn.Popover(chadcn.PopoverOptions{
		ID:      "dimensions",
		Open:    true,
		Side:    chadcn.PopupTop,
		Align:   chadcn.PopoverAlignEnd,
		Trigger: chadcn.Button("Open popover", chadcn.ButtonOptions{}).ID("dimensions-trigger"),
	},
		ui.Text("Dimensions"),
		ui.Text("Set the dimensions for the layer."),
	))
	for _, want := range []string{
		`class="of-popover of-popover--top of-popover--align-end"`,
		`id="dimensions"`,
		`class="of-popover-trigger"`,
		`class="of-popover-anchor"`,
		`id="dimensions-trigger"`,
		`aria-expanded="true"`,
		`aria-haspopup="dialog"`,
		`aria-controls="dimensions-content"`,
		`class="of-popover-content"`,
		`id="dimensions-content"`,
		`role="dialog"`,
		`>Dimensions<`,
		`>Set the dimensions for the layer.<`,
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("popover markup missing %q\n%s", want, source)
		}
	}
	anchorAt := strings.Index(source, `class="of-popover-anchor"`)
	triggerAt := strings.Index(source, `class="of-popover-trigger"`)
	if anchorAt < 0 || triggerAt < 0 || anchorAt > triggerAt {
		t.Fatalf("top popover anchor must precede the trigger\n%s", source)
	}
}

func TestPopoverClosedKeepsTriggerOnly(t *testing.T) {
	t.Parallel()
	source := renderPopupTree(t, chadcn.Popover(chadcn.PopoverOptions{
		ID:      "filters",
		Trigger: chadcn.Button("Filters", chadcn.ButtonOptions{}).ID("filters-trigger"),
	}, ui.Text("Panel body")))
	for _, want := range []string{
		`class="of-popover of-popover--bottom of-popover--align-start"`,
		`aria-expanded="false"`,
		`id="filters-trigger"`,
		`>Filters<`,
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("closed popover markup missing %q\n%s", want, source)
		}
	}
	for _, unwanted := range []string{"of-popover-content", "of-popover-anchor", "aria-controls", "Panel body"} {
		if strings.Contains(source, unwanted) {
			t.Fatalf("closed popover rendered %q\n%s", unwanted, source)
		}
	}
}

func TestPopoverWithoutIDOmitsRelations(t *testing.T) {
	t.Parallel()
	source := renderPopupTree(t, chadcn.Popover(chadcn.PopoverOptions{
		Open:    true,
		Trigger: chadcn.Button("More", chadcn.ButtonOptions{}),
	}, ui.Text("Body")))
	if strings.Contains(source, `id=""`) {
		t.Fatalf("popover invented an empty ID\n%s", source)
	}
	if strings.Contains(source, "aria-controls") {
		t.Fatalf("popover without an ID rendered aria-controls\n%s", source)
	}
	if !strings.Contains(source, `class="of-popover-content"`) {
		t.Fatalf("open popover without an ID lost its content\n%s", source)
	}
	triggerAt := strings.Index(source, `class="of-popover-trigger"`)
	anchorAt := strings.Index(source, `class="of-popover-anchor"`)
	if triggerAt < 0 || anchorAt < 0 || triggerAt > anchorAt {
		t.Fatalf("bottom popover anchor must follow the trigger\n%s", source)
	}
}

func TestPopoverWiresTriggerAction(t *testing.T) {
	t.Parallel()
	source := renderPopupTree(t, chadcn.Popover(chadcn.PopoverOptions{
		ID: "help",
		OnOpenChange: func(context.Context, bool) error {
			return nil
		},
		Trigger: chadcn.Button("Help", chadcn.ButtonOptions{}),
	}))
	if !strings.Contains(source, "data-action=") {
		t.Fatalf("popover trigger has no action\n%s", source)
	}
	if !strings.Contains(source, `aria-expanded="false"`) {
		t.Fatalf("popover trigger lost its closed state\n%s", source)
	}
	if strings.Contains(source, "of-popover-content") {
		t.Fatalf("closed popover rendered its panel\n%s", source)
	}
}

func TestTooltipOpenDescribesTrigger(t *testing.T) {
	t.Parallel()
	source := renderPopupTree(t, chadcn.Tooltip(chadcn.TooltipOptions{
		ID:      "library-tip",
		Open:    true,
		Content: "Add to library",
		Side:    chadcn.PopupTop,
	}, chadcn.Button("Hover", chadcn.ButtonOptions{}).ID("library")))
	for _, want := range []string{
		`class="of-tooltip of-tooltip--top"`,
		`id="library-tip"`,
		`class="of-tooltip-trigger"`,
		`class="of-tooltip-anchor"`,
		`class="of-tooltip-content"`,
		`id="library-tip-content"`,
		`role="tooltip"`,
		`aria-describedby="library-tip-content"`,
		`>Add to library<`,
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("tooltip markup missing %q\n%s", want, source)
		}
	}
}

func TestTooltipClosedRendersTriggerOnly(t *testing.T) {
	t.Parallel()
	source := renderPopupTree(t, chadcn.Tooltip(chadcn.TooltipOptions{
		ID:      "library-tip",
		Content: "Add to library",
	}, chadcn.Button("Hover", chadcn.ButtonOptions{}).ID("library")))
	if !strings.Contains(source, `class="of-tooltip of-tooltip--bottom"`) {
		t.Fatalf("closed tooltip lost its wrapper\n%s", source)
	}
	if !strings.Contains(source, ">Hover<") {
		t.Fatalf("closed tooltip lost its trigger\n%s", source)
	}
	for _, unwanted := range []string{"of-tooltip-content", "of-tooltip-anchor", `role="tooltip"`, "aria-describedby", "Add to library"} {
		if strings.Contains(source, unwanted) {
			t.Fatalf("closed tooltip rendered %q\n%s", unwanted, source)
		}
	}
}

func TestHoverCardOpenRendersContent(t *testing.T) {
	t.Parallel()
	source := renderPopupTree(t, chadcn.HoverCard(chadcn.HoverCardOptions{
		ID:   "nextjs-card",
		Open: true,
	}, ui.Text("Hover Here"),
		ui.Div().Children(
			ui.Text("@nextjs"),
			ui.Text("The React Framework, created and maintained by @vercel."),
		),
	))
	for _, want := range []string{
		`class="of-hover-card of-hover-card--bottom"`,
		`id="nextjs-card"`,
		`class="of-hover-card-trigger"`,
		`class="of-hover-card-anchor"`,
		`class="of-hover-card-content"`,
		`id="nextjs-card-content"`,
		`role="dialog"`,
		`>Hover Here<`,
		`>@nextjs<`,
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("hover card markup missing %q\n%s", want, source)
		}
	}
}

func TestHoverCardClosedRendersTriggerOnly(t *testing.T) {
	t.Parallel()
	source := renderPopupTree(t, chadcn.HoverCard(chadcn.HoverCardOptions{
		ID:   "nextjs-card",
		Side: chadcn.PopupRight,
	}, ui.Text("Hover Here"), ui.Text("Card body")))
	if !strings.Contains(source, `class="of-hover-card of-hover-card--right"`) {
		t.Fatalf("closed hover card lost its wrapper\n%s", source)
	}
	for _, unwanted := range []string{"of-hover-card-content", "of-hover-card-anchor", `role="dialog"`, "Card body"} {
		if strings.Contains(source, unwanted) {
			t.Fatalf("closed hover card rendered %q\n%s", unwanted, source)
		}
	}
}

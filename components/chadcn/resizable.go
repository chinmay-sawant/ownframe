package chadcn

import (
	"context"
	"math"
	"strconv"

	"github.com/chinmay-sawant/ownframe/ui"
)

const (
	defaultResizableRatio = 0.5
	resizableMinRatio     = 0.1
	resizableMaxRatio     = 0.9
	resizablePercent      = 100
)

// ResizableOrientation selects the split axis. The zero value places the panes
// left and right of a vertical divider.
type ResizableOrientation uint8

const (
	// ResizableRow splits horizontally: First on the left, Second on the right.
	ResizableRow ResizableOrientation = iota
	// ResizableColumn splits vertically: First above, Second below.
	ResizableColumn
)

// ResizableOptions describes a two-pane split. Ratio is the first pane's share
// of the space; zero, negative, or one and above fall back to one half. The
// divider id is ID+"-divider". The host maps the pointer with ResizableRatioAt
// and calls OnChange with that ratio, then rebuilds.
type ResizableOptions struct {
	ID          string
	Ratio       float64
	First       *ui.Element
	Second      *ui.Element
	Orientation ResizableOrientation
	OnChange    func(context.Context, float64) error
}

// Resizable builds a flex split with a focusable, labeled divider. The host
// sets the root size, routes pointer drag and arrow keys on the divider id,
// and passes the new ratio back in on the next build. A click on the divider
// carries no pointer, so Resizable does not call OnChange itself.
func Resizable(options ResizableOptions) *ui.Element {
	ratio := options.Ratio
	if !(ratio > 0 && ratio < 1) {
		ratio = defaultResizableRatio
	}
	vertical := options.Orientation == ResizableColumn
	axis, dividerAxis := "horizontal", "vertical"
	if vertical {
		axis, dividerAxis = "vertical", "horizontal"
	}
	first := ui.Div().Class(classes("of-resizable-pane", "of-resizable-pane--first")).
		FlexGrow(ui.Number(ratio)).Children(options.First)
	second := ui.Div().Class(classes("of-resizable-pane", "of-resizable-pane--second")).
		FlexGrow(ui.Number(1 - ratio)).Children(options.Second)
	divider := ui.Div().Class("of-resizable-divider").Attr("role", "separator").
		Attr("aria-orientation", dividerAxis).Attr("aria-label", "Resize panes").
		Attr("aria-valuemin", "0").Attr("aria-valuemax", "100").
		Attr("aria-valuenow", strconv.Itoa(int(math.Round(ratio*resizablePercent)))).
		Attr("tabindex", "0").Children(
		ui.Div().Class("of-resizable-grip").Attr("aria-hidden", "true"),
	)
	if dividerID := partID(options.ID, "-divider"); dividerID != "" {
		divider.ID(dividerID)
	}
	root := ui.Div().Class(classes("of-scope", "of-resizable", "of-resizable--"+axis))
	if options.ID != "" {
		root.ID(options.ID)
	}
	if firstID := partID(options.ID, "-first"); firstID != "" {
		first.ID(firstID)
	}
	if secondID := partID(options.ID, "-second"); secondID != "" {
		second.ID(secondID)
	}
	return root.Children(first, divider, second)
}

// ResizableRatioAt maps pointerX onto the first pane's share of a split whose
// drag axis starts at origin and runs span CSS pixels. A row passes the
// pointer's X, the root's left, and its width. A column passes the pointer's
// Y, the root's top, and its height in those same positions. The result is
// clamped to [0.1, 0.9]. A non-finite input or a non-positive span returns
// one half.
func ResizableRatioAt(pointerX, origin, span float64) float64 {
	if !finite(pointerX) || !finite(origin) || !finite(span) || span <= 0 {
		return defaultResizableRatio
	}
	ratio := (pointerX - origin) / span
	if !finite(ratio) || ratio < resizableMinRatio {
		return resizableMinRatio
	}
	if ratio > resizableMaxRatio {
		return resizableMaxRatio
	}
	return ratio
}

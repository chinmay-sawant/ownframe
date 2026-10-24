package render

import "github.com/chinmay-sawant/blinkless/layout"

// Operation kinds, re-stated so a caller can switch on them without importing
// the engine. A type alias does not carry the constants across, so these
// mirror the layout package's Display kinds one for one.
const (
	// OpNoop is a deactivated operation. The clipper writes it when overflow
	// removes an operation from the page, and it paints nothing. Skip it; its
	// value sits outside the sequence below.
	OpNoop = layout.DisplayOpNoop

	// OpUnknown is the zero kind. The engine emits it only as the boundary
	// marker of a blend or isolation group, which paints nothing. Treat any
	// kind outside this list as inert, so a kind added later cannot be
	// mistaken for a fill.
	OpUnknown = layout.DisplayOpUnknown

	// OpFillRect is a filled rectangle, with per-corner radii for a rounded
	// or elliptical box.
	OpFillRect = layout.DisplayOpFillRect

	// OpStrokeRect is a stroked rectangle. StrokeMask selects sides when it is
	// non-zero; zero means the complete rounded rectangle.
	OpStrokeRect = layout.DisplayOpStrokeRect

	// OpLine is one stroked segment, Width points wide.
	OpLine = layout.DisplayOpLine

	// OpText is one shaped text run. Y is the baseline, not the top.
	OpText = layout.DisplayOpText

	// OpImage is one image with its encoded payload and pixel bounds. Read it
	// with ImageBytes.
	OpImage = layout.DisplayOpImage

	// OpLinkURI is a link target. It paints nothing itself.
	OpLinkURI = layout.DisplayOpLinkURI

	// OpBullet is a generated list marker. Y is the baseline.
	OpBullet = layout.DisplayOpBullet

	// OpGridRun is one table row's collapsed border grid, carried as ordered
	// segments. Replay the segments in order.
	OpGridRun = layout.DisplayOpGridRun

	// OpPath is one corner-shape path for a non-round box. Read the command
	// list with CornerPathValue: Stroke is false for a background fill and
	// true for a border stroke painted with Width. A round box never uses
	// this kind; it keeps the fill and stroke rect radii.
	OpPath = layout.DisplayOpPath
)

// Kind is the operation kind type.
type Kind = layout.DisplayKind

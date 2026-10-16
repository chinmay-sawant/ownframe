package masking

import "github.com/chinmay-sawant/blinkless/layout"

// GridLine expands segment idx as a standalone line. Geometry, stroke width,
// color, and LineInset come from the segment; every other owner field (ID,
// Alpha and opacity, transform, Fixed, Pinned, ZIndex) propagates by copy and
// Grid is cleared so the line paints alone.

func GridLine(op *layout.DisplayOp, idx int) layout.DisplayOp {
	if op == nil || op.Grid == nil || idx < 0 || idx >= len(op.Grid.Segs) {
		return layout.DisplayOp{Kind: layout.DisplayOpLine}
	}
	seg := &op.Grid.Segs[idx]
	line := *op
	line.Kind = layout.DisplayOpLine
	line.X, line.Y, line.W, line.H = seg.X, seg.Y, seg.W, seg.H
	line.Width, line.LineInset = seg.Width, seg.LineInset
	line.R, line.G, line.B = seg.R, seg.G, seg.B
	line.Grid = nil
	return line
}

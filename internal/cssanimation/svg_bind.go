package cssanimation

import (
	"fmt"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/animation"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// SVGPaintMask marks a bound player whose only animated property is inline-SVG
// paint. An inline SVG is rasterized by the engine, so its paint is not a
// display field: the player targets the SVG root's image operation but writes
// no operation value. The bit sits above every cssmotion write field, so
// cssmotion.Read and cssmotion.Write leave the operation at its captured paint
// and the page restyles the raster through layout.RestyleInlineSVG instead.
const SVGPaintMask cssmotion.Mask = 1 << 8

// svgPaintClass is the pseudo target class for an animation whose declarations
// are all inline-SVG paint. The kind is unused: the class selects the element's
// image operation directly instead of going through cssmotion.Select.
func svgPaintClass() paintClass {
	return paintClass{kind: cssmotion.KindFill, mask: SVGPaintMask}
}

// declaresSVGPaint reports whether any keyframe declares a property that only
// the inline-SVG restyle path can apply. A rule that also declares a paint class
// keeps the SVG sample on that player; this predicate only decides whether an
// otherwise classless animation needs a player of its own.
func declaresSVGPaint(source []layout.Keyframe) bool {
	for i := range source {
		for _, declaration := range source[i].Declarations {
			if isSVGPaintProperty(declaration.Property) {
				return true
			}
		}
	}

	return false
}

// svgImageIndices returns the operation index of the inline SVG's raster. The
// SVG root owns exactly one image operation; a background or border the element
// declares paints before it, and the engine bakes the last image of the box as
// the raster. An empty slice means the element did not paint an SVG. A shared
// index resolves the lookup without scanning the display again.
func svgImageIndices(d *layout.Display, box layout.Box, boxIndex int, idx *cssmotion.Index) []int {
	if d == nil {
		return nil
	}
	if idx != nil {
		return idx.ImageIndex(box, boxIndex)
	}
	if boxIndex >= 0 && len(d.OpOwners) == len(d.Ops) {
		last := -1
		for i := range d.Ops {
			if d.OpOwners[i] == boxIndex && d.Ops[i].Kind == layout.DisplayOpImage {
				last = i
			}
		}
		if last >= 0 {
			return []int{last}
		}
	}

	// No trusted owner index, or the owner map is missing: match the box's
	// operation ids. A box that predates operation ids falls back to the same
	// geometric containment the paint classes use.
	ids := make(map[uint64]struct{}, len(box.OperationIDs))
	for _, id := range box.OperationIDs {
		ids[id] = struct{}{}
	}
	last := -1
	for i := range d.Ops {
		if d.Ops[i].Kind != layout.DisplayOpImage {
			continue
		}
		if len(ids) > 0 {
			if _, ok := ids[d.Ops[i].ID]; !ok {
				continue
			}
		} else if !cssmotion.Contains(d, box, d.Ops[i]) {
			continue
		}
		last = i
	}
	if last >= 0 {
		return []int{last}
	}

	return nil
}

// reportSVGPaintDeferred counts the declarations a paint server left at the
// element's base paint. The sample carries the reason on every frame; the bind
// records it once per element, next to the other fidelity gaps.
func reportSVGPaintDeferred(report *gapList, box layout.Box, frames []animation.Keyframe) {
	for i := range frames {
		if frames[i].SVG.Deferred == "" {
			continue
		}
		report.add(fmt.Sprintf("box %q: %s", box.ID, frames[i].SVG.Deferred))

		return
	}
}

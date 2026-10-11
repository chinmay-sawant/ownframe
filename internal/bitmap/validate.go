package bitmap

import (
	"fmt"

	"github.com/chinmay-sawant/blinkless/layout"
)

const maxPixels = 16 * 1024 * 1024

func validate(d *layout.Display) error {
	if d.Width > maxPixels/d.Height {
		return fmt.Errorf("bitmap: canvas exceeds %d pixels", maxPixels)
	}
	maxDepth := 0
	for _, op := range d.Ops {
		depth := 0
		for g := op.Group(); g != nil; g = g.Parent {
			depth++
			if depth > 7 {
				return fmt.Errorf("bitmap: blend nesting exceeds 7")
			}
		}
		maxDepth = max(maxDepth, depth)
		if !blendSupported(op.BlendModeName()) {
			return fmt.Errorf("bitmap: unsupported blend mode %q", op.BlendModeName())
		}
		if group := op.Group(); group != nil && !blendSupported(group.Mode) {
			return fmt.Errorf("bitmap: unsupported group mode %q", group.Mode)
		}
		if op.GroupBoundary() != 0 {
			continue
		}
		switch op.Kind {
		case layout.DisplayOpNoop, layout.DisplayOpLinkURI:
		case layout.DisplayOpFillRect, layout.DisplayOpLine:
		case layout.DisplayOpStrokeRect:
			if op.StrokeMask & ^uint8(15) != 0 {
				return fmt.Errorf("bitmap: unknown stroke mask")
			}
		case layout.DisplayOpGridRun:
			if op.Grid == nil {
				return fmt.Errorf("bitmap: missing grid")
			}
		case layout.DisplayOpImage:
			if data, _, _ := op.ImageBytes(); len(data) == 0 {
				return fmt.Errorf("bitmap: missing image")
			}
		case layout.DisplayOpText, layout.DisplayOpBullet:
			if op.Text != "" && op.Font == nil {
				return fmt.Errorf("bitmap: missing font")
			}
			if op.FontFeatures() != "" || op.TextAutospaceGap() != 0 {
				return fmt.Errorf("bitmap: unsupported text shaping options")
			}
		default:
			return fmt.Errorf("bitmap: unsupported operation %d", op.Kind)
		}
	}
	if d.Width > maxPixels/d.Height/(maxDepth+1) {
		return fmt.Errorf("bitmap: canvas and blend buffers exceed %d pixels", maxPixels)
	}
	return nil
}

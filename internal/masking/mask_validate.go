package masking

import (
	"bytes"
	"fmt"
	"image"
	"strings"

	"github.com/chinmay-sawant/blinkless/layout"
)

func Validate(mask *layout.MaskDescriptor) error {
	if len(mask.Layers) == 0 || len(mask.Layers) > 16 {
		return fmt.Errorf("mask: must have 1 to 16 layers")
	}
	for i, layer := range mask.Layers {
		if !strings.EqualFold(strings.TrimSpace(layer.Source), "none") && len(layer.ImageBytes) == 0 {
			return fmt.Errorf("mask: layer %d has no image data", i)
		}
		if !maskToken(layer.Mode, "", "alpha", "luminance", "match-source") ||
			!maskToken(layer.Composite, "", "add", "subtract", "intersect", "exclude") ||
			!maskToken(layer.Repeat, "", "no-repeat", "repeat", "repeat-x", "repeat-y") ||
			!maskToken(layer.Origin, "", "border-box") || !maskToken(layer.Clip, "", "border-box") {
			return fmt.Errorf("mask: unsupported layer %d values", i)
		}
		if _, _, err := maskSize(layer.Size); err != nil {
			return fmt.Errorf("mask: layer %d size: %w", i, err)
		}
		if _, _, err := maskPosition(layer.Position); err != nil {
			return fmt.Errorf("mask: layer %d position: %w", i, err)
		}
	}
	return nil
}

func validateSources(mask *layout.MaskDescriptor) error {
	for i, layer := range mask.Layers {
		if strings.EqualFold(strings.TrimSpace(layer.Source), "none") {
			continue
		}
		if _, _, err := image.DecodeConfig(bytes.NewReader(layer.ImageBytes)); err != nil {
			return fmt.Errorf("mask: layer %d image: %w", i, err)
		}
	}
	return nil
}

func maskToken(value string, allowed ...string) bool {
	for _, item := range allowed {
		if strings.EqualFold(strings.TrimSpace(value), item) {
			return true
		}
	}
	return false
}

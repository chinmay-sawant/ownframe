package masking

import (
	"bytes"
	"image"
	_ "image/jpeg"
	_ "image/png"

	"github.com/chinmay-sawant/blinkless/layout"
)

func ImageAlpha(op *layout.DisplayOp) float64 {
	if op == nil {
		return 0
	}
	return Clamp01(op.Opacity())
}

func DecodeImage(op *layout.DisplayOp) (img image.Image, alpha float64, skip bool, err error) {
	if op == nil {
		return nil, 0, true, nil
	}
	data, _, _ := op.ImageBytes()
	if len(data) == 0 {
		return nil, 0, true, nil
	}
	if cfg, _, err := image.DecodeConfig(bytes.NewReader(data)); err == nil {
		if cfg.Width <= 0 || cfg.Height <= 0 || PixelsTooBig(cfg.Width, cfg.Height) {
			return nil, 0, true, nil
		}
	}
	src, _, e := image.Decode(bytes.NewReader(data))
	if e != nil {
		return nil, 0, false, e
	}
	if src.Bounds().Empty() {
		return nil, 0, true, nil
	}
	if PixelsTooBig(src.Bounds().Dx(), src.Bounds().Dy()) {
		return nil, 0, true, nil
	}
	return src, Clamp01(op.Opacity()), false, nil
}

func SkipImageSize(op *layout.DisplayOp) bool {
	return op == nil || op.W <= 0 || op.H <= 0
}

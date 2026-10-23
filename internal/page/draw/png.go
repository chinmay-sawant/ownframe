package draw

import (
	"bytes"
	"context"
	"image"
	"image/png"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

// EncodePNG writes img as PNG bytes, or nil on failure.
func EncodePNG(img image.Image) []byte {
	if img == nil {
		return nil
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil
	}

	return buf.Bytes()
}

// Bytes encodes the frame. A cached encoding is copied. A display-list
// frame rasterizes once and keeps the bytes until the next paint change.
// A display that cannot be painted returns nil and does not fall through
// to source. An empty frame paints source at view when ctx is still live.
// The caller passes a non-nil ctx. The result is a copy.
func (f *Frame) Bytes(ctx context.Context, source string, view Viewport, state render.State) []byte {
	if b := f.CopyPNG(); b != nil {
		return b
	}

	img := f.picture(ctx, source, view, state)
	if img == nil {
		return nil
	}

	f.StorePNG(EncodePNG(img))

	return f.CopyPNG()
}

func (f *Frame) picture(ctx context.Context, source string, view Viewport, state render.State) image.Image {
	img := f.Img
	if img == nil && f.Display != nil {
		painted, err := Picture(f.Display)
		if err != nil || painted == nil {
			return nil
		}

		return painted
	}
	if img != nil {
		return img
	}
	if source == "" || ctx.Err() != nil {
		return nil
	}

	painted, err := PaintSource(ctx, source, view.Width, view.Height, state)
	if err != nil || painted == nil {
		return nil
	}

	return painted
}

package snip

import (
	"bytes"
	"image"
)

// setShot replaces the capture, selects the full frame, and recrops.
func (a *App) setShot(raw []byte) {
	if img, err := DecodePNG(raw); err == nil {
		a.shot = raw
		b := img.Bounds()
		a.sel = Rect{W: b.Dx(), H: b.Dy()}
		a.anns = ClearAnnotations(a.anns)
	}
	a.recrop()
}

// recrop rebuilds the preview. An empty selection keeps the capture.
func (a *App) recrop() {
	a.crop = a.shot
	if len(a.shot) == 0 {
		a.syncView()
		return
	}
	if out, err := CropPNG(a.shot, a.sel); err == nil {
		a.crop = out
	}
	if len(a.anns) > 0 {
		if img, err := DecodePNG(a.crop); err == nil {
			raw, err := EncodePNG(DrawAnnotations(img, a.anns))
			if err == nil {
				a.crop = raw
			}
		}
	}
	a.page.SetImage(shotSrc, a.crop)
	a.syncView()
}

// shotSize returns the capture pixel size, or zeros without one.
func (a *App) shotSize() (int, int) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(a.shot))
	if err != nil {
		return 0, 0
	}
	return cfg.Width, cfg.Height
}

// syncView refreshes the flag, size, selection, and count lines.
func (a *App) syncView() {
	a.view.HasImage = len(a.shot) > 0
	a.view.Size = describe(a.shot)
	a.view.Notes = len(a.anns)
	a.view.SelX, a.view.SelY = a.sel.X, a.sel.Y
	a.view.SelW, a.view.SelH = a.sel.W, a.sel.H
}

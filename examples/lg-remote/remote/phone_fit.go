package remote

import "math"

func (a *App) phoneFit() (scale, x, y float64) {
	w, h := a.page.Size()
	cw, ch := w, h
	if d := a.page.Display(); d != nil {
		cw, ch = d.Width, d.Height
	} else if img := a.page.Image(); img != nil {
		cw, ch = img.Bounds().Dx(), img.Bounds().Dy()
	}
	for _, b := range a.page.Boxes() {
		cw = max(cw, int(math.Ceil(b.X+b.W)))
		ch = max(ch, int(math.Ceil(b.Y+b.H)))
	}
	scale = math.Min(float64(w)/float64(cw), float64(h)/float64(ch))
	return scale, (float64(w) - float64(cw)*scale) / halfDivisor, (float64(h) - float64(ch)*scale) / halfDivisor
}

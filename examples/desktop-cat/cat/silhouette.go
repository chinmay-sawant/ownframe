package cat

import (
	"bytes"
	"image"
	"image/png"
)

type silhouette struct {
	pixels [200 * 200]bool
}

func makeSilhouette(data []byte) (*silhouette, error) {
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	s := &silhouette{}
	bounds := img.Bounds()
	for y := 0; y < 200; y++ {
		for x := 0; x < 200; x++ {
			p := image.Pt(bounds.Min.X+x*bounds.Dx()/200, bounds.Min.Y+y*bounds.Dy()/200)
			_, _, _, a := img.At(p.X, p.Y).RGBA()
			s.pixels[y*200+x] = a > 16*257
		}
	}
	return s, nil
}

// Draggable hit-tests opaque cat pixels, including its current bob offset.
func (c *Companion) Draggable(x, y int) bool {
	a := c.animation
	if a.shape == nil {
		return false
	}
	top := 142.0
	if a.image != nil && a.display == c.Page.Display() {
		top = a.image.Y / a.display.PointsPerPixel
	}
	px, py := x-100, int(float64(y)-top)
	return float64(y) >= top && px >= 0 && px < 200 && py >= 0 && py < 200 && a.shape.pixels[py*200+px]
}

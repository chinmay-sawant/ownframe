package snip

import (
	"image"
	"image/draw"
	"strings"
)

// Annotation is one serializable overlay for the template. Kind is
// "rect", "arrow", or "text". Label feeds the template only: text
// draws a placeholder box because the stdlib has no font engine, so
// the box reserves the label's space instead of rendering glyphs.
type Annotation struct {
	Kind       string
	X, Y, W, H int
	Label      string
	Color      string
}

// AddAnnotation appends a to anns.
func AddAnnotation(anns []Annotation, a Annotation) []Annotation {
	return append(anns, a)
}

// ClearAnnotations drops every annotation.
func ClearAnnotations(anns []Annotation) []Annotation { return nil }

// DrawAnnotations copies src, then paints each annotation with
// stdlib image/draw. Rects draw a 2px outline, arrows a line plus a
// head, and text draws an outline with a filled strip on top.
func DrawAnnotations(src image.Image, anns []Annotation) image.Image {
	b := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Bounds(), src, b.Min, draw.Src)
	for _, a := range anns {
		c := parseColor(a.Color)
		x, y := a.X-b.Min.X, a.Y-b.Min.Y
		switch strings.ToLower(a.Kind) {
		case "arrow":
			line(dst, x, y, x+a.W, y+a.H, c)
			head(dst, x, y, x+a.W, y+a.H, c)
		case "text":
			rectOutline(dst, x, y, a.W, a.H, c)
			textStrip(dst, x, y, a.W, c)
		default:
			rectOutline(dst, x, y, a.W, a.H, c)
		}
	}
	return dst
}

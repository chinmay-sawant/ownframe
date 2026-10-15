package web

import (
	"math"
	"net/url"
	"strconv"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// areas lists actionable boxes, inner ones first.
// An HTML image map uses the first matching area, which is the opposite
// of the engine's document order.
func areas(app host.Boxer) []shellArea {
	boxes := app.Boxes()
	out := make([]shellArea, 0)

	for i := len(boxes) - 1; i >= 0; i-- {
		b := boxes[i]
		if b.Action == "" {
			continue
		}

		q := url.Values{}
		q.Set("x", strconv.FormatFloat(b.X+b.W/2, 'f', -1, 64))
		q.Set("y", strconv.FormatFloat(b.Y+b.H/2, 'f', -1, 64))

		alt := b.ID
		if alt == "" {
			alt = b.Action
		}

		out = append(out, shellArea{
			Coords: rectCoords(b.X, b.Y, b.W, b.H),
			Href:   "/click?" + q.Encode(),
			Alt:    alt,
			X:      b.X + b.W/2,
			Y:      b.Y + b.H/2,
		})
	}

	return out
}

func rectCoords(x, y, w, h float64) string {
	x1 := int(math.Floor(x))
	y1 := int(math.Floor(y))
	x2 := int(math.Ceil(x + w))
	y2 := int(math.Ceil(y + h))

	return strconv.Itoa(x1) + "," + strconv.Itoa(y1) + "," + strconv.Itoa(x2) + "," + strconv.Itoa(y2)
}

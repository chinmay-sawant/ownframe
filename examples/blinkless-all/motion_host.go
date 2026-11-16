package main

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/chinmay-sawant/blinkless/css"
	"github.com/chinmay-sawant/blinkless/html"
	"github.com/chinmay-sawant/blinkless/layout"
)

// The engine moves the box (it folds the offset into the same matrix a
// transform gives). The host's job here is to read that matrix back from
// layout.Box.Transform and print where each box went, so the window shows the
// numbers and the test can check them.

// motionLane is one demo: a lane holding one 28px box whose motion is style.
type motionLane struct {
	ID    string
	Label string
	Style string
}

// motionLanes are the boxes of the tab. The path origin is the box's own
// top-left corner, so every box starts at the left edge of its lane.
//
//nolint:gochecknoglobals // read-only table of the demo lanes
var motionLanes = []motionLane{
	{"mp-dist-0", "offset-distance 0%", "offset-path: path('M 0 0 L 200 0'); offset-distance: 0%"},
	{"mp-dist-50", "offset-distance 50%", "offset-path: path('M 0 0 L 200 0'); offset-distance: 50%"},
	{"mp-dist-100", "offset-distance 100%", "offset-path: path('M 0 0 L 200 0'); offset-distance: 100%"},
	{"mp-rot-auto", "offset-rotate auto on a falling path", "offset-path: path('M 0 0 L 0 60'); offset-distance: 30px"},
	{"mp-rot-reverse", "offset-rotate reverse", "offset-path: path('M 0 0 L 0 60'); offset-distance: 30px; offset-rotate: reverse"},
	{"mp-rot-fixed", "offset-rotate 45deg", "offset-path: path('M 0 0 L 0 60'); offset-distance: 30px; offset-rotate: 45deg"},
	{"mp-anchor-center", "offset-anchor auto", "offset-path: path('M 100 10 L 200 10'); offset-rotate: 0deg"},
	{"mp-anchor-corner", "offset-anchor left top", "offset-path: path('M 100 10 L 200 10'); offset-rotate: 0deg; offset-anchor: left top"},
	{"mp-ray-origin", "offset-position 60px 30px with ray(90deg)", "offset-path: ray(90deg); offset-position: 60px 30px; offset-distance: 40px; offset-rotate: 0deg"},
	{"mp-short", "offset shorthand", "offset: path('M 0 0 L 200 0') 25% 0deg / left top"},
}

func motionLaneHTML(l motionLane) string {
	return fmt.Sprintf(`<div class="lane"><div id="%s" class="dot" style="%s">m</div><span class="lbl">%s</span></div>`,
		l.ID, l.Style, l.Label)
}

const motionCSS = `
.lane{position:relative;height:50px;margin:20px 0 0 0;border-bottom:1px dashed #d6d1c4}
.rd{font-family:monospace;font-size:12px;color:#374151;margin:2px 0 0 0}
.dot{width:28px;height:28px;background:#2563eb;color:#fff;text-align:center;line-height:28px}
.lbl{position:absolute;left:260px;top:4px;font-size:12px;color:#4b5563}
`

// motionSample is the lanes laid out alone, so the host can read the boxes.
func motionSample() string {
	var b strings.Builder

	b.WriteString(`<!DOCTYPE html><html><head><style>` + motionCSS + `</style></head><body style="margin:0">`)

	for _, l := range motionLanes {
		b.WriteString(motionLaneHTML(l))
	}

	b.WriteString(`</body></html>`)

	return b.String()
}

// MotionReading is where the engine put one box.
type MotionReading struct {
	ID       string
	Moved    bool
	CenterX  float64
	CenterY  float64
	TurnDeg  float64
	StartedX float64
	StartedY float64
}

// readMotion lays the sample out and reads each box's own matrix. Positions are
// CSS pixels in the sample page.
func readMotion(ctx context.Context) (map[string]MotionReading, error) {
	doc, err := html.Parse([]byte(motionSample()))
	if err != nil {
		return nil, err
	}

	styled, err := css.Apply(ctx, doc, css.Options{WidthPx: motionPageW, HeightPx: motionPageH, Media: "screen"})
	if err != nil {
		return nil, err
	}

	display, err := layout.DisplayList(ctx, styled)
	if err != nil {
		return nil, err
	}

	out := make(map[string]MotionReading, len(display.Boxes))
	k := display.PointsPerPixel

	for _, b := range display.Boxes {
		if !strings.HasPrefix(b.ID, "mp-") {
			continue
		}

		cx, cy := b.X+b.W*half, b.Y+b.H*half
		x, y := b.Transform.Own.Apply(cx*k, cy*k)
		out[b.ID] = MotionReading{
			ID: b.ID, Moved: b.Transform.HasMatrix,
			CenterX: x / k, CenterY: y / k,
			TurnDeg:  math.Round(math.Atan2(b.Transform.Own.B, b.Transform.Own.A) * 180 / math.Pi),
			StartedX: cx, StartedY: cy,
		}
	}

	return out, nil
}

// motionDemoHTML draws the lanes with the given ids, each followed by its
// readout line.
func motionDemoHTML(ids ...string) string {
	r, err := readMotion(context.Background())
	if err != nil {
		return `<p>could not read the layout: ` + err.Error() + `</p>`
	}

	var b strings.Builder

	for _, id := range ids {
		for _, l := range motionLanes {
			if l.ID != id {
				continue
			}

			b.WriteString(motionLaneHTML(l))

			if m, ok := r[id]; ok {
				fmt.Fprintf(&b, `<div id="%s-read" class="rd">%s: center moves by (%.0f, %.0f) px, turns %.0f deg</div>`,
					id, id, m.CenterX-m.StartedX, m.CenterY-m.StartedY, m.TurnDeg)
			}
		}
	}

	return b.String()
}

// Sizes of the motion demo page, in CSS pixels, and one half.
const (
	motionPageW = 600
	motionPageH = 800
	half        = 0.5
)

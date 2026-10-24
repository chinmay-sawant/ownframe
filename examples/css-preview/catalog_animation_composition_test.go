package main

import (
	"math"
	"testing"
)

// TestCatalogAnimationCompositionChangesOwnFrameSample drives the page's own
// driver. Two route effects of 100px each at a quarter of the duration:
// replace keeps the later 25px, while add and accumulate compose both.
func TestCatalogAnimationCompositionChangesOwnFrameSample(t *testing.T) {
	base := `animation-name:route,route;animation-duration:1000ms;` +
		`animation-delay:-250ms;animation-fill-mode:both;` +
		`animation-timing-function:linear;animation-play-state:paused;`
	cases := []struct {
		mode string
		want float64
	}{
		{"replace,replace", 25},
		{"add,add", 50},
		{"accumulate,accumulate", 50},
	}
	for _, c := range cases {
		html := `<style>@keyframes route{from{transform:translateX(0px)}to{transform:translateX(100px)}}` +
			`#target{width:40px;height:20px;background:#123456;` + base +
			`animation-composition:` + c.mode + `;}</style><div id="target"></div>`
		p := motionPage(t, html, "")
		if err := p.Testing().Tick(t.Context()); err != nil {
			t.Fatal(err)
		}
		got := xformX(targetOp(t, p, "target"), p.Probe().Display().PointsPerPixel)
		if math.Abs(got-c.want) > .1 {
			t.Fatalf("%s sample = %.2f px, want %.0fpx", c.mode, got, c.want)
		}
	}
}

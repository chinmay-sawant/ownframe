package main

import (
	"math"
	"testing"
)

// rangePose builds a scroll-driven route and samples it at half the page
// scroll range.
func rangePose(t *testing.T, declaration string) float64 {
	t.Helper()
	html := `<style>@keyframes route{from{transform:translateX(0px)}to{transform:translateX(100px)}}` +
		`body{margin:0;height:2000px}` +
		`#target{width:40px;height:20px;background:#123456;animation-name:route;` +
		`animation-duration:1000ms;animation-fill-mode:both;animation-timing-function:linear;` +
		`animation-timeline:scroll();` + declaration + `;}</style>` +
		`<body><div id="target"></div></body>`
	p := motionPage(t, html, "")
	p.Scroll().SetScrollOffset(0, 900)
	if err := p.Testing().Tick(t.Context()); err != nil {
		t.Fatal(err)
	}
	return xformX(targetOp(t, p, "target"), p.Probe().Display().PointsPerPixel)
}

func assertRangePose(t *testing.T, name, declaration string, want float64) {
	t.Helper()
	if got := rangePose(t, declaration); math.Abs(got-want) > .1 {
		t.Fatalf("%s sampled %.2f px, want %.2f", name, got, want)
	}
}

// TestCatalogAnimationRangeReachesOwnFrameFrame covers the range shorthand
// through the page driver at a fixed scroll position.
func TestCatalogAnimationRangeReachesOwnFrameFrame(t *testing.T) {
	assertRangePose(t, "animation-range", "animation-range:25% 75%", 50)
}

func TestCatalogAnimationRangeStartReachesOwnFrameFrame(t *testing.T) {
	assertRangePose(t, "animation-range-start", "animation-range-start:25%", 100.0/3)
}

func TestCatalogAnimationRangeEndReachesOwnFrameFrame(t *testing.T) {
	assertRangePose(t, "animation-range-end", "animation-range-end:75%", 200.0/3)
}

// TestCatalogAnimationRangeCenterReachesOwnFrameFrame checks the center
// longhand on the same driver. A center at the middle of the source range
// leaves the middle of the animation at the middle of the scroll range.
func TestCatalogAnimationRangeCenterReachesOwnFrameFrame(t *testing.T) {
	assertRangePose(t, "animation-range-center", "animation-range-center:50%", 50)
}

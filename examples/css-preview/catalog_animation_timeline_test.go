package main

import (
	"math"
	"testing"
)

// TestCatalogAnimationTimelineSamplesHostProgress runs the page's own
// driver against the page scroller. A scroll timeline reads the scroll
// offset, a named timeline reads the named scroller an ancestor declares,
// and a view timeline reads the subject's visibility.
func TestCatalogAnimationTimelineSamplesHostProgress(t *testing.T) {
	cases := []struct {
		name, timeline, extra string
		scroll                int
		want                  float64
	}{
		{"scroll", "scroll()", `<div id="target"></div>`, 900, 50},
		{"named", "--read", `<div id="tracks"><div id="target"></div></div>`, 900, 50},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			base := `#target{width:40px;height:20px;background:#123456;animation-name:route;` +
				`animation-duration:1000ms;animation-fill-mode:both;animation-timing-function:linear;` +
				`animation-timeline:` + c.timeline + `;}`
			decl := ""
			if c.name == "named" {
				decl = `#tracks{scroll-timeline:--read y}`
			}
			html := `<style>@keyframes route{from{transform:translateX(0px)}to{transform:translateX(100px)}}` +
				`body{margin:0;height:2000px}` + decl + base + `</style><body>` + c.extra + `</body>`
			p := motionPage(t, html, "")
			p.Scroll().SetScrollOffset(0, c.scroll)
			if err := p.Testing().Tick(t.Context()); err != nil {
				t.Fatal(err)
			}
			got := xformX(targetOp(t, p, "target"), p.Probe().Display().PointsPerPixel)
			if math.Abs(got-c.want) > .1 {
				t.Errorf("%s timeline sample = %.2f px, want %.2f", c.name, got, c.want)
			}
		})
	}
}

// TestCatalogAnimationViewTimelineSamplesVisibility checks that a view
// timeline moves with the subject's visibility and clamps at the end.
func TestCatalogAnimationViewTimelineSamplesVisibility(t *testing.T) {
	html := `<style>@keyframes route{from{transform:translateX(0px)}to{transform:translateX(100px)}}` +
		`body{margin:0;height:2000px}` +
		`#target{position:absolute;top:80px;left:0;width:40px;height:20px;background:#123456;` +
		`animation-name:route;animation-duration:1000ms;animation-fill-mode:both;` +
		`animation-timing-function:linear;animation-timeline:view();}</style>` +
		`<body><div id="target"></div></body>`
	p := motionPage(t, html, "")
	if err := p.Testing().Tick(t.Context()); err != nil {
		t.Fatal(err)
	}
	first := xformX(targetOp(t, p, "target"), p.Probe().Display().PointsPerPixel)
	p.Scroll().SetScrollOffset(0, 120)
	if err := p.Testing().Tick(t.Context()); err != nil {
		t.Fatal(err)
	}
	second := xformX(targetOp(t, p, "target"), p.Probe().Display().PointsPerPixel)
	if second <= first+1 {
		t.Fatalf("view timeline moved from %.2f to %.2f, want it past the first sample", first, second)
	}
	p.Scroll().SetScrollOffset(0, 3000)
	if err := p.Testing().Tick(t.Context()); err != nil {
		t.Fatal(err)
	}
	end := xformX(targetOp(t, p, "target"), p.Probe().Display().PointsPerPixel)
	if math.Abs(end-100) > .1 {
		t.Fatalf("view timeline end sample = %.2f px, want 100", end)
	}
}

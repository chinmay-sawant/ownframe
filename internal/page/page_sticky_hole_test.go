package page_test

import "testing"

func TestStickyViewportScrollContract(t *testing.T) {
	t.Parallel()
	source := `<style>body{margin:0}#before{height:80px}#region{height:300px}#head{position:sticky;top:0;height:40px;background:red;z-index:2}#body{height:1000px;background:blue}</style><div id="before"></div><div id="region"><div id="head" data-action="activate" role="button" tabindex="0"></div><div id="body"></div></div>`
	p := newViewportPage(t, source, 200, 100)
	for _, test := range []struct{ scroll, want int }{{0, 80}, {100, 100}, {200, 200}, {340, 340}, {360, 340}, {0, 80}} {
		before := p.Stats()
		p.SetScrollOffset(0, test.scroll)
		head := viewportBox(t, p, "head")
		if head.Y != float64(test.want) {
			t.Fatalf("scroll %d: hit box Y %g want %d", test.scroll, head.Y, test.want)
		}
		if p.Stats().Layouts != before.Layouts {
			t.Fatal("sticky scroll performed a layout")
		}
		if test.scroll == 200 {
			hit, ok := p.BoxAt(10, 210)
			if !ok || hit.ID != "head" {
				t.Fatalf("sticky paint/hit overlap target %+v", hit)
			}
		}
	}
}

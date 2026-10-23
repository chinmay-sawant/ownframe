package ownframe_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestPerfProbeStages(t *testing.T) {
	previousOps := 0
	for _, rows := range []int{1, 100, 1000} {
		t.Run(fmt.Sprint(rows), func(t *testing.T) {
			source := `<style>div{height:20px;background:#eee}</style>` + strings.Repeat(`<div>row</div>`, rows)
			p, err := ownframe.NewWithOptions(ownframe.Config{HTML: source, Width: 400, Height: 300}, ownframe.WithPerf(true))
			if err != nil {
				t.Fatal(err)
			}
			if err := p.Redraw(t.Context()); err != nil {
				t.Fatal(err)
			}
			first := p.Probe().Stats()
			if err := p.Redraw(t.Context()); err != nil {
				t.Fatal(err)
			}
			warm := p.Probe().Stats()
			if warm.Parses != first.Parses || warm.Cascades != first.Cascades {
				t.Fatal("warm redraw reparsed or recollected CSS")
			}
			if warm.Ops <= previousOps {
				t.Fatal("probe did not increase placed work with row count")
			}
			previousOps = warm.Ops
			sum := warm.LastTemplate + warm.StyleTime + warm.LayoutTime + warm.PaintTime + warm.OtherTime
			if sum != warm.LastRedraw {
				t.Fatalf("stage sum=%s redraw=%s", sum, warm.LastRedraw)
			}
			t.Logf("rows=%d ops=%d parses=%d cold=%s warm=%s style=%s layout=%s other=%s", rows, warm.Ops, warm.Parses, first.LastRedraw, warm.LastRedraw, warm.StyleTime, warm.LayoutTime, warm.OtherTime)
		})
	}
}

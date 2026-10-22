package replay

import (
	"context"
	"fmt"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

// Called only from parityGame.Draw, alongside the other pixel checks.
func checkTranslucentPixels() error {
	data, err := opaquePNG()
	if err != nil {
		return err
	}
	cases := []struct {
		name, source string
		grid         bool
	}{
		{name: "translucent fill", source: `<body style="margin:0;background:white;height:140px"><div style="width:40px;height:40px;background:red;opacity:.5"></div></body>`},
		{name: "translucent image", source: `<body style="margin:0;background:white;height:140px"><div style="opacity:.5"><img src="t.png" style="width:20px;height:20px"></div></body>`},
		{name: "translucent grid", source: `<body style="margin:0;background:white;height:140px"><table style="border-collapse:collapse"><tr><td style="border:4px solid red;width:40px;height:20px;opacity:.5">x</td></tr></table></body>`, grid: true},
	}
	for _, tc := range cases {
		d, err := render.DisplayListState(context.Background(), tc.source, 200, 140,
			render.State{Images: func(string) ([]byte, error) { return data, nil }})
		if err != nil {
			return err
		}
		point := [2]int{10, 10}
		if tc.grid {
			found := false
			for i := range d.Ops {
				op := &d.Ops[i]
				if op.Kind != layout.DisplayOpGridRun || len(op.Grid.Segs) == 0 {
					continue
				}
				op.Alpha = 0.5
				segment := op.Grid.Segs[0]
				point = [2]int{int((segment.X + segment.W/2) * d.PixelPerPoint), int((segment.Y + segment.H/2) * d.PixelPerPoint)}
				found = true
				break
			}
			if !found {
				return fmt.Errorf("translucent fixture produced no grid")
			}
		}
		if err := compareGroupPixels(d, tc.name, [][2]int{point, {190, 130}}); err != nil {
			return err
		}
	}
	return nil
}

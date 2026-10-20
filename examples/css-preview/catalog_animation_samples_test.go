package main

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/animation"
	"github.com/chinmay-sawant/ownframe/internal/cssanimation"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

func sampleMotion(t *testing.T, declarations string, input animation.FrameInput) float64 {
	t.Helper()
	style := `#target{width:40px;height:20px;background:#123456;` + declarations + `}` + keyframeCSS
	d, b, indices, base := motionFixture(t, `<div id="target"></div>`, style)
	start := time.Unix(100, 0)
	p, err := cssanimation.Bind(d, b, indices, base, start)
	if err != nil {
		t.Fatal(err)
	}
	if input.Now.IsZero() {
		input.Now = start
	}
	p.StepWithInput(input)
	return (cssmotion.Read(&d.Ops[indices[0]]).X - base[0].X) / d.PointsPerPixel
}

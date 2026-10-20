package main

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
	"github.com/chinmay-sawant/ownframe/internal/csstransition"
)

func TestCatalogTransitionBehaviorReachesOwnFrameFrames(t *testing.T) {
	for _, behavior := range []string{"allow-discrete", "normal"} {
		t.Run(behavior, func(t *testing.T) {
			css := `#target{width:50px;height:20px;left:80px;top:70px;background:#f00;` +
				`transition-property:position;transition-duration:1000ms;` +
				`transition-timing-function:linear;transition-behavior:` + behavior + `}`
			old, oldBox, _, _ := motionFixture(t, `<div id="target" style="position:static"></div>`, css)
			next, box, indices, _ := motionFixture(t, `<div id="target" style="position:absolute"></div>`, css)
			if len(box.Transitions) != 1 || box.Transitions[0].Behavior != behavior {
				t.Fatalf("Blinkless transition metadata = %+v", box.Transitions)
			}
			from := cssmotion.Capture(old, oldBox)
			to, ops := cssmotion.Capture(next, box), cssmotion.Resolve(next, indices)
			if len(from) != 1 || len(to) != 1 || len(ops) != 1 || from[0].X == to[0].X {
				t.Fatalf("old/new position paints = %+v / %+v", from, to)
			}
			start := time.Unix(9, 0)
			player, err := csstransition.Bind(next, box, "position", indices, from, start)
			if err != nil {
				t.Fatal(err)
			}
			active := player.Step(start.Add(499 * time.Millisecond))
			if behavior == "allow-discrete" && (!active || cssmotion.Read(ops[0]) != from[0]) {
				t.Fatal("allow-discrete did not hold the old frame before midpoint")
			}
			active = player.Step(start.Add(500 * time.Millisecond))
			if behavior == "allow-discrete" && !active || behavior == "normal" && active || cssmotion.Read(ops[0]) != to[0] {
				t.Fatalf("%s midpoint active=%v paint=%+v want=%+v", behavior, active, cssmotion.Read(ops[0]), to[0])
			}
		})
	}
}

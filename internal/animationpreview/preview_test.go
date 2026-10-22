package animationpreview

import (
	"context"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

func TestEachStartButtonPlaysItsOwnKeyframes(t *testing.T) {
	for _, item := range targets {
		t.Run(item.key, func(t *testing.T) {
			ctx := context.Background()
			data := &view{Started: map[string]bool{}}
			page := testPage(t, data)
			demo := newDemo(page, data)
			page.Handle(ownframe.Handlers{Click: demo.click})
			box, ok := findBox(page.Boxes(), item.id)
			if !ok {
				t.Fatalf("missing animation target %q", item.id)
			}
			indices := cssmotion.OperationIndices(page.Testing().Display(), box)
			clickStart(t, ctx, page, item.button)
			if err := demo.tick(ctx); err != nil {
				t.Fatal(err)
			}
			sample := demo.samples[item.key]
			if sample == nil || sample.player == nil || !data.Started[item.key] {
				t.Fatal("Start did not bind an animation player")
			}
			ops := cssmotion.Resolve(page.Testing().Display(), indices)
			if len(indices) == 0 || len(ops) != len(indices) {
				t.Fatalf("resolved %d of %d target operations", len(ops), len(indices))
			}
			before := paint(ops)
			box, ok = findBox(page.Boxes(), item.id)
			if !ok || len(box.Animations) == 0 {
				t.Fatal("redrawn target has no animation metadata")
			}
			record := box.Animations[0]
			active := sample.player.Step(sample.start.Add(time.Duration(record.DelayStartMS+record.DurationMS/2) * time.Millisecond))
			changed := false
			for i, op := range ops {
				changed = changed || cssmotion.Read(op) != before[i]
			}
			if !active || !changed {
				t.Fatalf("%s did not sample its keyframes: active=%v before=%+v after=%+v", item.key, active, before, paint(ops))
			}
		})
	}
}

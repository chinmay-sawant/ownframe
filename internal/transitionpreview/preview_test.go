package transitionpreview

import (
	"context"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

func TestEachStartButtonPlaysItsOwnTransition(t *testing.T) {
	for _, item := range targets {
		t.Run(item.key, func(t *testing.T) {
			ctx := context.Background()
			data := &view{Started: map[string]bool{}}
			page := testPage(t, data)
			demo := newDemo(page, data)
			page.Handle(ownframe.Handlers{Click: demo.click})
			box, ok := findBox(page.Boxes(), item.id)
			if !ok {
				t.Fatalf("missing transition target %q", item.id)
			}
			indices := cssmotion.OperationIndices(page.Testing().Display(), box)
			if len(indices) == 0 {
				t.Fatal("target has no display operations before the transition")
			}
			clickStart(t, ctx, page, item.button)
			if err := demo.tick(ctx); err != nil {
				t.Fatal(err)
			}
			sample := demo.samples[item.key]
			if sample == nil || sample.player == nil || !data.Started[item.key] {
				t.Fatal("Start did not bind a transition renderer")
			}
			ops := cssmotion.Resolve(page.Testing().Display(), indices)
			if len(ops) != len(indices) {
				t.Fatalf("resolved %d of %d captured operations", len(ops), len(indices))
			}
			before := paint(ops)
			box, ok = findBox(page.Boxes(), item.id)
			if !ok || len(box.Transitions) == 0 {
				t.Fatal("redrawn target has no transition metadata")
			}
			record := box.Transitions[0]
			now := sample.start.Add(time.Duration(record.DelayMS+record.DurationMS/2) * time.Millisecond)
			active := sample.player.Step(now)
			changed := false
			for i, op := range ops {
				changed = changed || cssmotion.Read(op) != before[i]
			}
			if !active || !changed {
				t.Fatalf("%s did not sample an intermediate frame: active=%v before=%+v after=%+v", item.key, active, before, paint(ops))
			}
		})
	}
}

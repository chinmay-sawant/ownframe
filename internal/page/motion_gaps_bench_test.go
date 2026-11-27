package page

import (
	"context"
	"testing"
	"time"
)

// BenchmarkMotionPropertyGaps measures per-frame step cost for the high-impact
// CSS motion property gaps before and after moving them onto fast in-place /
// compositor paths.
func BenchmarkMotionPropertyGaps(b *testing.B) {
	ctx := context.Background()

	runBench := func(b *testing.B, src string) {
		b.Helper()
		p, err := New(Config{HTML: src, Width: 300, Height: 160, RetainState: true})
		if err != nil {
			b.Fatal(err)
		}
		if err := p.Redraw(ctx); err != nil {
			b.Fatal(err)
		}
		now := time.Now()
		b.ReportAllocs()
		b.ResetTimer()
		for b.Loop() {
			now = now.Add(16 * time.Millisecond)
			stepLayoutMotion(p, ctx, now)
			p.motion.Step(p.draw.Display, now)
		}
	}

	b.Run("BorderRadiusAnim", func(b *testing.B) {
		runBench(b, `<!DOCTYPE html><html><head><style>
			@keyframes morph {
				from { border-radius: 4px; }
				to   { border-radius: 24px; }
			}
			#card { width: 140px; height: 60px; background: #2255cc; border-radius: 4px; animation: morph 1s linear infinite; }
		</style></head><body><div id="card">pill</div></body></html>`)
	})

	b.Run("ColorAlphaAnim", func(b *testing.B) {
		runBench(b, `<!DOCTYPE html><html><head><style>
			@keyframes glow {
				from { background-color: rgba(61, 220, 132, 0.1); }
				to   { background-color: rgba(61, 220, 132, 0.9); }
			}
			#badge { width: 120px; height: 40px; background-color: rgba(61, 220, 132, 0.1); animation: glow 1s linear infinite; }
		</style></head><body><div id="badge">live</div></body></html>`)
	})

	b.Run("IndividualTranslateAnim", func(b *testing.B) {
		runBench(b, `<!DOCTYPE html><html><head><style>
			@keyframes slide {
				from { translate: 0px 0px; }
				to   { translate: 80px 0px; }
			}
			#chip { width: 100px; height: 40px; background: #cc5522; animation: slide 1s linear infinite; }
		</style></head><body><div id="chip">chip</div></body></html>`)
	})
}

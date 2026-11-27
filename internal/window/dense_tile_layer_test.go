package window

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/page"
	"github.com/hajimehoshi/ebiten/v2"
)

// TestDenseSingleScreenTileLayerCompositing is an end-to-end Ebiten game-loop
// regression test for dense single-screen dashboards (like examples/trade) on
// the tile surface. It verifies that:
//  1. A 1280x720 page with >300 text runs and >150 rounded fills stays on the
//     tile surface (`drawTileSurface`) with top-composited retained layers.
//  2. An anonymous `overflow: hidden` animated marquee promotes to a retained
//     layer even with an unrelated `mix-blend-mode` badge elsewhere on the page.
//  3. Steady-state animation frames composite promoted layers with zero tile
//     repaints (`repaintBlit`) and zero extra layouts.
func TestDenseSingleScreenTileLayerCompositing(t *testing.T) {
	monitors := []struct {
		name      string
		w, h      int
		cellCount int
	}{
		{name: "1280x720", w: 1280, h: 720, cellCount: 165},
		{name: "1920x1080", w: 1920, h: 1080, cellCount: 220},
		{name: "2560x1600", w: 2560, h: 1600, cellCount: 260},
	}

	for _, mon := range monitors {
		t.Run(mon.name, func(t *testing.T) {
			ctx := context.Background()

			var b strings.Builder
			fmt.Fprintf(&b, `<!DOCTYPE html><html><head><style>
				html, body { margin: 0; padding: 0; width: %dpx; height: %dpx; overflow: hidden; background: #090b0e; color: #e8edf3; font-family: sans-serif; font-size: 11px; }
				.grid { display: flex; flex-wrap: wrap; gap: 4px; padding: 8px; height: %dpx; overflow: hidden; }
				.cell { width: 96px; height: 22px; padding: 2px 6px; border-radius: 5px; background: #141922; display: flex; justify-content: space-between; align-items: center; }
				.blend-badge { position: absolute; left: 10px; top: %dpx; width: 80px; height: 20px; background: #3ddc84; mix-blend-mode: multiply; }
				.tape { position: absolute; left: 120px; top: %dpx; width: %dpx; height: 28px; overflow: hidden; background: #0d1014; }
				.track { display: flex; gap: 24px; animation: marquee 20s linear infinite; }
				.tick { white-space: nowrap; color: #3ddc84; }
				@keyframes marquee { from { transform: translateX(0px); } to { transform: translateX(-400px); } }
			</style></head><body><div class="grid">`,
				mon.w, mon.h, mon.h-80, mon.h-70, mon.h-60, mon.w-140)
			for i := range mon.cellCount {
				fmt.Fprintf(&b, `<div class="cell"><span>SYM%03d</span><span>%d.25</span></div>`, i, 100+i)
			}
			b.WriteString(`</div><div class="blend-badge"></div><div class="tape"><div class="track">`)
			for i := range 12 {
				fmt.Fprintf(&b, `<span class="tick">NVDA %d.50</span>`, 120+i)
			}
			b.WriteString(`</div></div></body></html>`)

			p, err := page.New(page.Config{
				HTML:        b.String(),
				Width:       mon.w,
				Height:      mon.h,
				RetainState: true,
			})
			if err != nil {
				t.Fatalf("page.New: %v", err)
			}
			if err := p.Redraw(ctx); err != nil {
				t.Fatalf("page.Redraw: %v", err)
			}

			disp := p.Display()
			if disp == nil {
				t.Fatal("nil display after Redraw")
			}
			if len(disp.Ops) < 400 {
				t.Fatalf("expected dense display (>400 ops), got %d", len(disp.Ops))
			}

			s := newShell(ctx, p)
			s.Layout(mon.w, mon.h)
			s.display = disp
			dst := ebiten.NewImage(mon.w, mon.h)
			defer dst.Deallocate()

			// Step motion once so the marquee animation has an active pose before Frame 1.
			time.Sleep(20 * time.Millisecond)
			if err := p.Tick(ctx); err != nil {
				t.Fatalf("initial Tick: %v", err)
			}

			// Frame 1: builds tile grid, promotes the anonymous overflow-clipped marquee
			// layer despite the unrelated mix-blend-mode badge, and syncs promoted boxes.
			s.Draw(dst)

			if !s.partial.tileLayers {
				t.Fatalf("expected top-composited retained layers on tile surface, got tileLayers=false (mode=%v, specs=%d, active=%v)",
					s.partial.mode, len(s.layerSpecs(disp)), s.renderer.LayersActive(disp))
			}
			if boxes := s.renderer.ActiveLayerBoxIndices(disp); len(boxes) == 0 {
				t.Fatal("expected at least one promoted layer box index for anonymous clipped marquee")
			}

			layoutsBefore := p.Stats().FullLayouts

			// Frames 2..4: advance real motion time and draw through the Ebiten loop.
			for frame := range 3 {
				time.Sleep(20 * time.Millisecond)
				if err := p.Tick(ctx); err != nil {
					t.Fatalf("frame %d Tick: %v", frame, err)
				}
				s.Draw(dst)
				if !s.partial.tileLayers {
					t.Fatalf("frame %d lost tileLayers", frame)
				}
				if s.partial.mode != repaintBlit {
					t.Fatalf("frame %d repaint mode = %v, want repaintBlit (zero tile repaints)", frame, s.partial.mode)
				}
			}

			if got := p.Stats().FullLayouts; got != layoutsBefore {
				t.Fatalf("steady marquee animation triggered %d extra full layouts", got-layoutsBefore)
			}
		})
	}
}

# OwnFrame v0.0.7

Release on top of `v0.0.6` (`b65065a2`), delivering pure-Go animation, layout, Grid, and viewport-resize performance gains with Blinkless `v0.1.9`, dynamic multi-monitor refresh-rate animation pacing, a `departure-board` Windows startup fix, and updated CPU + native Windows GPU benchmark baselines.

- Range: `v0.0.6` (`b65065a2`) through `chore/performance-anim-grid-layout-pacing` (2026-11-25)
- Engine: pins `github.com/chinmay-sawant/blinkless` to `v0.1.9` (merged in `blinkless#56`)

## Summary

| Area | What changed |
| --- | --- |
| **Multi-Monitor Refresh-Rate Animation Pacing (`internal/window`)** | Added zero-cgo active-monitor refresh-rate detection for Windows (`user32.dll` `MonitorFromWindow` / `MonitorFromRect` → `GetMonitorInfoW` → `EnumDisplaySettingsW`), Linux (`/sys/class/drm` connected mode lines), and macOS (`purego` CoreGraphics). Default `MaxFPS == 0` on desktop now auto-caps continuous animations to the active monitor's Hz and dynamically updates when dragging a window across monitors with different refresh rates (e.g., 165 Hz ↔ 144 Hz), while preserving idle parking and explicit `MaxFPS` overrides. |
| **Motion & Template Scratch Pooling (`internal/motion`, `internal/page`)** | Reused `transformScratch` and stack-backed `appendPerspectiveStops` in `motion.Player`, cached per-effect transform/perspective/SVG flags, deduplicated unchanged CSS transform strings, stack-allocated small restamp/layout override arrays in `motion_wire.go`, and replaced `strings.Builder` pooling with `*bytes.Buffer` reuse (`ExecuteReuse`) in `internal/page/draw/template.go`. |
| **Blinkless v0.1.9 Engine Adoption** | Pulls in Blinkless `v0.1.9`: two-pass ancestor-only `stampStateAt` + pooled `restampScratch` + zero-copy 3-index patch/box sub-slices, lock-free bit-exact `[128]` ASCII glyph/advance/point tables + `MeasureSimpleASCII`, single-pass `O(N)` `placedTextIndex`, selector match-cache `id` elision, pooled `childSlabs`, multi-word single-line fast path, `O(1)` viewport-only `css.Relayout`, and pooled dense `[]uint64` Grid occupancy bitsets. |
| **Example Reliability & Benchmark Baselines** | Guarded `cellOp(d, b) == nil` in `examples/departure-board/board` (preventing a Windows startup tick panic), and updated `BENCHMARKS.md`, `documentation/performance-baseline.md`, and `frontend/src/Benchmarks.jsx` with sequential CPU and native Windows GPU (`RTX 4060 @ 144 Hz`) baselines. |

## Measured Impact (vs. `v0.0.6`)

| Benchmark | `v0.0.6` | `v0.0.7` | Improvement |
| --- | --- | --- | --- |
| `BenchmarkAnimationTick` *(534-op CSS gallery)* | 465.8 µs · 180 allocs · 226 KB | **97.7 µs · 38 allocs · 16.6 KB** | **4.77× faster**, **13.6× less memory** |
| `ScaleStages/Rows1000/CSSRelayout` | 19.7 µs · 7 allocs | **59.2 ns · 1 alloc** | **333× faster** (`O(1)`) |
| `ScaleStages/Rows1000/DisplayList` | 14.05 ms · 18,312 allocs | **12.57 ms · 4,338 allocs** | **4.22× fewer allocs**, **−1.5 ms** |
| `LayoutModels/Grid` *(200 rows × 3 cols)* | 2.98 ms · 7,204 allocs | **2.37 ms · 2,410 allocs** | **25% faster**, **2.99× fewer allocs** |
| `ListWorkloads/Resize` *(1,000 rows)* | 14.63 ms · 19,277 allocs | **11.59 ms · 4,298 allocs** | **26% faster**, **4.49× fewer allocs** |
| `LargeRedraw` *(unwindowed)* | 21.6 ms · 22,068 allocs | **17.14 ms · 7,234 allocs** | **26% faster**, **3.05× fewer allocs** |
| Windows GPU `animations.exe` *(RTX 4060, 144 Hz)* | 325.3 FPS · 162.3% CPU · 713 MB/s · 44.4 ms GC | **144.1 FPS · 89.6% CPU · 338 MB/s · 18.0 ms GC** | **Auto-paced 144 Hz**, **−45% CPU**, **2.47× less GC** |

## Breaking changes

None. Public root API and exported signatures are unchanged.

## Validation

- `make lint` (Go library + examples, docs, plans, tracked file size, examples inventory, internal imports, ESLint)
- `make site-check` (9 WebAssembly demos + 140 layout/theme/viewport checks)
- `make graphics-check` and `make test` / `make test-race`

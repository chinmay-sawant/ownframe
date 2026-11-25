# Performance baseline

One table per release. Fill it on the same machine when you can, and label every entry with the commit hash. Numbers from different machines do not compare.

Full procedure lives in [performance.md](performance.md).

## How to fill it

Run the automated script first. It runs page, replay-scan, platform-click, and Normal/Large/Flappy workload
benchmarks. It does not launch windows or fill idle CPU, RSS, or native frame
rows. Raw text and a sibling JSON report retain the samples and provenance.

```sh
sh scripts/perf-baseline.sh
```

Then run the manual steps. Do the 30 to 60 second idle watch from [performance.md](performance.md) for each app, and the soak scenario that matches the release. Write the soak verdict and the machine name by hand.

Copy the template section for each release. Leave cells empty when a script or counter is unavailable, and say so in the notes row instead of guessing.

## Template

| Metric | Normal UI | Large UI | Flappy |
|---|---|---|---|
| Idle RSS, MB | | | |
| Idle CPU, percent | | | |
| Go heap, MB | | | |
| Avg frame, ms | | | |
| p95 frame, ms | | | |
| p99 frame, ms | | | |
| Alloc per frame, KB | | | |
| Layout, ms | | | |
| Paint, ms | | | |
| Soak verdict | | | |
| Commit | | | |
| Machine | | | |
| Notes | | | |

## v0.0.2 measured numbers

Measured 2026-10-05 on commit 82ca687 plus the `perf-longrun.sh` key and paren fix, machine Chinmay. Idle captures ran 30 s per app, desktop and `-web`; the soak ran 1 minute, undriven, `-web` large. Full working notes live in `plans/v0.0.2/performance/phase-01-baseline.md`.

| Metric | Normal UI | Large UI | Flappy |
|---|---|---|---|
| Idle RSS, MB | 28.6 web; desktop crashes | 52.0 web; desktop crashes | 28.0 web; 164 desktop |
| Idle CPU, percent | ~0 web | 0.5-3.5 web | ~0 web; 16-23 desktop under Xvfb |
| Go heap, MB | unavailable, runtime nil in prod | unavailable, runtime nil in prod | unavailable, runtime nil in prod |
| Avg/p95/p99 frame, ms | unavailable headless, ring not exposed | unavailable headless, ring not exposed | unavailable headless, ring not exposed |
| Alloc per cold redraw | 373 KB | 26.15 MB | 362 KB |
| Cold redraw, ms | 0.63 | 88.76 | 0.40 |
| Display-list build, ms | 0.47 | 84.29 | 0.30 |
| Layout, ms | 0.08 | 1.66 | 0.045 |
| Paint, ms | 0 vector path, Lay 7.7 ms on small fixture | 0 vector path | 0 vector path |
| Soak verdict | flat counters, no errors | counters frozen at 1/1/1/1/1, no errors, heap columns empty | desktop held 30 s, no growth |
| Commit | 82ca687 + script fix | 82ca687 + script fix | 82ca687 + script fix |
| Machine | Chinmay | Chinmay | Chinmay |
| Notes | desktop panics in Ebiten atlas, buffer 480x16494 | desktop panics in Ebiten atlas, buffer 800x36579 | benchmark C tick animation |

Stage benches from `scripts/perf-baseline.sh`: TemplateExecute 578 ns, HTMLParse 1493 ns, CSSApply 4432 ns, DisplayList 169 us, Lay 7.7 ms, RedrawWarm 143 us, RedrawCold 163 us, RedrawHeavyWarm 333 us, DrawRectScan 150-166 ns at 0 allocs, ClickCount 607 us.

## v0.0.2 demo apps after the perf pass

Measured 2026-10-13 on commit `de786e6`, machine Chinmay (WSL2, Mesa software rasterizer, 24 cores). These rows cover the three demos that take `-perf` (departure-board, spotify-player, wispr-flow-dashboard), not the three benchmark apps in the sections above. Panel rows (Go heap, RSS, alloc, frame, FPS) were read from the devtools Frame tab in a 1908x999 window with the dock open; idle CPU and the RSS means come from `scripts/perf-apps.sh` with the dock closed, 20 one-second samples per app. The before column is the same tree before this pass: the original Frame tab capture for the panel rows, the pre-pacing loop for idle CPU.

| Metric | departure-board | spotify-player | wispr-flow-dashboard |
|---|---|---|---|
| Go heap, MB | 345.1 -> ~15 live | not sampled | not sampled |
| RSS, MB | 678.3 -> 183.0 | ~200 -> 211.4 | 192-226 -> 201.0 |
| Idle CPU, one core | ~28% -> 5.2% mean, 7.0% controlled | ~42% -> 38.4% | 7.8% -> 3.3% controlled, 4.5% harness |
| Frame p99, ms | 22.9 -> 19.2 | | |
| Long frames over 25 ms | 1 -> 0 | | |
| Alloc per 500 ms sample | 6.1 -> 2.3 MB | | |
| Vertex high-water, MB | 86.6 -> 9.4 | | |
| Static text repaint | 1.0 ms, 873 KB, 4,512 allocs -> 35 us, 27 KB, 144 allocs | | |
| Transform pages | bitmap -> replay, rotated card 6 ops | | |
| Blend groups | bitmap -> replay, 4 ops through 2 layers | | |
| Headless dump, worst cascade frame | 115 of 844 -> 115 of 844 | | |

Notes. The frame average moved from 14.2 to 16.8 ms because the loop now holds the 60 Hz cadence instead of running free at 68.5 FPS; the tail improved and the long frame is gone. RSS is dominated by the software GL stack (libLLVM, swrast, and the D3D12 libraries, about 150 MB of shared library pages), not by the app heap. Spotify's tick damages every frame, so idle pacing never throttles it and its CPU stays near its playing rate. Raw agent reports and per-run samples live under `temp/perf-parallel/`.

## v0.0.2 second perf pass (2026-10-13, perf/damage-list)

Merged in PR #27 for the v0.0.2 release. Same machine and method
as the section above. This pass added page-driven frame deadlines, dock
occlusion clipping, element paint chunks, compositor animations, and the diff
cost rule.

| Metric | Before (de786e6) | After | Delta |
|---|---:|---:|---|
| Board idle presents | 8/s | 1 to 2/s | about 5x fewer |
| Board idle CPU, median / mean | 5.0% / 5.1% | 2.0% / 3.6% (NoIdlePace control 20.4%) | 2.5x lower median |
| Board RSS mean, harness | 183.0 MB | 189.4 MB | +6.4 MB, driver-stack noise |
| Board Go heap live | ~15 MB | ~17 MB | +2 MB, chunk cache |
| Damaged card repaint | 8,589 ns, 5 ops, 18 allocs | 4,432 ns, 1 blit, 3 allocs | 1.9x faster, 6x fewer allocs |
| Animated card frame | 170,854 ns, 203 allocs (Redraw) | 176 ns, 0 allocs | about 970x |
| 20 small changes in one region | full-frame rebuild | partial 10x240, page diff +0.8 us | rebuild avoided |
| Dock-open damage, 1908x1080, 340 px dock | 152,400 px, 6 blits | 50,400 px, 4 blits | 3x fewer pixels, 1.5x fewer blits |
| Board dump, worst cascade frame | 115 of 844 | 115 of 844 | flat, tick side unchanged |
| Wispr idle CPU, harness mean | 4.5% | 4.3% | flat |
| Spotify idle CPU, harness mean | 38.4% | 43.4% | playback variation, not throttled |

Notes. The board's mean idle CPU is dominated by the one to two second flap
burst every 12 seconds; the median shows the parked rate. Spotify's tick
damages every frame, so pacing never throttles it and its CPU follows its
playback variation. Raw agent reports live under `temp/perf-parallel2/`.

## v0.0.2 prior numbers

These rows predate the baseline scripts. They come from `plans/v0.0.2/release-notes.md` and `repaint.md` at 480x640 and 640x480 on the author's machine. They are machine specific and use older harnesses, so treat them as history, not as targets.

| Metric | Value | Source |
|---|---|---|
| css Apply, 200 rules | about 200 us | release notes pipeline split |
| css Relayout, cached sheets | about 1.2 us | release notes pipeline split |
| Redraw heavy warm | 489 to 708 us, down from 1.34 to 1.83 ms | `BenchmarkRedrawHeavyWarm` |
| DrawRect op filter | about 12 ns per op | `BenchmarkDrawRectScan` in `internal/replay` |
| Click through handler, execute, relayout, dirty rect | about 0.5 ms | `BenchmarkClickCount` in platform example |
| platform `#count`, 11 ops | 5 repainted, 6 skipped | repaint.md measured numbers |
| login `#message`, 18 ops | 3 repainted, 15 skipped | repaint.md measured numbers |
| states `#hover-btn`, 16 ops | 4 repainted, 12 skipped | repaint.md measured numbers |
| forms `#remember`, 45 ops | 5 repainted, 40 skipped | repaint.md measured numbers |


## Local v0.0.3 review baseline, 2026-10-23

[Raw samples](../plans/validation/performance-2026-10-23.txt) and
[JSON provenance](../plans/validation/performance-2026-10-23.json) record five
100 ms samples on the local uncommitted review tree based on `24d4267a`.
This is local validation, not a published release result or a native-host test.
The cold benchmark changes template output and reports one parse per operation.

| Workload | Median ms/op | Allocs/op |
|---|---:|---:|
| Warm redraw | 0.139363 | 173.0 |
| Cold redraw | 0.150259 | 249.0 |
| Normal | 0.477405 | 424.0 |
| Large | 132.667189 | 134192.0 |
| Flappy tick | 0.000072 | 0.0 |

These are local descriptive samples. Runtime checks overlapped the capture,
so use a fresh isolated run for controlled release comparisons.

## v0.0.6 sequential baseline (2026-11-25)

Measured sequentially (`-p 1`, one suite at a time) on `ownframe` `master` commit `b65065a2` (`v0.0.6`) with `blinkless` `master` commit `787ca05` (`v0.1.8`) via `go.work`, machine Chinmay (`13th Gen Intel(R) Core(TM) i7-13700HX`, Linux `amd64`, Go `1.26.11`).

These are display-free Go CPU benchmarks. They measure template execution, HTML parsing, CSS application, layout, display-list construction, and in-place patching without GPU rasterization, software GL drivers (such as WSL2 Mesa), or Ebiten's 60 Hz vsync/TPS frame pacing.

Commands run sequentially:

```sh
sh scripts/perf-baseline.sh --count 3 --benchtime 500ms --out perf-baseline.txt
sh scripts/perf-baseline.sh --suite list --conditions "sequential isolated run" --count 3 --benchtime 500ms --out perf-list.txt
go test -p 1 -run '^$' -bench '^BenchmarkFrame_' -benchmem -benchtime 300ms ./test
go test -p 1 -run '^$' -bench . -benchmem -benchtime 300ms ./internal/motion ./internal/frame ./internal/textrun ./internal/page/patch
go test -p 1 -run '^$' -bench '^(BenchmarkScaleStages|BenchmarkLayoutModels)$' -benchmem -benchtime 500ms ./internal/page
go test -p 1 -run '^$' -bench '^BenchmarkListPatchVsRebuild$' -benchmem -count 3 -benchtime 500ms ./examples/perf-benchmarks
```

### Core pipeline and page (`scripts/perf-baseline.sh`, median of 3)

| Benchmark | Median time | Bytes/op | Allocs/op | What it measures |
|---|---:|---:|---:|---|
| `BenchmarkFlappyTick` | **57.2 ns** | 0 B | 0 | Per-frame game/tick loop |
| `BenchmarkDrawRectScan/dirty` | **177 ns** | 0 B | 0 | Display-list op culling scan (dirty) |
| `BenchmarkDrawRectScan/full` | **175 ns** | 0 B | 0 | Display-list op culling scan (full) |
| `BenchmarkRedrawWarm` | **485 ns** | 0 B | 0 | Unchanged warm frame check |
| `BenchmarkRedrawHeavyWarm` | **487 ns** | 0 B | 0 | Heavy unchanged warm frame check |
| `BenchmarkNormalRedraw` | **527 ns** | 0 B | 0 | Normal UI warm redraw |
| `BenchmarkRedrawStages/TemplateExecute` | **179 ns** | 560 B | 5 | Go `html/template` execution |
| `BenchmarkRedrawStages/HTMLParse` | **4.97 µs** | 27.9 KB | 16 | Blinkless HTML parse |
| `BenchmarkRedrawStages/CSSApply` | **2.25 µs** | 3.17 KB | 37 | Blinkless CSS selector match/apply |
| `BenchmarkRedrawStages/DisplayList` | **39.5 µs** | 34.1 KB | 115 | Layout, text measure, and display-list emit |
| `BenchmarkRedrawCold` | **72.2 µs** | 64.9 KB | 208 | Full cold parse, style, and layout (`1.000 parses/op`) |
| `BenchmarkClickCount` | **221 µs** | 118 KB | 536 | Click handler plus full redraw (`examples/platform`) |
| `BenchmarkLargeRedraw` | **21.6 ms** | 3.10 MB | 22,068 | Large unwindowed full rebuild |

### 1,000-row list workloads (`--suite list`, median of 3)

| Workload | PerfOff median | PerfOn median | PerfOff Bytes/op | PerfOff Allocs/op |
|---|---:|---:|---:|---:|
| `ScrollOffset` | **21.9 ns** | 21.9 ns | 0 B | 0 |
| `Unchanged` | **497 ns** | 542 ns | 0 B | 0 |
| `WindowedScroll` (overscan 4) | **588 µs** | 595 µs | 186 KB | 1,537 |
| `Resize` (all 1,000 rows) | **14.6 ms** | 14.6 ms | 3.15 MB | 19,277 |
| `TextUpdate` (all 1,000 rows) | **19.1 ms** | 18.9 ms | 6.42 MB | 50,293 |

### In-place patching vs. full rebuild (`BenchmarkListPatchVsRebuild`, 1,000 rows, median of 3)

| Benchmark | Median time | Bytes/op | Allocs/op | vs. `FullRebuild1Row` |
|---|---:|---:|---:|---|
| `FullRebuild1Row` | **18.97 ms** | 6.71 MB | 62,258 | 1.0× |
| `PatchElement1Row` | **599 µs** | 207 KB | 101 | **31.7× faster**, **616× fewer allocs** |
| `PatchElements10Rows` | **807 µs** | 271 KB | 614 | **23.5× faster**, **101× fewer allocs** |

### Pipeline stage scaling (`BenchmarkScaleStages`)

| Stage | 50 rows (ns/row) | 200 rows (ns/row) | 1,000 rows total | 1,000 rows Bytes/op | 1,000 rows Allocs/op | Share of 1,000-row rebuild |
|---|---:|---:|---:|---:|---:|---:|
| `TemplateExecute` | 2,386 | 2,375 | **2.41 ms** | 521 KB | 26,730 | 13.5% |
| `HTMLParse` | 1,391 | 1,316 | **1.32 ms** | 1.64 MB | 6,123 | 7.4% |
| `CSSApply` | 144.0 | 66.2 | **45.1 µs** | 6.66 KB | 75 | 0.3% |
| `CSSRelayout` | 26.4 | 20.7 | **19.7 µs** | 752 B | 7 | 0.1% |
| `DisplayList` | 15,046 | 13,883 | **14.05 ms** | 5.90 MB | 18,312 | **78.8%** |

### Layout model overhead (`BenchmarkLayoutModels`, 200 rows × 3 columns)

| Layout model | Total time | ns/row | Bytes/op | Allocs/op | vs. Block |
|---|---:|---:|---:|---:|---:|
| `Block` | **2.45 ms** | 12,237 | 1.05 MB | 2,588 | 1.00× |
| `Flex` | **2.68 ms** | 13,379 | 1.09 MB | 2,815 | 1.09× |
| `Grid` | **2.98 ms** | 14,891 | 1.77 MB | 7,204 | 1.22× |

### Frame and internal microbenchmarks

| Benchmark | Time/op | Bytes/op | Allocs/op | Package |
|---|---:|---:|---:|---|
| `BenchmarkPatchOpsPoolRentFree` | **11.2 ns** | 0 B | 0 | `internal/page/patch` |
| `BenchmarkAnimCardPatched` | **94.7 ns** | 4 B | 0 | `internal/frame` |
| `BenchmarkAnimCardRedraw` | **63.4 µs** | 66.8 KB | 202 | `internal/frame` |
| `BenchmarkFrame_Dirty_InvalidateAndRedraw` | **146 ns** | 0 B | 0 | `test` |
| `BenchmarkSampleTransform` | **158 ns** | 63 B | 1 | `internal/motion` |
| `BenchmarkFrame_Tick_Idle` | **234 ns** | 0 B | 0 | `test` |
| `BenchmarkSampleOneEffect` | **382 ns** | 0 B | 0 | `internal/motion` |
| `BenchmarkSampleTenEffects` | **2.76 µs** | 0 B | 0 | `internal/motion` |
| `BenchmarkCaretLongRun/offsetIn` | **495 ns** | 1.79 KB | 1 | `internal/textrun` |
| `BenchmarkCaretLongRun/lineText` | **1.59 µs** | 10.2 KB | 5 | `internal/textrun` |
| `BenchmarkFrame_WarmRedraw_Card` | **499 ns** | 0 B | 0 | `test` |
| `BenchmarkFrame_WarmRedraw_Dashboard` | **512 ns** | 0 B | 0 | `test` |
| `BenchmarkFrame_Click_Dispatch` | **2.19 µs** | 496 B | 5 | `test` |

## v0.0.6 native Windows GPU baseline (2026-11-25)

Measured sequentially on native Windows (`goos: windows`, `goarch: amd64`, `CGO_ENABLED=0` pure-Go cross-compile from Linux with `-tags ownframe_debug`, Go `1.26.11`) on machine Chinmay:
- **CPU:** 13th Gen Intel(R) Core(TM) i7-13700HX
- **GPUs / Displays:** NVIDIA GeForce RTX 4060 Laptop GPU (1920×1080 @ 143 Hz) and Intel(R) UHD Graphics (2560×1600 @ 165 Hz)

### Live native Windows GPU windows (`-perf -debug-addr`, 8 s per app)

| App | Window | Present FPS | Avg frame | p95 frame | p99 frame | Long frames (>25 ms) | Go heap (live) | Windows RSS | CPU (1-core eq.) |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|
| `animations.exe` (`-dev=false`, 534 ops, active CSS animations) | 1140×900 | **325.3 FPS** (319 repaints/s) | **2.92 ms** | **4.32 ms** | **4.83 ms** | **0** | 20.8 MB | 359.9 MB | 162.3% |
| `wispr-flow-dashboard.exe` (398 ops, 31 images) | 1638×950 | **442.3 FPS** (6 repaints/s) | **0.75 ms** | **1.12 ms** | **7.45 ms** | **0** | 21.4 MB | 344.3 MB | 28.9% |
| `spotify-player.exe` (169 ops, 46 images, timer-paced) | 1440×960 | **68.2 FPS** | **3.57 ms** | **5.35 ms** | **9.59 ms** | **0** | 22.0 MB | 265.5 MB | 18.6% |
| `departure-board.exe` (628 ops, idle-parked) | 1000×640 | **2.1 FPS** (parked) | **1.14 ms** | **1.48 ms** | **12.19 ms** | 1 | 17.7 MB | 224.5 MB | **4.1%** |

Notes:
- Windows RSS includes DirectX/GPU driver shared memory pages; live Go heap stays between **17.7 MB and 22.0 MB** across all four apps.
- `animations.exe` ran uncapped at 319 repaints/s (`713 MB/s` alloc rate). Profiling `BenchmarkAnimationTick` (`466 µs/op`, `226 KB/op`, `180 allocs/op`) with `go tool pprof` attributes **96.1%** of allocated bytes to `applyRestamp`, led by `blinkless/internal/layout.stampStateAt` (**67.8%** alloc bytes, **58.9%** CPU) walking the whole box tree per animated target instead of only the target's ancestor path.

### Native Windows Ebiten replay benchmarks (`replay.test.exe -test.run NONE -test.bench . -test.benchmem -test.benchtime 300ms`)

| Benchmark | Windows time/op | Bytes/op | Allocs/op | Notes |
|---|---:|---:|---:|---|
| `BenchmarkDrawRectScan/dirty` | **185 ns** | 0 B | 0 | Op culling scan (dirty rect) |
| `BenchmarkDrawRectScan/full` | **183 ns** | 0 B | 0 | Op culling scan (full frame) |
| `BenchmarkChunkCardDamage/direct` | **4.22 µs** | 2.90 KB | 20 | 5 ops/draw |
| `BenchmarkChunkCardDamage/chunked` | **2.00 µs** | 849 B | 8 | 1 blit/draw (**2.1× faster**) |
| `BenchmarkBlitRaster` | **406 ns** | 581 B | 4 | Cached raster blit |
| `BenchmarkCachedRoundedFill_Warm` / `Cold` | **668 ns** / 5.36 µs | 596 B / 5.97 KB | 4 / 43 | **8.0× faster** warm |
| `BenchmarkCachedStroke_Warm` / `Cold` | **605 ns** / 3.45 µs | 588 B / 4.33 KB | 4 / 31 | **5.7× faster** warm |
| `BenchmarkText_NoSpacing_Warm` / `Direct` / `Cold` | **612 ns** / 18.5 µs / 23.5 µs | 623 B / 18.9 KB | 4 / 91 | **30× faster** warm vs. direct |
| `BenchmarkText_LetterSpacing_Warm` / `Direct` / `Cold` | **902 ns** / 40.7 µs / 75.8 µs | 540 B / 29.6 KB | 4 / 215 | **45× faster** warm vs. direct |
| `BenchmarkCompositeLayer_Normal` | **463 ns** | 536 B | 3 | Standard layer composite |
| `BenchmarkCompositeLayer_MixBlend_Multiply` / `Screen` | **1.93 µs** / **1.82 µs** | 2.11 / 2.51 KB | 18 | Kage mix-blend shader |
| `BenchmarkLayerDrawLayers/ops_100` / `500` / `1000` | **1.99** / **3.95** / **4.84 µs** | 2.14 KB | 12 | Retained layer draw |
| `BenchmarkLayerDrawLayers/blend_mode` | **4.03 µs** | 5.67 KB | 27 | Retained layer with blend |
| `BenchmarkTextDamageRedraw/warm` / `scroll` / `miss` | **37.4 µs** / **48.6 µs** / 1.44 ms | 27.3 KB / 1.14 MB | 200 / 5,162 | Text damage cache hit vs. miss |
| `BenchmarkLayerRotatedCard` | **8.56 µs** | 2.04 KB | 11 | 1 blit/frame after 40-op bake |

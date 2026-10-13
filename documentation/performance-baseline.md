# Performance baseline

One table per release. Fill it on the same machine when you can, and label every entry with the commit hash. Numbers from different machines do not compare.

Full procedure lives in [performance.md](performance.md).

## How to fill it

Run the automated script first. It exercises each benchmark app in `examples/perf-benchmarks` and prints the rows below.

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

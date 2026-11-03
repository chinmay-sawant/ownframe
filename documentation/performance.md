# Performance

The native desktop default is uncapped for both rendering and input updates.
Updates follow the rendering cadence rather than a separate 60 TPS schedule.
An explicit `WindowOptions.MaxFPS` still selects a cap. Clean pages retain
their pixels and park until input or a declared frame deadline. Uncapped
rendering does not guarantee a minimum frame rate; hover layout, GPU work,
and the display environment must still fit the frame budget.

## Opt in first

Perf sampling is off by default while DevTools is closed. Opening DevTools
enables page timing, frame sampling, and runtime memory sampling, so its
PERFORMANCE, LAST REDRAW DETAIL, and MEMORY rows populate automatically. Closing
the dock stops that automatic sampling unless the app also opts in. `/debug/state`
requires a debug build and `ServeOptions{InsecureAllowDebugState: true}`. Add
`InsecureAllowPprof: true` to include server runtime memory and enable profiler
endpoints.

```go
page, err := ownframe.NewWithOptions(ownframe.Config{HTML: html, Width: 720, Height: 560}, ownframe.WithPerf(true))
```

That one call covers the common case when DevTools is closed. The pieces behind it are `Config.Perf` on the page, `WindowOptions.Perf` on the window, and `ServeOptions{InsecureAllowPprof: true}` for `ServeWithOptions` and its `/debug/pprof/*` endpoints. `Page.Window().SetPerf` and `Page.Window().Perf` flip and read the page flag at runtime. `WindowOptions.Perf` turns on frame and runtime sampling even for a custom screen. The benchmark apps under `examples/perf-benchmarks` opt in already. With Perf on, the redraw alloc counter samples about twice a second and other stage timers still record every redraw.

This page turns the release checklist into steps you can run. It assumes Perf is on and the DevTools Frame tab and `GET /debug/state` expose the counters named below. If a counter is missing, file it as a bug instead of guessing.

Record per release numbers in [performance-baseline.md](performance-baseline.md). The [Spotify player report](../examples/spotify-player/PERFORMANCE.md) records redraw measurements for all eight screens. Repaint internals live in [repaint.md](repaint.md). The Frame tab layout lives in [devtools.md](devtools.md).

## 1 Idle baseline

Leave the window alone for 30 to 60 seconds and confirm it does almost nothing.

A desktop window that has drawn clean frames and seen no input parks in the
event wait, so an idle page stops presenting. The tick still runs on a
keep-alive frame about eight times a second, and input wakes the full rate.
The Frame tab's FPS and TPS rows show the paced rate while a page sits still;
that is expected, not a stall.

The idle helper prints a manual desktop checklist. It neither finds a process nor measures idle CPU. Build the named app, capture the binary PID, and read CPU and RSS for that exact process after warmup. Leave it untouched, without animation or reload.

```sh
sh scripts/perf-idle.sh --example examples/perf-benchmarks/normal
```

A quiet page should sit near zero CPU. RSS should hold flat across the sample window. Write both numbers into the baseline table under Idle RSS and Idle CPU, with the machine name and the commit hash. Repeat on the same machine before comparing releases.

## 2 Frame budget and pipeline

A 60 fps window has 16.67 ms per frame. The PERFORMANCE section of the Frame tab breaks each frame into stages so a slow frame points at its cause.

The page reports `LastTemplate` for template execution, `StyleTime` for CSS
styling, `LayoutTime` for placement and display-list emission, and `PaintTime`
for software bitmap painting. Cache hits still run CSS state/viewport styling
and placement. `OtherTime` covers form synchronization, display diff,
callbacks, and bookkeeping. These stages sum to `LastRedraw`.

The frame ring reports average, p95, and p99 elapsed Update and Draw work over
120 frames. It sums all Updates since the previous Draw and the current Draw,
including overlay drawing. Idle waits between calls are excluded. Long frames
exceed one sixtieth of a second. This measures host work, not GPU completion or
presentation latency. Open it with F12 or set `Config.DevTools` to true at startup.

A healthy normal page keeps p99 under budget with margin. When p95 passes budget, read the stage rows before changing code. Layout cost points at the engine relayout path. Paint cost points at the replay path. Template cost points at the app.

`GET /debug/state` carries page pipeline stats and runtime memory in web mode.
It has no native window frame ring. Native percentiles are read from DevTools.

## 3 Incremental repaint goal

A small change should do small work. One click edits one box, and the window repaints only the rects the page reports.

The LAST REDRAW DETAIL group of the Frame tab proves it. With Perf on, it reports dirty ops, dirty regions, and changed ops. The page reports the changed regions with `page.Probe().TakeDirtyRects` and the window paints them through `drawReplayPartial`, either into a content-sized tile grid or into one window-sized viewport buffer (`replay.DrawVisibleUnfixedZ`). A frame with no dirty rect blits the buffer it already has. `replay.DrawRect` is a test helper, not the window path. [repaint.md](repaint.md) has the list rules, the coalescing gap, the 24-rect cap, the one-third area collapse, pinned chrome painted after the tile blit, and the cases that still repaint in full, such as bitmap fallback pages. Tick callbacks automatically report operation damage; a clean ticking page can remain idle.

For a single control change, expect one dirty rect, a small repainted count, and a large skipped count. Two far-apart display changes report two diff regions and their summed CSS-pixel
area. Dirty ops count each operation meeting any region once. These counters
precede host padding, coalescing, and tile replay, so they do not count GPU work. Many small changes inside one region stay partial, so a whole frame rect on an edit now means the change really covered more than a third of the frame or the page could not place it; a list longer than 24 rects, or a rect the size of the whole content, makes the window rebuild its buffer once. A rebuild still only tessellates each repeated rounded shape once, because the replay bakes repeated fills and borders into a raster cache ([repaint.md](repaint.md)).

## 4 Soak matrix

For a display-free page workload, run `make soak-check`. It runs three scenarios for 600 seconds: interaction and conditional content;
theme, image, shaped text and ticking; and navigation. JSON records cycle,
Redraw and PNG p95 separately for each scenario. Defaults reject Redraw p95
over 16.67 ms, PNG p95 over 100 ms, retained heap growth over 64 MiB, and more
than four extra goroutines. Override timing limits with
`OWNFRAME_SOAK_REDRAW_P95_MS` and `OWNFRAME_SOAK_PNG_P95_MS` for a documented
machine budget. These are software workload thresholds, not native frame-rate
guarantees. GPU replay and native IME are outside this test.

For an actual window, drive the scenarios below for 10 to 30 minutes. Web sampling is separate: start an explicitly enabled server first, then sample it from another terminal.

```sh
go run -tags ownframe_debug ./examples/perf-benchmarks/normal -web -addr 127.0.0.1:8128
```

```sh
sh scripts/perf-longrun.sh --minutes 15 --url http://127.0.0.1:8128/debug/state
```

The sampler writes CSV and JSON metadata before sampling and preserves failure
status on errors, SIGTERM, and interruption. SIGKILL leaves initial metadata
marked running. It fails on missing metrics, request errors, counter resets,
configured heap/goroutine growth, or Linux RSS growth above 20 MiB/hour after a
30-second warmup. RSS regression requires three samples spanning at least 60
seconds. `--max-rss-growth-mib-hour 0` explicitly disables that budget on hosts
without RSS support. It does not drive interactions or measure desktop frame pacing. Cover each row at least once per release.

| Scenario | What to do |
|---|---|
| Resize | Drag the window edge slowly, pause, release, repeat |
| Scroll | Wheel through a tall page top to bottom and back |
| Click | Toggle one control every few seconds |
| Text | Type into one field, select, delete, repeat |
| Animate | Run the flappy benchmark with its tick on |
| Create and remove | Add and remove repeated rows or cards |

The MEMORY section of the Frame tab shows allocation per frame, Go heap, RSS, and goroutine count. The run passes when RSS and heap stay flat after warmup, GC cadence looks steady, and goroutines do not climb. A slow RSS climb across the whole run is a leak until proven otherwise. Capture a heap profile before and after when the numbers move.

## 5 DevTools perf panel fields

The Frame tab groups perf counters under three sections. Page pipeline values come from `page.Probe().Stats` through `host.Inspector`.
Frame cost and native RSS belong to the window. The web runtime snapshot
reports its own process memory; it does not expose the window frame ring.

| Section | Rows |
|---|---|
| PERFORMANCE | Frame average, p95, p99, long frame count |
| LAST REDRAW DETAIL | Template, style, layout, paint and other work; dirty ops, regions, area and changed ops |
| MEMORY | Alloc per frame, Go heap, RSS, goroutine count |

WINDOW, RENDERING, PIPELINE, and RELOAD rows stay as documented in [devtools.md](devtools.md). Perf sections add timing and memory on top of those counters, they do not replace them.

## 6 Benchmark apps

`examples/perf-benchmarks` holds three small apps with one job each. Each runs on desktop and with `-web`.

| App | What it stresses | Run |
|---|---|---|
| normal | A typical page with controls and text | `go run ./examples/perf-benchmarks/normal` |
| large | A heavy page with many boxes and rules | `go run ./examples/perf-benchmarks/large` |
| flappy | A ticking page that edits operations per frame | `go run ./examples/perf-benchmarks/flappy` |

Use normal for the idle baseline and the soak default. Use large to find layout and paint cliffs. Use flappy to find per frame allocation and tick cost. The baseline table records all three side by side so a change that helps one and hurts another shows up.

The bigger demos take `-perf` too. `examples/departure-board`, `examples/spotify-player`, and `examples/wispr-flow-dashboard` each turn on the same sampling with `go run ./examples/<name> -perf`, so the Frame tab reports averages, percentiles, pipeline stages, and memory on a real app. With `-web` the same flag serves `/debug/pprof/*` from the page.

A running window can serve its own profiling without `-web`. Pass `-debug-addr` with a loopback address and a debug build, plus `-perf` for runtime and frame numbers:

```sh
go run -tags ownframe_debug ./examples/spotify-player -perf -debug-addr 127.0.0.1:8129
```

The window serves `GET /debug/state` with the Frame tab counters and `GET /debug/pprof/*` with CPU, heap, goroutine, and trace profiles from the live process. The flag defaults to empty, which serves nothing. Non-loopback addresses are rejected.

## 7 Baseline table per release

Fill [performance-baseline.md](performance-baseline.md) once per release, on the same machine when possible. Automated rows come from this script.

```sh
sh scripts/perf-baseline.sh
```

The script runs display-free Go benchmarks for redraw stages, replay scans,
the platform click handler, and the Normal/Large/Flappy workloads. Five samples
per benchmark produce raw text and a sibling JSON report with source, machine,
Go environment, count, benchtime, min, median and sample p95. Five samples do
not establish a reliable tail estimate. Compare a compatible report with
`--baseline REPORT.json --max-ratio 1.3`; any median above that ratio fails.
CI compares parent and candidate on the same runner at 1.5, excluding the
corrected cold benchmark when comparing with older source. It does not launch benchmark apps or record idle CPU, native frame percentiles, or RSS. Enter those session measurements and machine notes separately. Never copy numbers across machines. Label every entry with the commit hash.

## Engine lock

`internal/render/engine.go` holds a process-wide mutex around engine parse
and apply. Two pages in one process queue on that lock. It is load-bearing
engine state, not a tuning knob. Cut parse work with the render cache instead
of narrowing the mutex.

## 8 Measure before optimizing

No optimization lands without a number that asks for it.

The flow is short. Reproduce the slow case in a benchmark app. Read the PERFORMANCE and LAST REDRAW DETAIL groups to name the stage. Take a profile, change one thing, then remeasure with the same script and app.

CPU and heap profiles come from a profiler endpoint. In web mode it is served only when the app opted into Perf with `ServeOptions{InsecureAllowPprof: true}`. In a window it is the `-debug-addr` listener, which needs `-tags ownframe_debug` and serves the live process, so no second `-web` run is owed.

```sh
go tool pprof -top http://127.0.0.1:8091/debug/pprof/heap
go tool pprof -top http://127.0.0.1:8091/debug/pprof/profile?seconds=30
```

Use the port the app actually serves. Keep the profile files next to the baseline entry when the change lands, and quote before and after numbers in the commit message.

## Static desktop idle

A static page without a tick callback or file watch parks until input, a page
notification, or a declared deadline arrives. Caret blinking and CSS motion
supply their own deadlines. `WantFrameAt`, redraws, element patches, and queued
scroll requests notify a parked host. Callbacks and file watches retain the
125 ms compatibility tick when they do not declare a deadline.

After five seconds of clean idle frames, the host releases renderer caches
and returns freed Go pages to the OS once. This also applies with DevTools
open. A static page can wake once for that trim, then stop requesting frames. Ebiten can still wake at roughly 1 Hz in the
validated X11 session. Those unsolicited wakes skip OwnFrame updates and
replay when input, focus, size, and the page are unchanged. The backend may
still copy and swap its retained screen.
The document, fonts, retained screen buffers, and Ebiten command buffers
remain resident as needed. The host does not modify Ebiten's internal buffer
capacities; a heap profile is needed before proposing a dependency change.
Short pauses do not force a collection.

To measure a running native window on Linux, use its compiled process PID:

```sh
python3 scripts/profile-idle.py --pid 12345 --seconds 30 --assert-static \
  --output temp/idle-profile/measurement.json
```

Close or pause DevTools and leave the window untouched. The script waits ten
seconds for settling, then records process CPU, RSS, OwnFrame draws, and page
generation changes. `--assert-static` rejects any observed OwnFrame draw or
generation change. An HTTP snapshot includes `sampled_at` and `sample_age_ms`;
its age grows while the window is parked, because reading it does not wake
the game loop. Runtime counters use the same cached sample as DevTools,
limited to one sample per two seconds while idle. Sampling does not create
a frame deadline. Xvfb provides repeatability with software rendering, but
its CPU and memory figures do not establish hardware GPU or power savings.

The native `presents` counter counts calls that draw OwnFrame content. It is
not a GPU swap counter. A zero delta proves that OwnFrame skipped rendering;
it does not prove that the window backend submitted no buffers.

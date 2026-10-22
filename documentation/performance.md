# Performance

## Opt in first

Perf sampling is off by default, so a shipped app pays nothing for it. The DevTools PERFORMANCE, PIPELINE DETAIL, and MEMORY rows show "-" and `/debug/state` (only when `ServeOptions{Debug: true}` or `{Perf: true}`) omits runtime until a developer opts in.

```go
page, err := ownframe.NewWithOptions(ownframe.Config{HTML: html, Width: 720, Height: 560}, ownframe.WithPerf(true))
```

That one call covers the common case. The pieces behind it are `Config.Perf` on the page, `WindowOptions.Perf` on the window, and `ServeOptions{Perf: true}` for `ServeWithOptions` and its `/debug/pprof/*` endpoints. `Page.Window().SetPerf` and `Page.Window().Perf` flip and read the page flag at runtime. The window picks up a page with Perf on by itself, and `WindowOptions.Perf` turns on frame and runtime sampling even for a custom screen. The benchmark apps under `examples/perf-benchmarks` opt in already.

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

The stages are template execute, layout, display list build, and paint or replay. The panel reports average, p95, and p99 frame time plus a count of long frames over budget. Open it with F12 or set `Config.DevTools` to true at startup.

A healthy normal page keeps p99 under budget with margin. When p95 passes budget, read the stage rows before changing code. Layout cost points at the engine relayout path. Paint cost points at the replay path. Template cost points at the app.

`GET /debug/state` carries the same fields in web mode for scripted capture.

## 3 Incremental repaint goal

A small change should do small work. One click edits one box, and the window repaints only the rects the page reports.

The PIPELINE DETAIL group of the Frame tab proves it. With Perf on, it reports dirty ops, dirty regions, and changed ops. The page reports the changed regions with `page.Probe().TakeDirtyRects` and the window paints them through `drawReplayPartial` tiles (`replay.DrawVisibleUnfixedZ`). A frame with no dirty rect blits the buffer it already has. `replay.DrawRect` is a test helper, not the window path. [repaint.md](repaint.md) has the list rules, the coalescing gap, the eight-rect cap, the one-third area collapse, pinned chrome painted after the tile blit, and the cases that still repaint in full, such as bitmap fallback pages. Tick callbacks automatically report operation damage; a clean ticking page can remain idle.

For a single control change, expect one dirty rect, a small repainted count, and a large skipped count. Two far-apart changes report two rects instead of the bounding box around both. Many small changes inside one region stay partial, so a whole frame rect on an edit now means the change really covered more than a third of the frame or the page could not place it; a list longer than eight rects, or a rect the size of the whole content, makes the window rebuild its buffer once. A rebuild still only tessellates each repeated rounded shape once, because the replay bakes repeated fills and borders into a raster cache ([repaint.md](repaint.md)).

## 4 Soak matrix

For a display-free page workload, run `make soak-check`. It drives resize, hit-tested clicks, typing, backspace, scrolling, conditional content, and PNG encoding for 600 seconds. Its JSON log records cycle p95, retained heap after GC, and goroutine counts. Coarse limits reject heap growth over 64 MiB or more than four extra goroutines. It does not exercise a native window or IME.

For an actual window, drive the scenarios below for 10 to 30 minutes. Web sampling is separate: start an explicitly enabled server first, then sample it from another terminal.

```sh
go run -tags ownframe_debug ./examples/perf-benchmarks/normal -web -addr 127.0.0.1:8128
```

```sh
sh scripts/perf-longrun.sh --minutes 15 --url http://127.0.0.1:8128/debug/state
```

The sampler writes CSV and JSON metadata and fails on missing metrics, request errors, counter resets, or configured heap/goroutine growth. It does not drive interactions or measure desktop frame pacing. Cover each row at least once per release.

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

The Frame tab groups perf counters under three sections. Values come from `page.Probe().Stats` through `host.Inspector`, so `GET /debug/state` returns the same data as JSON.

| Section | Rows |
|---|---|
| PERFORMANCE | Frame average, p95, p99, long frame count |
| PIPELINE DETAIL | Per stage times for template, layout, display list, and paint, then dirty ops, dirty regions, and changed ops |
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

## 7 Baseline table per release

Fill [performance-baseline.md](performance-baseline.md) once per release, on the same machine when possible. Automated rows come from this script.

```sh
sh scripts/perf-baseline.sh
```

The script runs display-free Go microbenchmarks for redraw stages, replay scans, and the platform click handler. It does not launch benchmark apps or record idle CPU, native frame percentiles, or RSS. Enter those session measurements and machine notes separately. Never copy numbers across machines. Label every entry with the commit hash.

## Engine lock

`internal/render/engine.go` holds a process-wide mutex around engine parse
and apply. Two pages in one process queue on that lock. It is load-bearing
engine state, not a tuning knob. Cut parse work with the render cache instead
of narrowing the mutex.

## 8 Measure before optimizing

No optimization lands without a number that asks for it.

The flow is short. Reproduce the slow case in a benchmark app. Read the PERFORMANCE and PIPELINE DETAIL groups to name the stage. Take a profile, change one thing, then remeasure with the same script and app.

CPU and heap profiles come from the profiler endpoint in web mode, served only when the app opted into Perf with `ServeOptions{Perf: true}`.

```sh
go tool pprof -top http://127.0.0.1:8091/debug/pprof/heap
go tool pprof -top http://127.0.0.1:8091/debug/pprof/profile?seconds=30
```

Use the port the app actually serves. Keep the profile files next to the baseline entry when the change lands, and quote before and after numbers in the commit message.

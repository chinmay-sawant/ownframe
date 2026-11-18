# v0.0.4 rendering performance evidence

This record supports [track 01](01-documentation-and-performance.md). It carries
measurements and interpretation, not a second task checklist.

## Phase 2 baseline

Source base: `0a77375ad10c51c8f9dab2d9c075701e7a2249bb`, with the local Phase 1
documentation and Phase 2 benchmark changes. The [JSON record](evidence/list-before.json)
contains the tracked-diff digest and hashes of untracked source files. The host
was WSL2 Linux on an Intel Core i7-13700HX, using Go 1.26.10 and the runtime's
default GOMAXPROCS of 24. No agent test, lint, or graphics commands overlapped
the capture. Other Windows and Linux host activity was uncontrolled.

From `/home/chinmay/ChinmayPersonalProjects/ownframe`:

```sh
sh scripts/perf-baseline.sh --suite list --conditions 'Local WSL2 session; no concurrent agent test, lint, or graphics commands; other host activity uncontrolled' --count 5 --benchtime 250ms --profiles --out temp/v004-performance/before.txt
go tool pprof -top -nodecount=25 temp/v004-performance/before.cpu
go tool pprof -top -alloc_space -nodecount=25 temp/v004-performance/before.alloc
```

All commands exited 0. [Raw samples](evidence/list-before.txt) retain all five
observations for each of the ten variants. The initial frame is committed before
timing. The fixture uses the existing generated dataset, 1,000 rows, a 40 px
header, 40 px rows, and an 800 by 600 viewport. Resize alternates widths 801 and
800. Scroll alternates offsets 4,000 and 8,000. Windowing retains four overscan
rows on each side through the existing `examples/list.Window` helper.

| Workload | Median with perf off | Median with perf on | Bytes/op with perf off | Allocs/op with perf off |
|---|---:|---:|---:|---:|
| **Unchanged redraw** | 528.7 ns | 568.9 ns | 0 | 0 |
| **Text update** | 30.543 ms | 31.056 ms | 12,950,595 | 84,331 |
| **Scroll-offset submission** | 3.127 ns | 3.113 ns | 0 | 0 |
| **Resize** | 25.706 ms | 26.486 ms | 9,362,379 | 59,189 |
| **Windowed scroll** | 0.836 ms | 0.900 ms | 392,709 | 2,415 |

Scroll-offset submission measures the host's report to the page and the retained
row geometry checks after timing. It measures no viewport replay, wheel input,
GPU presentation, or native scrolling latency. Windowed scroll includes the
changed row slice and page redraw. The unchanged case verifies no new display
generation. Each mutating redraw case requires one generation per iteration,
so silently skipped updates fail the benchmark. Geometry tests compare visible
windowed rows against the complete list at the top, middle, end, and scroll return.

## Profile interpretation and optimization budget

The separate two-second perf-off resize profile includes setup and profiler
overhead. Its percentages describe sampled stacks, not precise stage timings.
`buildBlock` accounts for 63.4% cumulative CPU samples, including `buildFlex` at
60.4%. These are nested paths and must not be added together. Blinkless
`workspaceOps` and `layout.newDisplay` lead flat allocated bytes. The engine owns
the largest remaining cost; a local OwnFrame optimization cannot claim to solve it.

The profile also includes template execution and form-source copying during a
size-only change. `internal/page/page_size.go` marks form data dirty on every
resize, even though viewport dimensions already participate in the layout key.
The local optimization candidate is to preserve clean template data through a
size-only change while retaining layout invalidation and row-window callbacks.
Prove that data changes, live form edits, and changed row slices still publish.

Budgets for the local experiment, set before that code change:

- Keep unchanged redraw at zero allocations and below 1 microsecond median.
- Keep windowed-scroll page work below 2 ms median and 500 KB/op on this host.
- Reduce perf-off resize allocation by at least 5% from the baseline.
- Keep resize and text-update median time within 15% of the comparable baseline.
- Require a clean template getter to run zero extra times during size-only changes.

These are local page-work budgets. They are not native frame guarantees or a
claim that the unvirtualized list meets a 60 Hz presentation budget. Before and
after runs must use the same workload runner, conditions, and instrumentation.
Binary profiles remain regenerable under `temp/`; their textual summaries and
sample provenance are retained here for review.

## Phase 3 comparison

The final change in `internal/page/page_size.go` keeps template data clean for a
size-only update on pages without an active viewport callback. Layout dimensions
still change, so media queries and geometry are recalculated. An active callback
keeps invalidating template data even when it returns false. Telegram's pin
callback uses that return value to suppress a scroll redraw while updating bar
positions, so treating false as proof of unchanged data would break rotation.

The clean-template regression failed before the change with one extra template
execution. It passes after the change. Separate checks prove that SetData,
callbacks returning true or false, caret, selection, and live field values still
work after resize. A software-picture test compares the resized page against a
newly constructed page at each target width, including return to the initial width.

After `make lint`, `make test`, and `make graphics-check` each exited 0, run:

```sh
sh scripts/perf-baseline.sh --suite list --conditions 'Local WSL2 session; no concurrent agent test, lint, or graphics commands; other host activity uncontrolled' --count 5 --benchtime 250ms --baseline temp/v004-performance/before.json --max-ratio 1.15 --out temp/v004-performance/after-final.txt
```

This command exited 0. [Raw after samples](evidence/list-after.txt) and
[JSON provenance](evidence/list-after.json) retain the results. Same-host local
medians, with performance instrumentation off:

| Workload | Before | After | Before bytes/op | After bytes/op | Before allocs/op | After allocs/op |
|---|---:|---:|---:|---:|---:|---:|
| **Unchanged redraw** | 528.7 ns | 529.5 ns | 0 | 0 | 0 | 0 |
| **Text update** | 30.543 ms | 31.107 ms | 12,950,595 | 12,959,362 | 84,331 | 84,336 |
| **Resize** | 25.706 ms | 23.356 ms | 9,362,379 | 8,268,253 | 59,189 | 34,412 |
| **Windowed scroll** | 0.836 ms | 0.883 ms | 392,709 | 392,707 | 2,415 | 2,415 |

Resize allocated about 11.7% fewer bytes and made about 41.9% fewer allocations.
Its median page-work time was about 9.1% lower in these samples. Perf-on resize
measured 23.096 ms and 8,262,529 bytes/op. All ten timing comparisons stayed within
the predeclared 15% regression bound. The zero-allocation unchanged case and the
windowed-scroll budgets also passed. These observations do not remove the engine
layout cost or establish native presentation performance.

Final gate output is retained for [lint](evidence/ownframe-lint.txt),
[tests](evidence/ownframe-tests.txt), and [graphics](evidence/ownframe-graphics.txt).
The graphics gate passed all 13 required tests. Binary profiles can be
regenerated with the Phase 2 command. No engine pin or engine source was changed
for this local improvement.

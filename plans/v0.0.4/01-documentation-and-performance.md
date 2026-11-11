# v0.0.4 / 01 - Documentation and rendering performance

> **Parent:** [v0.0.4 index](README.md)
> **Status:** all four phases complete for the scoped OwnFrame change. Engine CSS work belongs to track 04.
> **Estimated effort:** size the optimization after profiles identify the dominant cost. Documentation and baseline capture are the first slice.

---

## Overview

Correct the flex-absolute support claim and measure realistic rendering
workloads before changing the renderer.

## Executive Summary

The review of `085ba7ea` found that `TestFlexChildAbsoluteOverlays` expects
absolute positioning to remove a flex child from flow and apply its offsets,
while the public guides described that case as unsupported. Phase 1 corrects
those guides and records the current regression result below.

`examples/perf-benchmarks/large_test.go` repeatedly resizes an unvirtualized,
text-heavy 1,000-row page between widths 800 and 801 at height 600. Its constructor
in `benchutil/large.go` enables performance instrumentation. Preliminary review
samples measured 41 to 58 ms and about 9 MB allocated per operation. Those
samples are investigation context, not a controlled release baseline. Allocated
bytes are not retained memory, and forced relayout time is not native frame time.

## Phase 1: Correct the supported-behavior reference

### 1.1 Flex-absolute documentation

- [x] `DOC-01` Run `go test -p 1 -count=1 -run '^TestFlexChildAbsoluteOverlays$' -v ./internal/page` on the implementation base and retain its result. Evidence: the Phase 1 validation record below reports PASS and exit 0.
- [x] `DOC-02` Correct the flex-child absolute-positioning row in `documentation/features.md` to describe the proven behavior and its tested scope. Evidence: the row is in the supported table and links to the regression scope in `documentation/screen.md`.
- [x] `DOC-03` Remove the conflicting flex-absolute workaround claim from `documentation/screen.md` without claiming sticky support. Evidence: the positioning section describes the tested offsets and out-of-flow sibling placement; the sticky limitation stayed explicit at this checkpoint; track 04 subsequently replaces it with the verified viewport contract.
- [x] `DOC-04` Search public guides for the same obsolete flex-absolute claim and reconcile each occurrence against current tests. Evidence: the scan below also identified and corrected `documentation/chadcn.md` and `components/chadcn/README.md`.
- [x] `DOC-05` Run `make docs-check plans-check examples-check` and record the successful exit status after the documentation edits. Evidence: exit 0; 39 snippets compiled, 24 maintenance rows validated, and the example inventory passed.

### 1.2 Phase 1 validation record

Implementation base: `0a77375ad10c51c8f9dab2d9c075701e7a2249bb`.
Host: Linux WSL2 `6.6.87.2-microsoft-standard-WSL2`, Go `go1.26.10 linux/amd64`.
The changes in this phase affect documentation and checklist state only.

From `/home/chinmay/ChinmayPersonalProjects/ownframe`:

```sh
go test -p 1 -count=1 -run '^TestFlexChildAbsoluteOverlays$' -v ./internal/page
make docs-check plans-check examples-check
make test
git diff --check
```

Results: each command exited 0. The focused test reported
`PASS: TestFlexChildAbsoluteOverlays` and package success. Documentation checks
compiled 39 snippets and listed 12 external illustrations explicitly. Plan
validation accepted 24 parent-ledger rows. The example inventory and full test
command passed. Some full-suite package results reused Go's test cache.

The public-guide scan used:

```sh
rg -n -i '(flex.absolute|absolute.{0,90}flex|flex.{0,90}absolute|child stays in flow|top.*left.*do nothing)' documentation components examples frontend README.md AGENTS.md --glob '*.md' --glob '!**/node_modules/**' --glob '!**/build/**'
```

Review of the matches found the obsolete claim in four guides. All four now
describe or link to the tested scope. Remaining matches describe supported
positioning or unrelated absolute-positioning behavior. Generated `docs/`
bundles were not regenerated on this feature branch, as required by the website
publication policy. This phase makes no new sticky, rendering, or native-host
behavior claim. Phase 2 is the next unchecked phase.

## Phase 2: Build a repeatable workload baseline

### 2.1 Benchmark contracts

- [x] `PERF-01` Add comparable unchanged-redraw, single-row text-update, scroll, resize, and viewport-windowed list workloads under `examples/perf-benchmarks` using the same 1,000-row dataset. Evidence: `BenchmarkListWorkloads` uses `benchutil.NewListWorkload` for all ten variants; the legacy `BenchmarkLargeRedraw` is retained.
- [x] `PERF-02` Assert that each mutating workload changes its intended state or pixels so a skipped update cannot appear as a speedup. Evidence: retained text, display dimensions, exact generation counts, reported offsets, and visible-row geometry are asserted; `TestListWorkloads` and `TestListWindowGeometryMatchesFull` pass.
- [x] `PERF-03` Add instrumentation-off and instrumentation-on variants through the benchmark support constructor while retaining the existing benchmark's meaning. Evidence: `ListOptions.Perf` controls page timing; tests check the effective timing mode after the initial frame.
- [x] `PERF-04` Add a repeatable runner under `scripts/` that records samples, fixture parameters, cache state, instrumentation mode, source commit, Go version, OS, CPU, and concurrent workload conditions. Evidence: `scripts/perf_baseline.py --suite list` requires conditions, rejects incomplete variant inventories, and records provenance and commands in [the baseline JSON](evidence/list-before.json).
- [x] `PERF-05` Capture CPU and allocation profiles for the large workload with explicit output paths under `temp/`; inspect the actual expensive call stacks. Evidence: [CPU summary](evidence/list-before-cpu.txt), [allocation summary](evidence/list-before-alloc.txt), and [interpretation](performance-evidence.md).

Phase 2 validation: `go test -p 1 -count=1 -run '^TestList' ./examples/perf-benchmarks`,
`python3 -m unittest discover -s scripts -p 'test_perf_baseline.py'`, `make lint`,
and `make test` all exited 0 after the workload and runner changes. The runner
completed all ten variants with five samples each and exited 0. Its separate
profile capture uses two seconds, independent of the sample benchtime. Commands,
conditions, medians, and scope are retained in [performance evidence](performance-evidence.md).

The existing baseline command is:

```sh
go test -p 1 -run '^$' -bench '^BenchmarkLargeRedraw$' -benchmem -benchtime=1s -count=5 ./examples/perf-benchmarks
```

Run from the absolute OwnFrame checkout path. Setup commits the initial frame
before timing, so this command measures warm forced resize with instrumentation
enabled. Record that distinction. Add exact runnable commands for each new
variant when it exists. Measure native scrolling and frame latency separately
from page-only benchmarks.

## Phase 3: Optimize the measured bottleneck

### 3.1 Evidence-led implementation

- [x] `PERF-06` Identify the dominant cost among template execution, style recalculation, layout, display construction, and OwnFrame commit processing, with profile evidence. Evidence: [profile interpretation](performance-evidence.md#profile-interpretation-and-optimization-budget) identifies engine layout and display buffers as dominant, plus avoidable OwnFrame template execution during resize.
- [x] `PERF-07` Set workload budgets from the baseline and the intended interaction before selecting an optimization; state whether a budget applies to page work or native presentation. Evidence: the local page-work budgets in `performance-evidence.md` were recorded before the size-only invalidation change.
- [x] `PERF-08` Implement the smallest change addressing the measured cost in its owning package; document a separate Blinkless dependency when the expensive code belongs to the engine. Evidence: `internal/page/page_size.go` preserves clean template output on size-only changes without an active viewport callback. Callback pages retain invalidation, including Telegram's false-returning pin callback. The engine remains the dominant cost and its separate ownership is recorded in track 04.
- [x] `PERF-09` Compare repeated before-and-after samples on the same machine and instrumentation mode, reporting timing, bytes allocated, and allocation counts. Evidence: [comparison](performance-evidence.md#phase-3-comparison), [after samples](evidence/list-after.txt), and [after provenance](evidence/list-after.json); all predeclared local budgets passed.
- [x] `PERF-10` Verify changed-workload pixels and unchanged-frame behavior, including scroll return and resize, against a fresh rendering reference. Evidence: `TestResizePictureMatchesFreshPage`, clean-template and callback regressions in `internal/page/page_cache_test.go`, `TestCaretAndSelectionSurviveResize`, row-window geometry comparisons, and all 13 graphics checks passed.

Do not choose pooling, concurrency, or another cache before profiles justify
that mechanism. Windowing is an explicit workload strategy, not evidence that
the unvirtualized renderer became faster. A 16.7 ms comparison is meaningful
only for an interaction that must fit a 60 Hz frame budget.

## Phase 4: Close with retained evidence

### 4.1 Gates

- [x] `PERF-11` Run `make lint` on the final implementation and record its exit status. Evidence: exit 0; [retained output](evidence/ownframe-lint.txt).
- [x] `PERF-12` Run `make test` on the final implementation and record its exit status. Evidence: exit 0; [retained output](evidence/ownframe-tests.txt), including Telegram rotation and keyboard regressions.
- [x] `PERF-13` Run `make graphics-check` after paint or window changes and record all required test results; record an explicit scope reason if no such change is made. Evidence: exit 0; all 13 required graphics tests passed in [retained output](evidence/ownframe-graphics.txt).
- [x] `PERF-14` Publish the comparison commands, samples, profile findings, budgets, and remaining limits in a durable performance record linked from this file. Evidence: [performance evidence](performance-evidence.md) and its linked raw records.

## Dependencies

- Track 04 owns support claims for height queries and sticky after engine adoption.
- Preserve software and native rendering behavior. Do not infer native latency,
  memory leaks, or cross-platform performance from this page-only benchmark.

# v0.0.4 / 01 - Documentation and rendering performance

> **Parent:** [v0.0.4 index](README.md)
> **Status:** planned. Review observations are inputs, not completed work.
> **Estimated effort:** size the optimization after profiles identify the dominant cost. Documentation and baseline capture are the first slice.

---

## Overview

Correct the flex-absolute support claim and measure realistic rendering
workloads before changing the renderer.

## Executive Summary

At `085ba7ea`, `TestFlexChildAbsoluteOverlays` expects absolute positioning to
remove a flex child from flow and apply its offsets. `documentation/features.md`
and `documentation/screen.md` still describe that case as unsupported.

`examples/perf-benchmarks/large_test.go` repeatedly resizes an unvirtualized,
text-heavy 1,000-row page between widths 800 and 801 at height 600. Its constructor
in `benchutil/large.go` enables performance instrumentation. Preliminary review
samples measured 41 to 58 ms and about 9 MB allocated per operation. Those
samples are investigation context, not a controlled release baseline. Allocated
bytes are not retained memory, and forced relayout time is not native frame time.

## Phase 1: Correct the supported-behavior reference

### 1.1 Flex-absolute documentation

- [ ] `DOC-01` Run `go test -p 1 -count=1 -run '^TestFlexChildAbsoluteOverlays$' -v ./internal/page` on the implementation base and retain its result.
- [ ] `DOC-02` Correct the flex-child absolute-positioning row in `documentation/features.md` to describe the proven behavior and its tested scope.
- [ ] `DOC-03` Remove the conflicting flex-absolute workaround claim from `documentation/screen.md` without claiming sticky support.
- [ ] `DOC-04` Search public guides for the same obsolete flex-absolute claim and reconcile each occurrence against current tests.
- [ ] `DOC-05` Run `make docs-check plans-check examples-check` and record the successful exit status after the documentation edits.

## Phase 2: Build a repeatable workload baseline

### 2.1 Benchmark contracts

- [ ] `PERF-01` Add comparable unchanged-redraw, single-row text-update, scroll, resize, and viewport-windowed list workloads under `examples/perf-benchmarks` using the same 1,000-row dataset.
- [ ] `PERF-02` Assert that each mutating workload changes its intended state or pixels so a skipped update cannot appear as a speedup.
- [ ] `PERF-03` Add instrumentation-off and instrumentation-on variants through the benchmark support constructor while retaining the existing benchmark's meaning.
- [ ] `PERF-04` Add a repeatable runner under `scripts/` that records samples, fixture parameters, cache state, instrumentation mode, source commit, Go version, OS, CPU, and concurrent workload conditions.
- [ ] `PERF-05` Capture CPU and allocation profiles for the large workload with explicit output paths under `temp/`; inspect the actual expensive call stacks.

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

- [ ] `PERF-06` Identify the dominant cost among template execution, style recalculation, layout, display construction, and OwnFrame commit processing, with profile evidence.
- [ ] `PERF-07` Set workload budgets from the baseline and the intended interaction before selecting an optimization; state whether a budget applies to page work or native presentation.
- [ ] `PERF-08` Implement the smallest change addressing the measured cost in its owning package; document a separate Blinkless dependency when the expensive code belongs to the engine.
- [ ] `PERF-09` Compare repeated before-and-after samples on the same machine and instrumentation mode, reporting timing, bytes allocated, and allocation counts.
- [ ] `PERF-10` Verify changed-workload pixels and unchanged-frame behavior, including scroll return and resize, against a fresh rendering reference.

Do not choose pooling, concurrency, or another cache before profiles justify
that mechanism. Windowing is an explicit workload strategy, not evidence that
the unvirtualized renderer became faster. A 16.7 ms comparison is meaningful
only for an interaction that must fit a 60 Hz frame budget.

## Phase 4: Close with retained evidence

### 4.1 Gates

- [ ] `PERF-11` Run `make lint` on the final implementation and record its exit status.
- [ ] `PERF-12` Run `make test` on the final implementation and record its exit status.
- [ ] `PERF-13` Run `make graphics-check` after paint or window changes and record all required test results; record an explicit scope reason if no such change is made.
- [ ] `PERF-14` Publish the comparison commands, samples, profile findings, budgets, and remaining limits in a durable performance record linked from this file.

## Dependencies

- Track 04 owns support claims for height queries and sticky after engine adoption.
- Preserve software and native rendering behavior. Do not infer native latency,
  memory leaks, or cross-platform performance from this page-only benchmark.

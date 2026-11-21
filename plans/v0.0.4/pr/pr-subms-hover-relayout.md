# PR: Sub-millisecond Hover Performance (`blinkless v0.1.4` + `ownframe` Dirty/Motion/Hover Optimizations)

## Summary

Brings `loglens` hover latency below 1 ms and cuts per-hover allocations by ~88% across both edge-region hover (with `display: none <-> display: block` tooltips) and command-palette hover, using 100% generic improvements across `blinkless v0.1.4` and `ownframe`.

## Benchmark Comparison (`loglens`, baseline `v0.1.3` -> current `v0.1.4`)

| Benchmark | Metric | Baseline | Current | Improvement |
|---|---|---|---|---|
| `BenchmarkEdgeRegionHover` | Time/op | 2.08 ms | **0.80 ms** | **-62%** (~2.6× faster) |
| `BenchmarkEdgeRegionHover` | B/op | 1.74 MB | **203 KB** | **-88%** |
| `BenchmarkEdgeRegionHover` | Allocs/op | 3,455 | **1,047** | **-70%** |
| `BenchmarkCommandPaletteHover` | Time/op | 1.77 ms | **0.66 ms** | **-62%** (~2.7× faster) |
| `BenchmarkCommandPaletteHover` | B/op | 1.89 MB | **221 KB** | **-88%** |
| `BenchmarkCommandPaletteHover` | Allocs/op | 3,350 | **892** | **-73%** |

## Changes

### 1. `blinkless v0.1.4` Bump (`go.mod`, `examples/go.mod`)
- Bumped both root `go.mod` and `examples/go.mod` to `github.com/chinmay-sawant/blinkless v0.1.4` (picking up `SameBoxes` incremental box-list patching, cross-relayout flex override & style resolution memo caching, pooled clone/batch buffers, and pointer-receiver `Op.SameExtra` / `Op.MixExtraStamp` / `Op.HasExtra` / `Op.HasBoundExtra`).

### 2. Dirty Diff & KidIndex Optimizations (`internal/page/dirty/diff.go`, `internal/page/dirty/op.go`, `internal/page/dirty/boxes.go`, `internal/page/dirty.go`)
- Bit-packed discrete op flags into a single FNV-1a word in `mixOpStamp`, fast-pathed zero optional floats (`LetterSpacing`, `InkDescent`, 10 `Radius*` fields), and delegated rare-payload comparison/hashing to `Op.SameExtra` / `Op.MixExtraStamp` to eliminate 13 × 256-byte value-receiver struct copies per op.
- Added `KidIndex.Rebuild(boxes)` reusing existing `byID`, `head`, `link`, and `rects` backings in place.

### 3. Hover Guard & Transition Capture (`internal/page/render.go`, `internal/page/patch.go`, `internal/motion/runtime.go`, `internal/motion/transition.go`)
- Fast-pathed `hoverGuardHash` using `(sourceRev, sourceHash, sourceLen, themeSrc)` header equality to avoid re-hashing the full HTML+CSS document on every mouse move.
- Reused `Runtime` scratch maps (`capIDs`, `capOrdinals`, `capOldBoxes`, `capRecords`) in `CaptureTransitions`, added an aligned-boxes O(1) pairing fast path, iterated declared transition properties via a 34-bit bitmask, filtered `indexBoxes` in `BindTransitions` / `AdoptTransitions`, and checked paint equality inline before allocating `[]Paint` slices.

### 4. Regression Tests (`internal/page/dirty/diff_alloc_test.go`)
- Added `TestKidIndexRebuildMatchesNewAndReusesBacking` (asserting 0 steady-state allocs) and `TestSameOpAndStampEmptyVsBoundExtra`.

## Verification

- `make lint` (all 9 checks including `golangci-lint`, docs, plans, repo-size, examples, internal imports, ui properties, website ESLint): **PASS**
- `make test`: **PASS**
- `make graphics-check` (all 19 GPU pixel tests): **PASS**
- `go vet -p 1 ./examples/builder/loglens`: **PASS**

# OwnFrame v0.0.10

Release on top of `v0.0.9`, delivering real multi-line `<textarea>` text editing with visual-line navigation, the `cmd/ownframegen` compile-time HTML-to-Go code generator CLI (28.2x faster HTML generation with 1 alloc/op and 73x faster typed `PatchBuilder` updates), and two new interactive examples (`examples/retained-dom` and `examples/codegen`).

- Range: `v0.0.9` through `6707d14a` (2026-11-28)
- Engine: pins `github.com/chinmay-sawant/blinkless` to `v0.1.11`

## Summary

| Area | What changed |
| :--- | :--- |
| **Multi-line `<textarea>` Editing (`internal/textrun`, `internal/page`, `internal/window`, `page_edit.go`)** | Reconstructed visual lines (hard newlines and soft-wrapped runs, including leading, middle, and trailing empty lines) in `internal/textrun/visual.go`. Wrapped `<textarea>` hard-newline segments in `<span data-ownframe-line="1">` so empty lines keep their height under `pre-wrap`. Added `ArrowUp`/`ArrowDown` (+ `Shift+`) vertical navigation with preserved `desiredX` across short/empty lines, line-aware `Home`/`End` (+ `Shift+`), document `Ctrl+Home`/`Ctrl+End`, triple-click line selection, `Enter` newline insertion (`Ctrl/Cmd+Enter` submits), and delegated `Page.Edit().Selection() (caret, start, end int)`. |
| **Compile-time HTML-to-Go Codegen (`cmd/ownframegen`, `documentation/codegen.md`)** | Added zero-dependency CLI `cmd/ownframegen` (`-in`, `-out`, `-pkg`, `-type`, `//go:generate` ready). Compiles `.html` templates into `go/format`-canonical Go with fluent builder structs (`New...().With...()`), zero-reflection `RenderTo` / `Config.Render` (all 688 Blinkless CSS properties and HTML tags emitted verbatim in 1 allocation), compile-time `ID*` and `Action*` constants, `{{raw ...}}` slots for `ui.Element` builders, and fluent `PatchBuilder` (`view.Patch()...Apply(ctx, page)`). |
| **Examples & Documentation (`examples/input-lab`, `examples/codegen`, `examples/retained-dom`, `documentation/`)** | Added interactive multi-line `<textarea>` demo with live caret/line:col/selection readout in `examples/input-lab`. Added `examples/codegen` (port `8152`) with headless tests and `html/template` vs codegen benchmarks. Added `examples/retained-dom` v0.0.9 visualization demo. Added `documentation/codegen.md` (with 4-step `counter.html` quickstart) and updated `documentation/forms.md`, `documentation/editing.md`, `documentation/keys.md`, and `documentation/features.md`. |

## Measured Impact (`examples/codegen`, Linux amd64, Go 1.26.11)

| Benchmark / Workload | `html/template` / Full Redraw | `ownframegen` (`RenderTo` / `PatchBuilder`) | Improvement |
| :--- | ---: | ---: | ---: |
| HTML generation (50 rows) | `191.3 µs` / `26.3 KB` / `1,322 allocs` | **`6.8 µs`** / **`18.4 KB`** / **`1 alloc`** | **`28.2x faster`**, **`1,322x` fewer allocs** |
| HTML generation (200 rows) | `686.2 µs` / `103.3 KB` / `5,222 allocs` | **`26.9 µs`** / **`65.5 KB`** / **`1 alloc`** | **`25.5x faster`**, **`5,222x` fewer allocs** |
| HTML generation (1,000 rows) | `3,573.4 µs` / `520.6 KB` / `26,767 allocs` | **`126.8 µs`** / **`270.3 KB`** / **`1 alloc`** | **`28.2x faster`**, **`26,767x` fewer allocs** |
| 200-row update (1 element) | `4.24 ms` / `1.45 MB` / `9,351 allocs` | **`58.3 µs`** / **`62.0 KB`** / **`72 allocs`** | **`73x faster`**, **`130x` fewer allocs** |
| 200-row update (10 elements batched) | `4.24 ms` / `1.45 MB` / `9,351 allocs` | **`317.2 µs`** / **`292.6 KB`** / **`924 allocs`** | **`13.4x faster`**, **`10x` fewer allocs** |

## Breaking changes

None. Single-line `<input>` submit-on-Enter and direct programmatic `Page.Edit().Submit(ctx)` keep prior behavior.

## Validation

- `make lint` (all 10 checks)
- `make graphics-check` (all GPU tests)
- `make test TEST_P=1` (all library and example packages)
- PR #126 CI (all 19 checks green, including `ci/test-race`, `ci/sec`, `ci/android-bind`, `ci/module`, `ci/performance`, `ci/site`)

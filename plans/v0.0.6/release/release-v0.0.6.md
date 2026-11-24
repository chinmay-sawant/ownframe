# OwnFrame v0.0.6

Release on top of `v0.0.5` (`e50a6e76`), completing native vector replay across all static CSS properties in `examples/blinkless-all`, adopting Blinkless `v0.1.8`, and optimizing incremental `PatchElement` adoption.

- Range: `v0.0.5` (`e50a6e76`) through `chore/improvement2` (2026-11-25)
- Engine: pins `github.com/chinmay-sawant/blinkless` to `v0.1.8` (merged in `blinkless#55`)

## Summary

| Area | What changed |
| --- | --- |
| **Native Vector Replay & Masking** | Added nearest-neighbor `image-rendering` (`pixelated`, `crisp-edges`, `optimizespeed`), native `background-blend-mode` (11 separable modes via Kage), tab expansion in both Ebiten replay and software bitmap painter, two-pass Gaussian `filter: blur(...)` on buffered groups, `isolation: isolate` layer buffering, `mask-size: cover/contain`, 3-/4-value `mask-position`, `mask-origin`/`mask-clip` box insets, OpenType-to-BCP47 shaping language mapping, and horizontal unprojected `FakeOblique` replay in `TextMatrix`. |
| **Blinkless v0.1.8 Engine Adoption** | Pulls in Blinkless `v0.1.8` CSS layout/grid/anchor/SVG/filter/mask-border/text fixes, retained table scroll range fix, subtree relayout performance improvements, and MIT-licensed `Blinkless OT`, `Blinkless Var`, and `Blinkless Palette` demo faces. |
| **CSS Conformance Showcase (`examples/blinkless-all`)** | Cleared all 79 static `partial` overrides (`painterDeclines`, `opOnly`, `noPaint` all empty), promoted 98 additional dynamically-`partial` demos to `implemented` (587 / 672 generated rows visibly distinct from control; 0 bitmap fallbacks across all 49 tabs), and lowered `genPageRows` to 16. |
| **Incremental Patching Performance** | Optimized `PatchElement` adoption and added scaling/patch benchmarks in `examples/perf-benchmarks`. |

## Breaking changes

None. Public root API and exported signatures are unchanged.

## Validation

- `make lint` (Go library + examples, docs, plans, tracked file size, examples inventory, internal imports, ESLint)
- `make graphics-check` (all window, GPU replay, and example pixel checks)
- `go test ./internal/masking ./internal/render ./internal/replay ./internal/textrun ./examples/blinkless-all`

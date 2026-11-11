# v0.0.4 / 04 - CSS conformance and engine adoption

> **Parent:** [v0.0.4 index](README.md)
> **Status:** complete. Published engine adopted; geometry, native pixel/hit, browser reference, and all closure gates passed.
> **Estimated effort:** estimate height queries and sticky separately after engine investigation. Sticky requires native scrolling validation.

---

## Overview

Implement height media queries and sticky positioning in their owning engine,
then adopt the published engine revision and prove OwnFrame integration.

## Executive Summary

At `085ba7ea`, `TestMediaHeightFollowsTheFrame` expects a min-height query to
remain unmatched. `TestStickyScrollsWithWindow` expects a sticky header to lack
fixed operations and translate with the page. These tests preserve known limits.
Replace their expectations only when the corresponding implementation lands.

The immediate flex-absolute documentation correction belongs to track 01.
This track owns claims about newly delivered height-query and sticky support.

## Phase 1: Specify conformance and engine ownership

### 1.1 Expected behavior

- [x] `CSS-01` Record height-query fixtures below, at, and above min-height and max-height boundaries, including resize in both directions.
- [x] `CSS-02` Record sticky geometry for activation, containing-block limits, release, and scroll return, with an explicit supported scroll-container scope.
- [x] `CSS-03` Define sticky behavior under clipping, overlap, zoom, hit testing, and dirty repaint; distinguish viewport scrolling from nested scrolling.
- [x] `CSS-04` Create a separately scoped Blinkless change with the fixtures, ownership boundaries, expected results, and publication dependency linked here.

Inspecting or editing another checkout requires that separately authorized
engine task. Creating this OwnFrame plan does not change another repository.
Do not emulate engine CSS by adding app-specific counter-scrolling rules.

### 1.2 Engine change and supported contract

The user separately authorized `/home/chinmay/ChinmayPersonalProjects/blinkless`.
The engine change is based on `3c75a6e2faf2bd1fa77945784368d9e8e2560c5e`.
It owns height parsing/evaluation in `internal/css`, viewport sticky constraints
in `internal/layout/sticky_placed.go`, and the public read-only
`layout.StickyView` contract in `layout/sticky.go`. OwnFrame owns passing the
native viewport after zoom, repainting the current sticky view, and routing
pointer hits in paint order. The two pre-existing engine release-record edits
are unrelated and excluded from this change.

Height fixtures cover 399/400/401 and 499/500/501 CSS px in both resize directions.
Exact, min/max, and single comparison forms accept px, pt, em, and rem. Linked
stylesheet gates reevaluate on resize without HTML or stylesheet reparsing.
Container queries stay width/inline-size only, including under negation.

The sticky fixture is `internal/window/testdata/sticky.html`, shared by the native
and Chrome drivers. The region begins at document y=80 and is 300 px tall; its
40 px header uses top:0. Scroll offsets 0, 100, 200, 360, 400, and return produce
viewport header y=80, 0, 0, -20, -60, and 0. The header stops at document y=340,
then scrolls away with its containing block. `overflow:clip` stays a clip rather
than creating a scroll container. A higher-z lime overlay wins both paint and
hits while in-flow content remains beneath the sticky header.

The verified scope is ordinary horizontal-writing viewport scrolling. Native
checks cover 1.5 zoom, clicked header coordinates, live color changes while
scrolled, and return to the original position. Nested overflow scrolling,
nested sticky elements, transformed/fixed ancestors, vertical writing, 3D
content, and percentage insets are outside the verified contract. This is an
explicit support boundary, not a claim of complete browser CSS conformance.

Sticky pages take a full viewport replay so old positions cannot survive in
page tiles. Ordinary pages keep their previous caches. Derived view cache
entries are evicted after painting; source layout and motion ownership stay
with the original display. No application counter-scrolling emulates sticky.
The source rule is [CSS Positioned Layout 3 section 3.4](https://www.w3.org/TR/css-position-3/#sticky-pos).

Local integration uses the ignored absolute-path workspace
`temp/v004-performance/integration.work`. It joins OwnFrame, examples, and
Blinkless only for validation. It is not a published dependency and will not
replace the required root module pin.

## Phase 2: Deliver engine behavior

### 2.1 Height queries

- [x] `CSS-05` Implement height media-feature evaluation in Blinkless using the actual viewport height and define the supported query forms.
- [x] `CSS-06` Verify engine query matching at boundaries and after viewport changes without reparsing solely for a resize.

### 2.2 Sticky positioning

- [x] `CSS-07` Implement the agreed sticky geometry and scroll-dependent placement in Blinkless with a clear contract for host scrolling.
- [x] `CSS-08` Verify containing-block limits and release behavior against independent expected geometry, including overlap and clipping cases.
- [x] `CSS-09` Publish a tested Blinkless revision and record its commit, module version, fixture results, and remaining unsupported cases.

Keep the OwnFrame adoption phase open until the engine contract and published
revision are available. The host may need to consume scroll-dependent metadata;
a permanently fixed header is not an implementation of sticky positioning.

## Phase 3: Adopt the engine in OwnFrame

### 3.1 Dependency and host integration

- [x] `CSS-10` Update the root `go.mod` engine pin and the required checksum files to the published revision without a local engine replace.
- [x] `CSS-11` Integrate any required engine scroll contract into `internal/window`, keeping pinned application chrome behavior intact.
- [x] `CSS-12` Replace the height-query limitation expectations in `internal/page/page_viewport_test.go` with boundary and resize assertions for the supported forms.
- [x] `CSS-13` Replace the sticky limitation expectations in `internal/page/page_sticky_hole_test.go` with assertions for the adopted engine contract.
- [x] `CSS-14` Add native framebuffer tests for sticky activation, containing-block release, scroll return, overlap, and changes while scrolled.
- [x] `CSS-15` Verify pointer hit testing targets the currently painted sticky position after scrolling and zooming.
- [x] `CSS-16` Preserve passing fixed-position, viewport windowing, and Telegram pinned-chrome integration tests.

## Phase 4: Publish only verified support

### 4.1 Closure gates

- [x] `CSS-17` Update `documentation/features.md`, `documentation/screen.md`, and applicable compatibility rows with the exact supported scope and residual limits.
- [x] `CSS-18` Retain browser-reference comparisons and native scrolling captures with fixture, viewport, scroll position, source commits, and host recorded.
- [x] `CSS-19` Run `make lint` on the final OwnFrame implementation and record its exit status.
- [x] `CSS-20` Run `make test` on the final OwnFrame implementation and record its exit status.
- [x] `CSS-21` Run `make graphics-check` and retain the required pixel-test results.
- [x] `CSS-22` Run `make docs-check plans-check examples-check` after support documentation changes and record its exit status.
- [x] `CSS-23` Record the Blinkless result separately from the OwnFrame adoption result, naming every changed module manifest and each verified host.

## Validation and adoption evidence

Blinkless commit [`27722d9cedd27f9c8959e706178cd108ada7616c`](https://github.com/chinmay-sawant/blinkless/commit/27722d9cedd27f9c8959e706178cd108ada7616c)
is pushed on `feature/v004-viewport-css-and-layout`. Both OwnFrame modules require
`v0.1.2-0.20261111104327-27722d9cedd2`. Changed manifests are root `go.mod`,
root `go.sum`, `examples/go.mod`, and `examples/go.sum`. The engine's manifests
are unchanged. No engine replace is committed or used for these final gates.
The temporary local workspace was used during development only.

| Gate | Exit | Evidence |
|---|---|---|
| **Engine build, tests, lint, golden, claim scan** | 0 each | [Engine validation and owned source hashes](evidence/engine-validation.json), [build](evidence/engine-build.txt), [tests](evidence/engine-tests.txt), [lint](evidence/engine-lint.txt), [golden](evidence/engine-golden.txt), [claim scan](evidence/engine-claim-scan.txt). |
| **Engine line-height race checks** | 0 | [Race result](evidence/engine-lineheight-race.txt). |
| **make test** | 0 | [Final library and examples suite](evidence/css-tests.txt), including fixed positioning, viewport windowing, and Telegram pinned chrome. |
| **make lint** | 0 | [Final gates](evidence/css-lint.txt), including documentation, plans, examples, and internal imports. |
| **make graphics-check** | 0 | [16 required scenarios](evidence/css-graphics.txt), including sticky pixels and hits. |
| **make docs-check plans-check examples-check** | 0 | [Final documentation and maintenance checks](evidence/css-docs.txt). |
| **Chrome reference** | 0 | [Boundary and hit observations](evidence/browser-reference.json). Command: `node scripts/check-sticky-reference.mjs`. |
| **Native/Chrome RGBA comparison** | 0 | [Eight equal framebuffers](evidence/sticky-parity.json). Command: `python3 scripts/compare-sticky-captures.py --native temp/v004-performance/sticky-captures --reference temp/v004-performance/browser-reference --out temp/v004-performance/sticky-parity.json`. |
| **Final Linux live AT-SPI recheck** | 0 | [Focus, activation, disabled-target refusal, password redaction](evidence/atspi-live-final.json). |

[Host provenance](evidence/css-validation.json) records final source hashes,
module adoption, commands, and input scope. [Workspace resolution](evidence/engine-module.json),
[standalone library resolution](evidence/engine-module-standalone.json), and
[standalone examples resolution](evidence/engine-examples-standalone.json) all
select the published module with no engine replacement.

Native captures are [initial](evidence/sticky-0.png), [activation](evidence/sticky-1.png),
[overlap](evidence/sticky-2.png), [release](evidence/sticky-3.png),
[fully released](evidence/sticky-4.png), [return](evidence/sticky-5.png),
[edit while scrolled](evidence/sticky-6.png), [zoom](evidence/sticky-7.png),
[zoomed hit](evidence/sticky-8.png), and [return to start](evidence/sticky-9.png).
The matching Chrome files use `chrome-sticky-<step>.png`. Native proof is Linux
Xvfb with software OpenGL and input submitted through the page APIs on the
native loop. It does not establish physical-device input or other native hosts.

## Dependencies

- Blinkless implementation, validation, and a published engine revision precede
  the OwnFrame pin change. An unadopted patch is not downstream support.
- Track 01 owns the current flex-absolute documentation correction.
- Xvfb checks establish the tested Linux software graphics scenarios. Other
  native hosts need separate evidence before platform-wide support claims.

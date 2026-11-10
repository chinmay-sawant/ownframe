# v0.0.4 / 04 - CSS conformance and engine adoption

> **Parent:** [v0.0.4 index](README.md)
> **Status:** planned. Engine implementation and publication are external dependencies.
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

- [ ] `CSS-01` Record height-query fixtures below, at, and above min-height and max-height boundaries, including resize in both directions.
- [ ] `CSS-02` Record sticky geometry for activation, containing-block limits, release, and scroll return, with an explicit supported scroll-container scope.
- [ ] `CSS-03` Define sticky behavior under clipping, overlap, zoom, hit testing, and dirty repaint; distinguish viewport scrolling from nested scrolling.
- [ ] `CSS-04` Create a separately scoped Blinkless change with the fixtures, ownership boundaries, expected results, and publication dependency linked here.

Inspecting or editing another checkout requires that separately authorized
engine task. Creating this OwnFrame plan does not change another repository.
Do not emulate engine CSS by adding app-specific counter-scrolling rules.

## Phase 2: Deliver engine behavior

### 2.1 Height queries

- [ ] `CSS-05` Implement height media-feature evaluation in Blinkless using the actual viewport height and define the supported query forms.
- [ ] `CSS-06` Verify engine query matching at boundaries and after viewport changes without reparsing solely for a resize.

### 2.2 Sticky positioning

- [ ] `CSS-07` Implement the agreed sticky geometry and scroll-dependent placement in Blinkless with a clear contract for host scrolling.
- [ ] `CSS-08` Verify containing-block limits and release behavior against independent expected geometry, including overlap and clipping cases.
- [ ] `CSS-09` Publish a tested Blinkless revision and record its commit, module version, fixture results, and remaining unsupported cases.

Keep the OwnFrame adoption phase open until the engine contract and published
revision are available. The host may need to consume scroll-dependent metadata;
a permanently fixed header is not an implementation of sticky positioning.

## Phase 3: Adopt the engine in OwnFrame

### 3.1 Dependency and host integration

- [ ] `CSS-10` Update the root `go.mod` engine pin and the required checksum files to the published revision without a local engine replace.
- [ ] `CSS-11` Integrate any required engine scroll contract into `internal/window`, keeping pinned application chrome behavior intact.
- [ ] `CSS-12` Replace the height-query limitation expectations in `internal/page/page_viewport_test.go` with boundary and resize assertions for the supported forms.
- [ ] `CSS-13` Replace the sticky limitation expectations in `internal/page/page_sticky_hole_test.go` with assertions for the adopted engine contract.
- [ ] `CSS-14` Add native framebuffer tests for sticky activation, containing-block release, scroll return, overlap, and changes while scrolled.
- [ ] `CSS-15` Verify pointer hit testing targets the currently painted sticky position after scrolling and zooming.
- [ ] `CSS-16` Preserve passing fixed-position, viewport windowing, and Telegram pinned-chrome integration tests.

## Phase 4: Publish only verified support

### 4.1 Closure gates

- [ ] `CSS-17` Update `documentation/features.md`, `documentation/screen.md`, and applicable compatibility rows with the exact supported scope and residual limits.
- [ ] `CSS-18` Retain browser-reference comparisons and native scrolling captures with fixture, viewport, scroll position, source commits, and host recorded.
- [ ] `CSS-19` Run `make lint` on the final OwnFrame implementation and record its exit status.
- [ ] `CSS-20` Run `make test` on the final OwnFrame implementation and record its exit status.
- [ ] `CSS-21` Run `make graphics-check` and retain the required pixel-test results.
- [ ] `CSS-22` Run `make docs-check plans-check examples-check` after support documentation changes and record its exit status.
- [ ] `CSS-23` Record the Blinkless result separately from the OwnFrame adoption result, naming every changed module manifest and each verified host.

## Dependencies

- Blinkless implementation, validation, and a published engine revision precede
  the OwnFrame pin change. An unmerged patch is not downstream support.
- Track 01 owns the current flex-absolute documentation correction.
- Xvfb checks establish the tested Linux software graphics scenarios. Other
  native hosts need separate evidence before platform-wide support claims.

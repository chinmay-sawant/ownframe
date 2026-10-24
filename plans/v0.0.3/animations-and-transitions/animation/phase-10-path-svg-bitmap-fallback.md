# Phase 10 - path, SVG paint, and bitmap fallback

Goal: close the last implemented-property gaps that keep a page off the live
display list. A page that animates a corner shape, an inline SVG, or a fallback
picture either animates correctly or reports why it cannot.

This phase is the catch-all for the items in
[coverage.md](../coverage.md) section 3. It is last because it depends on the
selected display path, not on the timeline.

Depends on [phase 11](phase-11-engine-named-timelines-and-svg-paint.md) for SVG paint. The path
replay and bitmap fallback parts are OwnFrame-only.

## Where it lives

- `blinkless@v0.0.3` emits `DisplayOpPath` for corner shapes
  (`internal/layout/corner_shape_paint.go:21`, `displaylist.go:151-154`).
- `internal/render/display_ops.go:8-50` restates every op kind and has no
  `OpPath`. The replay painter `internal/replay/op.go:14-39` has no case for it.
- `internal/render/replayable.go` rejects an unknown kind by default, so one
  `OpPath` sends the whole page to the bitmap path.
- `internal/page/page_draw.go:206-224` `paintFallback` commits a bitmap and keeps
  the display list. `internal/window/sync.go:29-40` stores the picture; the
  window draws the picture, not the list.
- `documentation/frames.md` records that a bitmap-fallback page has no list to
  patch.
- Inline SVG bakes to `OpImage` with a serialized SVG payload
  (`css_behavior_svg_paint_test.go`, `TestBehaviorFillBakesPaint`). There is no
  path op to patch.

## The three gaps

### A. `DisplayOpPath` replay

Corner shapes are implemented by the engine and emitted as paths. OwnFrame cannot
paint them, so any page with a `corner-*` declaration falls back to a bitmap.
That is a rendering gap, but it blocks animation too: a bitmap page ignores op
patches.

Work:
- add `OpPath` to `internal/render/display_ops.go`,
- add a painter in `internal/replay` that reads the corner path command list,
  fills it with the op colour, and honours the op matrix and opacity,
- add it to `replayableOp`, to `opBounds`, and to the chunk signature,
- add a bitmap-path test that fails today, then passes,
- verify that a corner-shape page stays on `PathList` after the change.

This is the highest-value item in the phase because it removes a whole class of
fallbacks.

### B. SVG paint

`fill`, `stroke`, `stroke-width`, `stroke-dasharray`, and `stroke-dashoffset`
change a serialized SVG that the engine turned into an image. There is no field
to patch.

[Phase 11](phase-11-engine-named-timelines-and-svg-paint.md) provides `RestyleInlineSVG`: it
re-serializes one inline SVG with a paint override, rasterizes it, and returns an
`ElementPatch`. This phase integrates that API under the phase 04 budget.

If phase 11 ships root-only, this phase covers inherited root paint. Per-part
animation waits for part metadata. If the API has not landed, count the property
as deferred. Do not rewrite the SVG string in OwnFrame; the serializer lives in
the engine.

### C. Bitmap fallback

Some pages will still fall back: a layout the replay gate refuses, a filter the
gate does not accept, or a kind added later. The display list is retained, so the
driver can patch it, but the window draws the picture.

Options:
1. Re-render the bitmap from the patched display each frame through the existing
   `draw.PaintBitmap` path. Correct and simple, but a full-page paint per frame.
2. Leave the page static and report the reason in `Probe().Motion()`.
3. A hybrid: re-render only while a fallback page has active motion, and stop
   when it finishes.

Recommendation: start with 2, measure 1 on one small example, and adopt 3 only if
the measured cost is acceptable. A silent static page is not acceptable; the
counter and the matrix note are mandatory.

## Checklist

### A. Path replay

- [ ] Add `OpPath` to the render kind table.
- [ ] Add the painter, including fill, stroke, matrix, and opacity.
- [ ] Add `OpPath` to `replayableOp`.
- [ ] Add path bounds to `opBounds` and the chunk signature.
- [ ] Add a corner-shape page that stays on the list path.
- [ ] Add a golden for a corner-shape box.

### B. SVG paint

- [ ] Consume `RestyleInlineSVG` for fill and stroke paint.
- [ ] Adopt its `ElementPatch` through the existing splice path.
- [ ] Route it through the phase 04 budget with a per-element cap.
- [ ] Root-only mode: document the per-part gap.
- [ ] No API: count the property as deferred and add the matrix note.
- [ ] Add a card for `fill` and `stroke-width` with the chosen behaviour.

### C. Bitmap fallback

- [ ] Detect a fallback page in the runtime and stop patching ops that cannot be
      seen.
- [ ] Report `MotionBlockedFallback` in `Probe()`.
- [ ] Measure a re-render from the retained list on one small page.
- [ ] Adopt re-render only with a measured budget; otherwise keep the report.

### D. Docs and tests

- [ ] Update `frames.md` and `repaint.md` with the path and fallback rules.
- [ ] Add the blocked rows to `compatibility-matrix.md` while they exist.
- [ ] Add the phase 09 conformance cards for the three gaps.
- [ ] Run `make graphics-check`.

Gate: a corner-shape page animates on the list path, an SVG paint card either
animates or reports deferred, and a fallback page reports why it is static.

## Risks

- A path painter is a new raster path with its own parity risk. Keep it behind
  the same golden gate as the other op kinds.
- Re-rendering a bitmap per frame can look like a working feature until a large
  page hits it. Cap it from the start.
- The engine re-serialize hook in phase 11 is the only clean SVG answer. Do not
  ship a string rewrite in OwnFrame.
- SVG rasterization per frame is the cost limit. Cap it and report a snap rather
  than dropping frames.

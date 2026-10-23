# Phase 2: Underlying Layout Engine & Blinkless Gaps

> **Parent:** [README.md](README.md)  
> **Area:** HTML/CSS Layout Engine (`blinkless`), Display Replay & CSS Fidelity  
> **Status:** Open `[ ]`  

---

## Overview

OwnFrame uses [blinkless](https://github.com/chinmay-sawant/blinkless) for CSS cascading, box generation, and element placement. While core Flexbox, Grid, tables, and box models are functional, specific layout engine limitations force applications into manual workarounds or trigger full bitmap fallbacks:

1. **`position: absolute` on Flex Children:** Blinkless keeps out-of-flow children in the flex flow, ignoring `top`, `bottom`, `left`, and `right`.
2. **`position: sticky` Across Scrolled Contexts:** Elements styled with `position: sticky` are not translated dynamically during scroll replay; apps must manually coordinate `Page.SetScrollWindow` and `SetViewportPinZ`.
3. **Missing Media Queries:** `@media (height)` and `@media (prefers-color-scheme)` never match.
4. **Bitmap Fallback for Non-Separable Blends:** Modes like `hue`, `saturation`, `color`, and `luminosity` bypass the GPU display list and force full CPU rasterization.
5. **Container Queries vs. Element Patching:** Documents declaring `@container` queries trigger immediate rejection by `Page.PatchElement`.

Closing these gaps eliminates awkward application workarounds and expands CSS compatibility.

---

## Strategy & Principles

1. **Upstream First:** Land layout and cascading fixes in `github.com/chinmay-sawant/blinkless` with regression tests, then bump the module pin in `go.mod`.
2. **No Faux Implementations:** Do not paper over missing layout calculations with fragile heuristic transforms inside OwnFrame.
3. **GPU First:** Migrate remaining CPU bitmap fallback paths into hardware-accelerated Kage shaders in `internal/replay`.

---

## Checklist

### 2.1 Out-of-Flow Positioning in Flex Containers

- [ ] `LAYOUT-01` Upstream in `blinkless`: Remove `position: absolute` children from flex line sizing and placement calculations. Calculate their containing block relative to the flex container’s padding box (or nearest positioned ancestor) and apply `top/right/bottom/left` insets.  
  *Acceptance Criteria:* New blinkless layout test with flex container and absolute child passes; child position matches CSS Flexbox §4.1.

- [ ] `LAYOUT-02` In OwnFrame: Update `blinkless` dependency pin in `go.mod` and verify `examples/telegram` chat thread layout with native CSS `position: absolute` on thread controls without auxiliary wrapper blocks.  
  *Acceptance Criteria:* `make build`, `make test`, and `TestFlexChildAbsolutePosition` pass without workarounds.

### 2.2 First-Class `position: sticky`

- [ ] `LAYOUT-03` Upstream in `blinkless`: Mark sticky boxes with `OpSticky` in `layout.DisplayList`, carrying their scroll parent id, normal flow bounds, and sticky containment bounds.  
  *Acceptance Criteria:* Blinkless emits display ops with sticky constraint metadata.

- [ ] `LAYOUT-04` In `internal/replay`: Calculate sticky offsets dynamically during window scroll replay based on the scroll viewport offset, clamping translation within the container bounds.  
  *Acceptance Criteria:* A sticky header remains pinned to the viewport top during mouse wheel and touch scroll without calling `Page.SetScrollWindow` or `SetViewportPinZ`.

### 2.3 Extended Media Queries & Dark Mode

- [ ] `LAYOUT-05` Implement `@media (height)`, `(min-height)`, and `(max-height)` evaluation in `blinkless` based on the frame viewport height passed via `css.Options`.  
  *Acceptance Criteria:* Unskip `TestMediaHeightFollowsTheFrame`; test passes dynamically on window resize.

- [ ] `LAYOUT-06` Implement `@media (prefers-color-scheme: dark | light)` matching in `blinkless`. Pass host OS theme state from `internal/window` through `css.Options.DarkMode`.  
  *Acceptance Criteria:* Switching OS theme switches stylesheet rules matching `@media (prefers-color-scheme: dark)` without requiring an explicit `Page.SetTheme` override.

### 2.4 GPU Shaders for Advanced Blend Modes

- [ ] `LAYOUT-07` In `internal/replay`: Implement non-separable blend modes (`hue`, `saturation`, `color`, `luminosity`) as custom Kage fragment shaders (`group_blend.go`). Remove the full CPU bitmap fallback path for these CSS mix-blend-modes.  
  *Acceptance Criteria:* Replay keeps display-list execution for all standard blend modes; `bitmap fallback` badge is not displayed.

### 2.5 Element Patching with Container Queries

- [ ] `LAYOUT-08` In `blinkless` and `internal/render`: Track container query dependency scopes. If an element patch does not alter the geometry of a container, allow `PatchElement` to proceed rather than returning `ErrPatchRefused`.  
  *Acceptance Criteria:* `PatchElement` successfully updates content inside container-queried documents when container dimensions remain constant.

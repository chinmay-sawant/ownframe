# Phase 04: scrolling and chrome

Goal: scroll a long file while the window frame stays still, without turning
the app into a manual canvas.

## Where it lives

- `app.go` owns `windowSlice`, `contentHeight`, `clampScroll`, `scrollTo`, and
  `onScroll`.
- `components/editor.css` owns the fixed editor and the 20 px code row.
- `components/diff.css` owns the 20 px diff row.
- `components/motion.css` owns the chevron, drawer, and toast transitions.

## Design

1. The page scrolls. `.void` and `.scroll-space` are the only normal-flow
   content, and their height is the page height.
2. The chrome is `position: fixed` with explicit heights. The window keeps the
   offset and paints the fixed ops untranslated.
3. `SetScrollWindow` receives the page offset and the viewport height. It slices
   the code or diff rows and calls `SetData` only when the slice changes.
4. `SetWindowing(true)` makes a scroll offset relayout instead of a full frame
   repaint.
5. Every code and diff row is 20 px. The track sits at `start*20 - scrollY`.
   The slice is trimmed so no row extends past the code body.
6. Eight overscan rows above the band hide the seam while a wheel notch moves
   the page 48 px. There is no bottom overscan, so a row cannot paint over the
   status bar.
7. The details drawer changes the editor width, which changes how many rows fit.
   The next callback rebuilds the slice for the new width.
8. `Page.ScrollTo` queues the target. In a headless test,
   `Page.Scroll().SetScrollOffset` supplies the same offset before the redraw.

## Checklist

- [x] Fixed shell with explicit pane heights.
- [x] Page height from the code or diff row count.
- [x] Row slice with eight-row top overscan.
- [x] Track offset and no bottom overhang.
- [x] Rebuild the slice on resize.
- [x] Reset to the top when a file, mode, or diff selection changes.
- [x] Arrow, page, home, and end keys.
- [x] Drawer width change rebuilds the slice.
- [ ] Horizontal scroll or wrap for long lines.
- [ ] Mouse-wheel row step that matches the 48 px notch.
- [ ] Save and restore the offset per tab.

## Gate

`TestScrollWindow` confirms the track is negative, the first row advances, the
callback reports a change, and the reset returns to row 1.
`TestShellMetrics` confirms the code body height at 1500x900.

## Risks

- The 20 px row exists in Go (`rowH`) and CSS (`--row-h`). Change both or the
  last row lands on the status bar.
- A fixed box needs an explicit height. A future layout that uses only `top`
  and `bottom` collapses the code body.
- The scroll callback runs before the template executes. It must set the data
  for the new slice, not call `Redraw`.

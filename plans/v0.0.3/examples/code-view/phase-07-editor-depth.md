# Phase 07: editor depth

Goal: read two files at once and read a long line without losing the row math.

Status: planned.

## Design

1. Split the editor into two groups. Each group has its own tab list, active
   path, and scroll offset. The page scroll belongs to the focused group.
2. Add a horizontal offset to the code track. A wide track defines the page
   width, and the body clips to the viewport.
3. Add a wrap toggle. Wrapped source lines become fixed 20 px visual rows, so
   the existing slice and spacer math stays exact.
4. Between hunks, show a collapsed context row with the skipped line count.
   Clicking it expands the context.
5. Add go-to-line and a line-range copy.
6. Add breadcrumb dropdowns for the directories and the outline.
7. For files over 500 lines, add a minimap or a scroll overview. The overview
   is a fixed strip of marks, not a second render of the code.

## Checklist

- [ ] Add the second editor group and its tab list.
- [ ] Give each group an active path and a scroll offset.
- [ ] Route the scroll callback to the focused group.
- [ ] Add a horizontal track offset.
- [x] Add a wrap toggle with fixed visual rows. Evidence:
  `TestWrapLongCodeLine`, `wrap.go`.
- [ ] Add expandable diff context.
- [ ] Add go-to-line and range copy.
- [ ] Add breadcrumb dropdowns.
- [ ] Add the minimap or overview.
- [ ] Add tests for the split groups, wrap, and context expansion.

## Gate

Two files are readable side by side, and a wrapped line keeps the row math
exact.

## Risks

- Wrapped rows must stay one visual row per 20 px slice. A variable row height
  would break the spacer math; the plan keeps every visual row fixed.
- Two editor groups double the state. Keep one owner per group and pass the
  focused group to the scroll callback.
- A minimap is a paint cost. It needs a separate budget and an off switch.

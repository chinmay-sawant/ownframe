# Phase 07: editor depth

Goal: read two files at once and read a long line without losing the row math.

Status: planned.

## Design

1. Split the editor into two groups. Each group has its own tab list, active
   path, and scroll offset. The page scroll belongs to the focused group.
2. Add a horizontal offset to the code track. A wide track defines the page
   width, and the body clips to the viewport.
3. Add a wrap toggle. A wrapped row changes height, so switch to a variable
   row layout and stop using the fixed 20 px slice while wrap is on.
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
- [ ] Add a wrap toggle with a variable-row layout.
- [ ] Add expandable diff context.
- [ ] Add go-to-line and range copy.
- [ ] Add breadcrumb dropdowns.
- [ ] Add the minimap or overview.
- [ ] Add tests for the split groups, wrap, and context expansion.

## Gate

Two files are readable side by side, and a wrapped line keeps the row math
exact.

## Risks

- A variable row height breaks the fixed slice. Wrap needs its own windowing
  path, or the row math must measure every line.
- Two editor groups double the state. Keep one owner per group and pass the
  focused group to the scroll callback.
- A minimap is a paint cost. It needs a separate budget and an off switch.

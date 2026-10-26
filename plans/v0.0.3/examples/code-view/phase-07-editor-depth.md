# Phase 07: editor depth

Goal: read two files at once and read a long line without losing the row math.

Status: partly landed. Two editor groups, the wrap toggle, diff wrapping, and
inline Markdown preview are in. Each group has its own tab list and active
path, but the two panes share one page scroll. Horizontal scroll, diff context
expansion, go-to-line, breadcrumb dropdowns, and a minimap are open.

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

- [x] Add the second editor group and its tab list. Evidence:
      `TestFeatureSplitEditor`, `editor_group.go`, `pane.go`.
- [x] Give each group an active path. Evidence: `TestFeatureSplitEditor`.
- [ ] Give each group an independent scroll offset. The panes share the page
      offset today.
- [ ] Route the scroll callback to the focused group. It still drives both.
- [ ] Add a horizontal track offset.
- [x] Add a wrap toggle with fixed visual rows. Evidence:
  `TestWrapLongCodeLine`, `wrap.go`.
- [x] Wrap unified and split diff cells. Evidence: `TestDiffWrapUnified`,
  `TestFeatureDiffWrapIntegration`, `diff_wrap.go`.
- [x] Add edit, selection, and save in the source view. Evidence:
  `TestFeatureEditAndSave`, `edit_buffer.go`, `workspace_write.go`.
- [x] Start editing by clicking inside the file, with gutter markers on
  changed lines and a line card. Evidence:
  `TestEditSurfaceClickBeginsEditing`, `TestFeatureChangedLineMarkers`.
- [x] Add a bottom-center changed-file navigator. Evidence:
  `TestFeatureFileNav`, `modified_nav.go`.
- [ ] Add expandable diff context.
- [ ] Add go-to-line and range copy.
- [ ] Add breadcrumb dropdowns.
- [ ] Add the minimap or overview.
- [x] Add tests for the split groups, wrap, diff wrap, and editing.

## Gate

Two files are readable side by side, and a wrapped line keeps the row math
exact.

## Risks

- Wrapped rows must stay one visual row per 20 px slice. A variable row height
  would break the spacer math; the plan keeps every visual row fixed.
- Two editor groups double the state. Keep one owner per group and pass the
  focused group to the scroll callback.
- A minimap is a paint cost. It needs a separate budget and an off switch.

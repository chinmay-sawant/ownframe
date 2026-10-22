# Scoped style invalidation

This page is historical. `internal/styleinvalid` is not in the tree. Live
layout and style work lives in blinkless and `internal/render`. The lists
below are the old plan, not a current package.

`internal/styleinvalid` tracked which boxes and style paths actually moved
during a state change so unmodified subtrees could be skipped during style
and layout.

## Supported

- Per-box CSS dependency tracking via `Tracker`.
- Dirty ranges via `RangesFromChanges`.
- Skipping unmodified matched selectors via `SkipUnmodifiedSelectors`.
- Style path invalidation via `StylePathsFromChanges` and `InvalidateStylePaths`.
- Fixed/sticky/absolute layer promotion via `PromoteLayer`.
- Anchor-based offsets via `Anchor`.
- Stable text runs via `TextRunKey` and `RunCache`.

## Unsupported under scoped invalidation

- Global percentage reflow that can affect the whole document.
- More than two style dirty passes per frame before paint.
- Full-relayout style features that cannot be localized to a changed subtree.

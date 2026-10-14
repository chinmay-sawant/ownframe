# Scoped style invalidation

`internal/styleinvalid` tracks which boxes and style paths actually move during
a state change so unmodified subtrees can be skipped during style and layout.

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

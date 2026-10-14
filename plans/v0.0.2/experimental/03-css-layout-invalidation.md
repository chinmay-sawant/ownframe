# Phase 3: Scoped CSS and Layout Invalidation

Checklist for style and layout damage that does not affect the entire page.

- [x] Track per-box CSS dependencies rather than rebuilding every rule.
- [x] Skip unmodified matched selectors when state changes.
- [x] Invalidate only style paths touched by state changes.
- [x] Re-measure subtrees instead of the full document.
- [x] Add fixed and pinned layer promotion rules for opt-in compositing.
- [x] Support anchor-based positioned elements without full relayout.
- [x] Avoid global percentages where they can affect only the changed subtree.
- [x] Bound style dirty twice per frame at most before paint.
- [x] Record style dirty ranges in page stats.
- [x] Make repeated text runs stable enough to reuse baked caches.
- [x] Document which CSS features remain unsupported under scoped invalidation.

## Acceptance criteria

- A small style change leaves unrelated sections unstyled and unmeasured.
- Layout cache hit rate excludes only visibly changed elements.
- Scroll still uses document-level offsets instead of re-flowing fixed content.
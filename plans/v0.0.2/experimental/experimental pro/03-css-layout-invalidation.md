# Phase 3: Scoped CSS and Layout Invalidation

Checklist for style and layout damage that does not affect the entire page.

- [ ] Track per-box CSS dependencies rather than rebuilding every rule.
- [ ] Skip unmodified matched selectors when state changes.
- [ ] Invalidate only style paths touched by state changes.
- [ ] Re-measure subtrees instead of the full document.
- [ ] Add fixed and pinned layer promotion rules for opt-in compositing.
- [ ] Support anchor-based positioned elements without full relayout.
- [ ] Avoid global percentages where they can affect only the changed subtree.
- [ ] Bound style dirty twice per frame at most before paint.
- [ ] Record style dirty ranges in page stats.
- [ ] Make repeated text runs stable enough to reuse baked caches.
- [ ] Document which CSS features remain unsupported under scoped invalidation.

## Acceptance criteria

- A small style change leaves unrelated sections unstyled and unmeasured.
- Layout cache hit rate excludes only visibly changed elements.
- Scroll still uses document-level offsets instead of re-flowing fixed content.
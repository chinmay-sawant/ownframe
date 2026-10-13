# Phase 4: Display Patching and Retained Rendering

Checklist to replace whole-display rebuild with operation-level patching.

- [ ] Build operation diffs instead of replacing the entire display.
- [ ] Patch only operations that dominate existing bounds.
- [ ] Keep unchanged operations untouched between frames.
- [ ] Move freed operation memory into a recyclable pool.
- [ ] Rasterize reused element chunks to stay cheap.
- [ ] Extend raster signature fields for animation and text operations.
- [ ] Precompute transformed op bounds for moved elements.
- [ ] Merge tile damage at operation boundaries, not just pixel boxes.
- [ ] Preserve cache invariants on partial buffer reuse.
- [ ] Avoid re-blitting unchanged tiles during identity frames.
- [ ] Add a debug path showing which ops were patched vs added.
- [ ] Keep fallback bitmap rendering working as a last resort.

## Acceptance criteria

- A one-key edit checks fewer than all operations and ends in under a millisecond.
- No visible ghost appears after removal.
- Paint time tracks touched ops, not page size.
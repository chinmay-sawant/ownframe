# Phase 4: Display Patching and Retained Rendering

Checklist to replace whole-display rebuild with operation-level patching.

- [x] Build operation diffs instead of replacing the entire display.
- [x] Patch only operations that dominate existing bounds.
- [x] Keep unchanged operations untouched between frames.
- [x] Move freed operation memory into a recyclable pool.
- [x] Rasterize reused element chunks to stay cheap.
- [x] Extend raster signature fields for animation and text operations.
- [x] Precompute transformed op bounds for moved elements.
- [x] Merge tile damage at operation boundaries, not just pixel boxes.
- [x] Preserve cache invariants on partial buffer reuse.
- [x] Avoid re-blitting unchanged tiles during identity frames.
- [x] Add a debug path showing which ops were patched vs added.
- [x] Keep fallback bitmap rendering working as a last resort.

## Acceptance criteria

- A one-key edit checks fewer than all operations and ends in under a millisecond.
- No visible ghost appears after removal.
- Paint time tracks touched ops, not page size.
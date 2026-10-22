# Migrating existing template apps

Existing template apps keep calling `SetData` and work unchanged. The
retained/reactive layer is opt-in: adopt it per page by opting into
field-level patches and retained element IDs. Full repaint remains the
fallback when partial data is unavailable.

## Steps

1. Keep the current template and `SetData` calls working. No migration is
   required for behavior parity.
2. Add stable IDs to repeated elements so list diffs stay deterministic.
3. Prefer field-level binding events over whole-view `SetData` rebuilds
   when only a scalar changes.
4. Let handlers patch projected state slices; full document redraws remain
   the default on handler failure.
5. Keep focus and scroll state attached to stable element identities.
6. Preserve Undo/Redo references on stable element identity, not box hits.
7. Surface changed-property metadata from handlers to the page layer.
8. Use `page.Probe().Invalidate` when a fetch, timer, or clock mutates pixels the
   diff does not know about.
9. Use loading markers and suspense placeholders instead of blank repaints
   while deferred data loads.
10. Keep accessibility fallback tied to retained element state; removed
    elements drop their semantic nodes.

## Rollback

Disable partial updates and return to full `Redraw` at any time; the page
still reports the same PNG and boxes.

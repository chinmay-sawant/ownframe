# Phase 06: live workspace

Goal: keep the tree, the open file, and the diff current without a restart.

Status: planned.

## Design

1. A file watcher reports added, changed, and removed paths. The app debounces
   the batch and rescans once.
2. Git status polls on a slow clock, not on every frame. The dirty file list and
   the branch update only when they change.
3. A rescan keeps folder expansion, the active tab, and the diff selection when
   those paths still exist.
4. The explorer becomes a windowed list with its own offset. The page wheel
   handler scrolls it under the pointer; the draggable rail, footer arrows, and
   keyboard remain the alternatives.
5. Recent files persist under the user config directory, with a bounded list.
6. A scan budget stops at one clear file count and says so. A repository larger
   than the budget is still usable through the filter.
7. The active file reloads when it changes on disk. A conflict opens a notice,
   not a silent overwrite.

## Checklist

- [x] Add a debounced watcher behind `-root`. Evidence: `watch.go`,
      `TestWatcherUpdatesWorkingTree`.
- [x] Scroll the explorer with the wheel. Evidence: `onWheel`,
      `TestTreeWheelAfterPageScroll`.
- [ ] Rescan and preserve expansion, tab, and selection.
- [ ] Poll `git status` and update the changed count.
- [x] Window the explorer tree. Evidence: `TestTreeExpandedFolderScrolls`.
- [ ] Persist recent files.
- [ ] Enforce the scan budget with a visible message.
- [ ] Reload the active file on disk change.
- [ ] Add tests for add, change, remove, and a preserved selection.

## Gate

A new file on disk appears in the tree without a restart, and the expanded
folders stay open.

## Risks

- A watcher on a large repository can flood the loop. The debounce and the
  batch are the guard.
- Reloading an open file can move the scroll. Keep the offset when the row count
  grows or shrinks.
- A second scroll offset in the sidebar needs a clear ownership rule: the page
  offset drives the editor; the sidebar offset is app state.

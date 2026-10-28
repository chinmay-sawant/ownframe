# Phase 06: live workspace

Goal: keep the tree, the open file, and the diff current without a restart.

Status: mostly landed. The watcher rescans files and git metadata, the tree
keeps expansion, the scan budget is visible, recent files and layout persist,
and open documents reload unless they have unsaved edits. The Source Control
lists scroll with draggable rails, and the History rail loads larger pages on a
background worker as it moves. Git metadata is watched directly, so there is
no slow status poll.

## Design

1. A file watcher reports added, changed, and removed paths. The app debounces
   the batch and rescans once.
2. Git metadata changes are watched directly through `.git/HEAD`, refs, and the
   index; there is no slow status poll.
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
8. The changed-file list and the History rail each own a vertical offset and a
   drag rail. The History rail draws a lane graph and asks for a larger page
   off the page loop when the scroll passes half of the rows it holds.

## Checklist

- [x] Add a debounced watcher behind `-root`. Evidence: `watch.go`,
      `TestWatcherUpdatesWorkingTree`.
- [x] Scroll the explorer with the wheel. Evidence: `onWheel`,
      `TestTreeWheelAfterPageScroll`.
- [x] Rescan and preserve expansion, tab, and selection. Evidence:
      `TestWatcherRescanAddsAndRemovesFiles`, `rescan.go`.
- [x] Follow `.git/HEAD`, refs, and the index so an external commit or branch
      switch updates history. Evidence:
      `TestWatcherSeesExternalBranchChange`, `isGitMetadataChange`.
- [x] Window the explorer tree. Evidence: `TestTreeExpandedFolderScrolls`.
- [x] Persist recent files, tabs, and layout. Evidence: `session.go`,
      `TestFeatureSessionRestoresLayout`.
- [x] Enforce the scan budget with a visible message and `-files` control.
      Evidence: `TestFeatureScanBudget`, `TestLoadDiskWorkspaceLimitTruncates`.
- [x] Keep a bottom-center changed-file navigator in sync with the diff.
      Evidence: `TestFeatureFileNav`, `file_nav_ops.go`.
- [x] Scroll the changed-file list with a draggable rail and the wheel instead
      of a pager. Evidence: `TestChangesScrollE2E`, `TestScrollRailsPaint`.
- [x] Scroll the History rail with a draggable rail and the wheel instead of a
      pager. Evidence: `TestHistoryScrollE2E`, `TestHistoryScrollWheel`.
- [x] Draw commit graph lanes and preload a larger history page at half scroll.
      Evidence: `TestHistoryScrollPreloadAtHalf`, `TestHistoryAsync`.
- [x] Show git-ignored rows as solid muted rows with an `IGNORED` badge, and
      keep skipped ignored directories visible. Evidence:
      `TestGitignoreRows`, `TestWorkspaceMarksGitIgnoredPaths`.
- [x] Float the commit message beside the history rail on hover. Evidence:
      `TestFeatureCommitTipFloatsBesideHistory`, `commit_tip.go`.
- [x] Reload the active file on disk change, but keep unsaved edits.
      Evidence: `applyWatchChanges`, `installRescan`.
- [x] Add tests for add, change, remove, and a preserved selection. Evidence:
      `TestWatcherRescanAddsAndRemovesFiles`,
      `TestApplyWatchChangesPreservesSelection`.

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

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

const (
	watchDebounce   = 150 * time.Millisecond
	maxWatchDirs    = 2000
	rescanNoteParts = 3
)

// watchUpdate is one debounced background result. Workspace is non-nil after a
// rescan; Git marks a change that may have moved a ref or the index.
type watchUpdate struct {
	Changes   changeSet
	Workspace *workspace
	Delta     workspaceDelta
	Git       bool
	Branch    string
	Branches  []branchRow
}

// StartWatcher watches the live workspace and wakes the UI when files or git
// metadata change. It is desktop only; Serve does not start it.
func (a *App) StartWatcher(ctx context.Context) error {
	if a.ws.Sample || a.watch != nil {
		return nil
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	if err := a.addWatchTree(watcher, a.ws.Root); err != nil {
		_ = watcher.Close()

		return err
	}

	a.addGitWatches(watcher, a.ws.Root)

	watchCtx, cancel := context.WithCancel(ctx)
	a.watch = watcher
	a.watchCancel = cancel
	a.watchUpdates = make(chan watchUpdate, 1)

	root := a.ws.Root
	limit := a.ws.Limit
	start := a.ws
	go a.watchLoop(watchCtx, watcher, root, limit, start)

	return nil
}

// StopWatcher closes the fsnotify watcher.
func (a *App) StopWatcher() {
	if a.watch == nil {
		return
	}

	if a.watchCancel != nil {
		a.watchCancel()
		a.watchCancel = nil
	}

	_ = a.watch.Close()
	a.watch = nil

	for {
		select {
		case <-a.watchUpdates:
		default:
			return
		}
	}
}

// addWatchTree adds every directory under root except skipped and .git ones.
func (a *App) addWatchTree(watcher *fsnotify.Watcher, root string) error {
	count := 0

	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return nil //nolint:nilerr // an unreadable subdirectory must not stop the walk
		}

		if !entry.IsDir() {
			return nil
		}

		if path != root && skipDir(entry.Name()) {
			return filepath.SkipDir
		}

		if strings.Contains(filepath.ToSlash(path), "/.git") {
			return filepath.SkipDir
		}

		if count >= maxWatchDirs {
			return filepath.SkipAll
		}

		if watchErr := watcher.Add(path); watchErr == nil {
			count++
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("watch %s: %w", root, err)
	}

	return nil
}

// addGitWatches watches the metadata files that move HEAD, refs, or the index.
// .git/objects is deliberately left out: object churn is not a branch change.
func (a *App) addGitWatches(watcher *fsnotify.Watcher, root string) {
	gitDir := filepath.Join(root, ".git")

	for _, path := range []string{
		gitDir,
		filepath.Join(gitDir, "refs"),
		filepath.Join(gitDir, "refs", "heads"),
		filepath.Join(gitDir, "refs", "tags"),
		filepath.Join(gitDir, "refs", "remotes"),
		filepath.Join(gitDir, "logs"),
	} {
		info, err := os.Stat(path)
		if err == nil && info.IsDir() {
			_ = watcher.Add(path)
		}
	}
}

// watchLoop debounces events and sends one update back to the UI loop.
func (a *App) watchLoop(ctx context.Context, watcher *fsnotify.Watcher, root string, limit int, start *workspace) {
	timer := time.NewTimer(time.Hour)
	if !timer.Stop() {
		<-timer.C
	}

	pending := false
	rescan := false
	gitChange := false
	last := start

	flush := func() {
		pending = false

		update := watchUpdate{Git: gitChange}

		if rescan {
			next, delta, err := rescanDiskWorkspace(root, last, limit)
			if err == nil {
				update.Workspace = next
				update.Delta = delta
				last = next
			}
		}

		changes, err := loadLiveChanges(ctx, root)
		if err == nil {
			update.Changes = changes
			update.Branch = changes.Branch
		}

		if branches, head, branchErr := loadBranches(ctx, root); branchErr == nil {
			update.Branches = branches
			update.Branch = head
		}

		select {
		case a.watchUpdates <- update:
		default:
			select {
			case <-a.watchUpdates:
			default:
			}

			select {
			case a.watchUpdates <- update:
			default:
			}
		}

		rescan = false
		gitChange = false
		a.page.WantFrameIn(0)
	}

	for {
		select {
		case <-ctx.Done():
			timer.Stop()

			return
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			if event.Op&fsnotify.Create != 0 {
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() && !skipDir(filepath.Base(event.Name)) {
					_ = a.addWatchTree(watcher, event.Name)
				}
			}

			if isWatchIgnored(event.Name) {
				continue
			}

			if isGitMetadataChange(event.Name) {
				gitChange = true
			} else {
				rescan = true
				gitChange = true
			}

			if !pending {
				pending = true
				timer.Reset(watchDebounce)
			}
		case _, ok := <-watcher.Errors:
			if !ok {
				return
			}
		case <-timer.C:
			flush()
		}
	}
}

// tickWatch drains one background update and any finished patch or search.
func (a *App) tickWatch(ctx context.Context) error {
	select {
	case update := <-a.watchUpdates:
		a.applyWatchUpdate(ctx, update)
	default:
	}

	a.pollPatches(ctx)
	a.receiveSearch(ctx)

	return nil
}

// applyWatchUpdate installs a rescan and then the refreshed git state.
func (a *App) applyWatchUpdate(ctx context.Context, update watchUpdate) {
	if update.Workspace != nil {
		a.installRescan(update.Workspace, update.Delta)
	}

	if update.Changes.Branch != "" || len(update.Changes.Files) > 0 || update.Changes.Note != "" {
		a.applyWatchChanges(ctx, update.Changes)
	}

	if update.Git {
		a.applyGitRefresh(ctx, update)
	}
}

// installRescan swaps the workspace, preserves open documents that still
// exist, and drops tabs for files that disappeared.
func (a *App) installRescan(next *workspace, delta workspaceDelta) {
	if next == nil {
		return
	}

	a.ws = next
	a.scanLimit = next.Limit

	a.tabs = keepExistingTabs(a.tabs, next)
	a.tabs2 = keepExistingTabs(a.tabs2, next)
	a.active = clampSessionIndex(a.active, len(a.tabs))
	a.active2 = clampSessionIndex(a.active2, len(a.tabs2))

	for path := range a.docs {
		if next.ByPath[path] == nil {
			delete(a.docs, path)
			delete(a.dirty, path)

			continue
		}
	}

	for _, path := range a.groupPaths() {
		if next.ByPath[path] == nil {
			continue
		}

		if a.dirty[path] || (a.editing && a.editPath() == path) {
			continue
		}

		a.docs[path] = a.ws.openDoc(path)
	}

	a.recent = keepExistingTabs(a.recent, next)
	a.statuses = statusMap(a.changes)
	a.syncFileNav()

	if a.diffPath != "" && findFile(a.changes.Files, a.diffPath) == nil {
		a.selectFirstDiff()
	}

	a.watchNote = rescanNote(delta)
	if a.split && len(a.tabs2) == 0 {
		a.split = false
		a.focusGroup = 0
	}
}

// keepExistingTabs drops paths the rescan no longer lists.
func keepExistingTabs(paths []string, ws *workspace) []string {
	out := make([]string, 0, len(paths))

	for _, path := range paths {
		if ws.ByPath[path] != nil {
			out = append(out, path)
		}
	}

	return out
}

// rescanNote is the short status line for added, removed, or changed files.
func rescanNote(delta workspaceDelta) string {
	parts := make([]string, 0, rescanNoteParts)

	if len(delta.Added) > 0 {
		parts = append(parts, fmt.Sprintf("%d added", len(delta.Added)))
	}

	if len(delta.Removed) > 0 {
		parts = append(parts, fmt.Sprintf("%d removed", len(delta.Removed)))
	}

	if len(delta.Changed) > 0 {
		parts = append(parts, fmt.Sprintf("%d changed", len(delta.Changed)))
	}

	if len(parts) == 0 {
		return ""
	}

	return "Workspace rescan: " + strings.Join(parts, ", ")
}

// applyWatchChanges updates the diff without moving the user's selection.
func (a *App) applyWatchChanges(ctx context.Context, changes changeSet) {
	a.changes = changes
	a.statuses = statusMap(a.changes)
	a.syncFileNav()

	if a.diffSource == "working" {
		if findFile(a.changes.Files, a.diffPath) == nil {
			if len(a.changes.Files) > 0 {
				a.diffPath = a.changes.Files[0].Path
			} else {
				a.diffPath = ""
			}
		}

		for i, file := range a.changes.Files {
			if file.Path == a.diffPath {
				a.changePage = i / a.changePageRows()

				break
			}
		}
	}

	for _, path := range a.groupPaths() {
		if a.ws.ByPath[path] == nil {
			continue
		}

		if a.dirty[path] || (a.editing && a.editPath() == path) {
			continue
		}

		a.docs[path] = a.ws.openDoc(path)
	}

	if a.findOpen {
		a.reindexFind()
	}

	a.rebuildDiffRows()
	a.scrollY = a.clampScroll(a.scrollY)
	a.view = a.buildView()
	a.page.SetData(&a.view)
	_ = a.page.Redraw(ctx)
}

// applyGitRefresh installs a new branch list and reloads the visible history.
func (a *App) applyGitRefresh(ctx context.Context, update watchUpdate) {
	if len(update.Branches) > 0 {
		a.branches = update.Branches
	}

	if update.Branch != "" {
		a.headBranch = update.Branch
	}

	if a.historyRef != "" {
		if _, ok := branchByRef(a.branches, a.historyRef); !ok {
			a.historyRef = ""
		}
	}

	history, err := loadCommitLog(ctx, a.ws.Root, a.historyRef)
	if err != nil {
		return
	}

	if a.historyRef != "" {
		a.branchLogs.Put(a.historyRef, history)
	}

	a.commits = history
	a.historyLabel = history.Branch

	if a.commitOpen != "" && a.findCommit(a.commitOpen) == nil {
		a.commitOpen = ""
		a.commitFile = ""
		a.commitRows = nil
		a.commitPreview = nil

		if a.diffSource == "commit" {
			a.diffSource = "working"
			a.rebuildDiffRows()
		}
	}

	a.refresh()
}

// isWatchIgnored filters editor churn and git object writes. HEAD, refs, and
// the index stay visible because isGitMetadataChange accepts them first.
func isWatchIgnored(path string) bool {
	slash := filepath.ToSlash(path)

	if isGitMetadataChange(path) {
		return false
	}

	if strings.Contains(slash, "/.git/") {
		return true
	}

	return strings.HasSuffix(slash, ".swp") ||
		strings.HasSuffix(slash, ".swx") ||
		strings.HasSuffix(slash, "~") ||
		strings.HasSuffix(slash, ".tmp")
}

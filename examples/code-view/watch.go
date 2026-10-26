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
	watchDebounce = 150 * time.Millisecond
	maxWatchDirs  = 2000
)

// StartWatcher watches the live workspace and wakes the UI when git changes.
// It is desktop only; Serve never runs the page tick.
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

	watchCtx, cancel := context.WithCancel(ctx)
	a.watch = watcher
	a.watchCancel = cancel
	a.watchChanges = make(chan changeSet, 1)
	a.page.SetTick(a.tickWatch)

	go a.watchLoop(watchCtx, watcher)

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
	a.page.SetTick(nil)

	for {
		select {
		case <-a.watchChanges:
		default:
			return
		}
	}
}

// addWatchTree adds every skipped directory under root.
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

// watchLoop debounces events and sends one change set back to the UI loop.
func (a *App) watchLoop(ctx context.Context, watcher *fsnotify.Watcher) {
	timer := time.NewTimer(time.Hour)
	if !timer.Stop() {
		<-timer.C
	}

	pending := false

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

			if !pending {
				pending = true
				timer.Reset(watchDebounce)
			}
		case _, ok := <-watcher.Errors:
			if !ok {
				return
			}
		case <-timer.C:
			pending = false

			changes, err := loadLiveChanges(ctx, a.ws.Root)
			if err != nil {
				continue
			}

			select {
			case a.watchChanges <- changes:
			default:
			}

			a.page.WantFrameIn(0)
		}
	}
}

// tickWatch drains one background change set on the page loop.
func (a *App) tickWatch(ctx context.Context) error {
	select {
	case changes := <-a.watchChanges:
		a.applyWatchChanges(ctx, changes)
	default:
	}

	return nil
}

// applyWatchChanges updates the diff without moving the user's selection.
func (a *App) applyWatchChanges(ctx context.Context, changes changeSet) {
	a.changes = changes
	a.statuses = statusMap(a.changes)

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

	for _, path := range a.tabs {
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

// isWatchIgnored filters editor and VCS churn that cannot change the diff.
func isWatchIgnored(path string) bool {
	slash := filepath.ToSlash(path)

	return strings.Contains(slash, "/.git/") ||
		strings.HasSuffix(slash, ".swp") ||
		strings.HasSuffix(slash, ".swx") ||
		strings.HasSuffix(slash, "~") ||
		strings.HasSuffix(slash, ".tmp")
}

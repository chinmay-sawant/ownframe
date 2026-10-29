package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

// scmGuard blocks git writes in the bundled sample workspace.
func (a *App) scmGuard() bool {
	if a.ws.Sample {
		a.showToast("info", "Sample workspace", "The bundled fixture is read-only. Use -root for a live repository.")
		a.refresh()

		return true
	}

	if a.ws.Root == "" {
		a.showToast("error", "No workspace", "Open a folder before changing files.")
		a.refresh()

		return true
	}

	return false
}

// workingStagedAll returns staged files, falling back to empty when the set
// was never split.
func (a *App) workingStagedAll() []changedFile {
	return a.changes.Staged
}

// workingUnstagedAll returns unstaged files. Synthetic change sets that only
// fill Files read as unstaged so old fixtures keep rendering.
func (a *App) workingUnstagedAll() []changedFile {
	if len(a.changes.Staged) == 0 && len(a.changes.Unstaged) == 0 {
		return a.changes.Files
	}

	return a.changes.Unstaged
}

// workingAllFiles returns staged rows then unstaged rows for display and
// scrolling. A path with both staged and unstaged edits appears twice.
func (a *App) workingAllFiles() []changedFile {
	staged := a.workingStagedAll()
	unstaged := a.workingUnstagedAll()

	out := make([]changedFile, 0, len(staged)+len(unstaged))
	out = append(out, staged...)
	out = append(out, unstaged...)

	return out
}

// workingAllPaths returns unique working-tree paths in display order.
func (a *App) workingAllPaths() []string {
	seen := map[string]bool{}
	var out []string

	for _, file := range a.workingAllFiles() {
		if seen[file.Path] {
			continue
		}

		seen[file.Path] = true
		out = append(out, file.Path)
	}

	return out
}

// isStaged reports whether a path has a staged entry.
func (a *App) isStaged(path string) bool {
	return findFile(a.changes.Staged, path) != nil
}

// scmChangeFile returns the working-tree entry for a validated path.
func (a *App) scmChangeFile(path string) (changedFile, bool) {
	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(path)))
	if clean == "" || clean == "." || filepath.IsAbs(clean) ||
		clean == ".." || strings.HasPrefix(clean, "../") {
		return changedFile{}, false
	}

	if file := findFile(a.changes.Staged, clean); file != nil {
		return *file, true
	}

	if file := findFile(a.changes.Unstaged, clean); file != nil {
		return *file, true
	}

	file := findFile(a.changes.Files, clean)
	if file == nil {
		return changedFile{}, false
	}

	return *file, true
}

// ensureSelected makes the multi-select map usable.
func (a *App) ensureSelected() {
	if a.selected == nil {
		a.selected = map[string]bool{}
	}
}

// pruneSelected drops multi-select paths that left the working tree.
func (a *App) pruneSelected() {
	a.ensureSelected()

	valid := map[string]bool{}
	for _, path := range a.workingAllPaths() {
		valid[path] = true
	}

	for path := range a.selected {
		if !valid[path] {
			delete(a.selected, path)
		}
	}
}

// setSingleSelect replaces the multi-select set with one path.
func (a *App) setSingleSelect(path string) {
	a.ensureSelected()

	for key := range a.selected {
		delete(a.selected, key)
	}

	if path != "" {
		a.selected[path] = true
	}
}

// toggleSelect flips one path in the multi-select set for Ctrl-click.
func (a *App) toggleSelect(path string) {
	a.ensureSelected()

	if a.selected[path] {
		// Keep at least the preview file selected when the set would go
		// empty from a Ctrl-click.
		if len(a.selected) == 1 {
			return
		}

		delete(a.selected, path)

		return
	}

	a.selected[path] = true
}

// selectAllFiles selects every working-tree path for Ctrl+A.
func (a *App) selectAllFiles() {
	a.ensureSelected()

	for _, path := range a.workingAllPaths() {
		a.selected[path] = true
	}
}

// selectedList returns the multi-select set in display order.
func (a *App) selectedList() []string {
	a.ensureSelected()

	ordered := a.workingAllPaths()
	out := make([]string, 0, len(a.selected))

	for _, path := range ordered {
		if a.selected[path] {
			out = append(out, path)
		}
	}

	// Paths that are selected but no longer visible stay at the end so a
	// bulk action does not silently drop them.
	var extra []string

	for path := range a.selected {
		found := false

		for _, existing := range out {
			if existing == path {
				found = true

				break
			}
		}

		if !found {
			extra = append(extra, path)
		}
	}

	sort.Strings(extra)

	return append(out, extra...)
}

// selectedCount counts selected paths that still exist.
func (a *App) selectedCount() int {
	n := 0

	for _, path := range a.workingAllPaths() {
		if a.selected[path] {
			n++
		}
	}

	return n
}

// discardFile reverts one working-tree file. Tracked files go through
// git restore, untracked files are removed from disk.
func (a *App) discardFile(ctx context.Context, path string) error {
	if a.scmGuard() {
		return nil
	}

	file, ok := a.scmChangeFile(path)
	if !ok {
		return nil
	}

	if err := a.discardPaths(ctx, []string{file.Path}); err != nil {
		return err
	}

	a.showToast("success", "Changes discarded", file.Path)
	a.refresh()

	return nil
}

// discardPaths reverts every listed path without asking. The click and menu
// layers open the confirm dialog first.
func (a *App) discardPaths(ctx context.Context, paths []string) error {
	var failed []string
	seen := map[string]bool{}

	for _, path := range paths {
		if seen[path] {
			continue
		}

		seen[path] = true
		file, ok := a.scmChangeFile(path)

		if !ok {
			continue
		}

		if file.Status == "U" {
			full := filepath.Join(a.ws.Root, filepath.FromSlash(file.Path))
			if err := os.Remove(full); err != nil && !os.IsNotExist(err) {
				failed = append(failed, file.Path)
			}

			a.forgetSCMDoc(file.Path)

			continue
		}

		restore := []string{file.Path}
		if file.OldPath != "" && file.OldPath != file.Path {
			restore = append(restore, file.OldPath)
		}

		args := append([]string{"restore", "--source=HEAD", "--staged", "--worktree", "--"}, restore...)
		if _, err := gitBytes(ctx, a.ws.Root, args...); err != nil {
			failed = append(failed, file.Path)
		}

		a.forgetSCMDoc(file.Path)
		if file.OldPath != "" && file.OldPath != file.Path {
			a.forgetSCMDoc(file.OldPath)
		}
	}

	a.refreshAfterSCMMutation(ctx)

	if len(failed) > 0 {
		a.showToast("error", "Discard incomplete", strings.Join(failed, ", "))
	}

	return nil
}

// stageFile stages one working-tree file with git add.
func (a *App) stageFile(ctx context.Context, path string) error {
	if a.scmGuard() {
		return nil
	}

	file, ok := a.scmChangeFile(path)
	if !ok {
		return nil
	}

	if _, err := gitBytes(ctx, a.ws.Root, "add", "--", file.Path); err != nil {
		a.showToast("error", "Stage failed", err.Error())
		a.refresh()

		return nil
	}

	a.refreshAfterSCMMutation(ctx)
	a.showToast("success", "Changes staged", file.Path)
	a.refresh()

	return nil
}

// unstageFile removes one file from the index, keeping worktree edits.
func (a *App) unstageFile(ctx context.Context, path string) error {
	if a.scmGuard() {
		return nil
	}

	file, ok := a.scmChangeFile(path)
	if !ok {
		return nil
	}

	if _, err := gitBytes(ctx, a.ws.Root, "restore", "--staged", "--", file.Path); err != nil {
		a.showToast("error", "Unstage failed", err.Error())
		a.refresh()

		return nil
	}

	a.refreshAfterSCMMutation(ctx)
	a.showToast("success", "Changes unstaged", file.Path)
	a.refresh()

	return nil
}

// discardAll reverts every working-tree file.
func (a *App) discardAll(ctx context.Context) error {
	if a.scmGuard() {
		return nil
	}

	if len(a.changes.Files) == 0 {
		return nil
	}

	if err := a.discardPaths(ctx, a.workingAllPaths()); err != nil {
		return err
	}

	a.showToast("success", "All changes discarded", fmt.Sprintf("%d files", len(a.changes.Files)))
	a.refresh()

	return nil
}

// stageAll stages the whole working tree.
func (a *App) stageAll(ctx context.Context) error {
	if a.scmGuard() {
		return nil
	}

	if len(a.changes.Files) == 0 {
		return nil
	}

	if _, err := gitBytes(ctx, a.ws.Root, "add", "--all", "--", "."); err != nil {
		a.showToast("error", "Stage failed", err.Error())
		a.refresh()

		return nil
	}

	a.refreshAfterSCMMutation(ctx)
	a.showToast("success", "All changes staged", fmt.Sprintf("%d files", len(a.changes.Files)))
	a.refresh()

	return nil
}

// unstageAll removes every staged file from the index.
func (a *App) unstageAll(ctx context.Context) error {
	if a.scmGuard() {
		return nil
	}

	if len(a.changes.Staged) == 0 {
		return nil
	}

	if _, err := gitBytes(ctx, a.ws.Root, "restore", "--staged", "--", "."); err != nil {
		a.showToast("error", "Unstage failed", err.Error())
		a.refresh()

		return nil
	}

	a.refreshAfterSCMMutation(ctx)
	a.showToast("success", "All staged changes unstaged", fmt.Sprintf("%d files", len(a.changes.Staged)))
	a.refresh()

	return nil
}

// stageSelected stages the multi-select set.
func (a *App) stageSelected(ctx context.Context) error {
	if a.scmGuard() {
		return nil
	}

	paths := a.selectedList()
	if len(paths) == 0 {
		return nil
	}

	args := append([]string{"add", "--"}, paths...)
	if _, err := gitBytes(ctx, a.ws.Root, args...); err != nil {
		a.showToast("error", "Stage failed", err.Error())
		a.refresh()

		return nil
	}

	a.refreshAfterSCMMutation(ctx)
	a.showToast("success", "Selected changes staged", fmt.Sprintf("%d files", len(paths)))
	a.refresh()

	return nil
}

// unstageSelected removes the multi-select set from the index.
func (a *App) unstageSelected(ctx context.Context) error {
	if a.scmGuard() {
		return nil
	}

	paths := a.selectedList()
	if len(paths) == 0 {
		return nil
	}

	args := append([]string{"restore", "--staged", "--"}, paths...)
	if _, err := gitBytes(ctx, a.ws.Root, args...); err != nil {
		a.showToast("error", "Unstage failed", err.Error())
		a.refresh()

		return nil
	}

	a.refreshAfterSCMMutation(ctx)
	a.showToast("success", "Selected changes unstaged", fmt.Sprintf("%d files", len(paths)))
	a.refresh()

	return nil
}

// discardSelected reverts the multi-select set after confirm.
func (a *App) discardSelected(ctx context.Context) error {
	if a.scmGuard() {
		return nil
	}

	paths := a.selectedList()
	if len(paths) == 0 {
		return nil
	}

	if err := a.discardPaths(ctx, paths); err != nil {
		return err
	}

	a.showToast("success", "Selected changes discarded", fmt.Sprintf("%d files", len(paths)))
	a.refresh()

	return nil
}

// openDiscardConfirm shows the discard dialog for one file, the selection,
// or the whole tree.
func (a *App) openDiscardConfirm(kind string, paths []string) {
	if a.scmGuard() {
		return
	}

	unique := make([]string, 0, len(paths))
	seen := map[string]bool{}

	for _, path := range paths {
		if path == "" || seen[path] {
			continue
		}

		seen[path] = true

		if _, ok := a.scmChangeFile(path); ok {
			unique = append(unique, path)
		}
	}

	if len(unique) == 0 {
		return
	}

	sort.Strings(unique)

	a.confirmOpen = true
	a.confirmKind = kind
	a.confirmPaths = unique

	switch kind {
	case "discard-selected":
		a.confirmTitle = fmt.Sprintf("Discard %d files?", len(unique))
	case "discard-all":
		a.confirmTitle = fmt.Sprintf("Discard all %d files?", len(unique))
	default:
		a.confirmTitle = "Discard changes?"
	}

	a.confirmMsg = "Do you really want to discard these changes? This cannot be undone."
	a.refresh()
}

// closeConfirm hides the discard dialog without acting.
func (a *App) closeConfirm() {
	a.confirmOpen = false
	a.confirmKind = ""
	a.confirmPaths = nil
	a.confirmTitle = ""
	a.confirmMsg = ""
	a.refresh()
}

// confirmDiscard runs the pending discard after the dialog confirms.
func (a *App) confirmDiscard(ctx context.Context) error {
	if !a.confirmOpen {
		return nil
	}

	paths := append([]string(nil), a.confirmPaths...)
	kind := a.confirmKind
	a.closeConfirm()

	switch kind {
	case "discard-selected", "discard-all":
		if err := a.discardPaths(ctx, paths); err != nil {
			return err
		}

		a.showToast("success", "Changes discarded", fmt.Sprintf("%d files", len(paths)))
	default:
		if len(paths) == 0 {
			return nil
		}

		if err := a.discardPaths(ctx, paths); err != nil {
			return err
		}

		a.showToast("success", "Changes discarded", paths[0])
	}

	a.refresh()

	return nil
}

// confirmViewState publishes the discard dialog.
func (a *App) confirmViewState() confirmView {
	if !a.confirmOpen {
		return confirmView{}
	}

	files := append([]string(nil), a.confirmPaths...)
	if len(files) > 6 {
		files = files[:6]
	}

	return confirmView{
		Show:    true,
		Title:   a.confirmTitle,
		Message: a.confirmMsg,
		Files:   files,
		Count:   len(a.confirmPaths),
	}
}

// copyCommitMessage copies the subject and body of one commit.
func (a *App) copyCommitMessage(hash string) {
	commit := a.findCommit(hash)
	if commit == nil {
		return
	}

	text := commit.Subject
	if commit.Body != "" {
		text += "\n\n" + commit.Body
	}

	ownframe.Write(text)
	a.showToast("success", "Message copied", commit.Short)
	a.refresh()
}

// copyCommitID copies the full commit hash.
func (a *App) copyCommitID(hash string) {
	commit := a.findCommit(hash)
	if commit == nil {
		return
	}

	ownframe.Write(commit.Hash)
	a.showToast("success", "Commit ID copied", commit.Short)
	a.refresh()
}

// forgetSCMDoc drops the open document and dirty state for a path whose
// bytes changed under git, so the next open reads from disk.
func (a *App) forgetSCMDoc(path string) {
	if a.editing && a.editPath() == path {
		a.finishEdit()
	}

	delete(a.docs, path)
	delete(a.dirty, path)
}

// refreshAfterSCMMutation rescans the workspace and reloads the change set
// without announcing the routine refresh toast.
func (a *App) refreshAfterSCMMutation(ctx context.Context) {
	if !a.ws.Sample && a.ws.Root != "" {
		if next, delta, err := rescanDiskWorkspace(a.ws.Root, a.ws, a.scanLimit); err == nil {
			a.installRescan(next, delta)
		}
	}

	a.reloadChanges(ctx, false)
	a.pruneSelected()
}

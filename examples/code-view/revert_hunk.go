package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

// revertUndoLimit caps the hunk restore stack. Patches are small, so 50
// entries stay bounded while covering consecutive mistaken reverts.
const revertUndoLimit = 50

// revertUndoEntry stores the forward patches to restore one reverted hunk.
// WorktreePatch is the `git diff HEAD` hunk(s); StagedPatch is the
// `git diff --cached` hunk(s). Untracked files keep the deleted bytes
// instead since no git patch exists for them.
type revertUndoEntry struct {
	Path          string
	Status        string
	WorktreePatch string
	StagedPatch   string
	DeletedBytes  []byte
	DeletedMode   os.FileMode
}

// revertHunk reverts one @@ hunk in the working-tree diff, VS Code-style.
// The hunk header button in the unified and split views calls it with the
// index into the file's Hunks slice. It reverses the matching raw git hunks
// for the worktree and the index, so staged and unstaged edits in that block
// both return to HEAD without touching other hunks.
func (a *App) revertHunk(ctx context.Context, hunkIdx int) error {
	if a.scmGuard() {
		return nil
	}

	if a.diffSource != "working" {
		a.showToast("info", "Working tree only", "Hunk revert is available for working-tree changes.")
		a.refresh()

		return nil
	}

	file := findFile(a.changes.Files, a.diffPath)
	if file == nil {
		return nil
	}

	if file.Binary || file.Oversized {
		return nil
	}

	if file.Status == "R" {
		a.showToast("info", "Rename revert", "Revert renames at the file level to keep the rename.")
		a.refresh()

		return nil
	}

	if hunkIdx < 0 || hunkIdx >= len(file.Hunks) {
		return nil
	}

	if file.Status == "U" {
		full := filepath.Join(a.ws.Root, filepath.FromSlash(file.Path))
		raw, readErr := os.ReadFile(full) //nolint:gosec // path came from the scanned change list under ws.Root
		var mode os.FileMode
		if info, statErr := os.Stat(full); statErr == nil {
			mode = info.Mode()
		}
		if err := a.discardPaths(ctx, []string{file.Path}); err != nil {
			return err
		}

		if readErr == nil {
			a.pushRevertUndo(revertUndoEntry{Path: file.Path, Status: file.Status, DeletedBytes: raw, DeletedMode: mode})
		}
		a.showToast("success", "Hunk reverted", file.Path+" (Ctrl+Z restores)")
		a.refresh()

		return nil
	}

	uiHunk := file.Hunks[hunkIdx]

	// Added files stage the whole file, so the index must go last: unstaging
	// first would turn the worktree file untracked, which `git diff HEAD`
	// no longer reports, leaving the file behind. Other statuses revert
	// the index first so a worktree failure stays retryable as unstaged.
	if file.Status == "A" {
		return a.revertAddedHunk(ctx, file, uiHunk)
	}

	// Staged hunks overlap the combined hunk's HEAD range. Revert them in
	// the index first so a worktree failure stays retryable as unstaged.
	stagedPatch, hasStaged := a.stagedHunkPatch(ctx, file.Path, uiHunk.OldStart, uiHunk.OldCount)
	if hasStaged {
		if _, err := gitBytesStdin(ctx, a.ws.Root, stagedPatch, "apply", "--cached", "--reverse", "--whitespace=nowarn"); err != nil {
			a.showToast("error", "Hunk revert failed", err.Error())
			a.refresh()

			return nil
		}
	}

	patch, ok := a.worktreeHunkPatch(ctx, file.Path, uiHunk.OldStart, uiHunk.OldCount)
	if !ok {
		a.showToast("info", "Hunk already reverted", "The working tree no longer has this hunk.")
		a.refresh()

		return nil
	}

	if _, err := gitBytesStdin(ctx, a.ws.Root, patch, "apply", "--reverse", "--whitespace=nowarn"); err != nil {
		a.showToast("error", "Hunk revert failed", err.Error())
		a.refreshAfterSCMMutation(ctx)
		a.refresh()

		return nil
	}

	a.pushRevertUndo(revertUndoEntry{Path: file.Path, Status: file.Status, WorktreePatch: patch, StagedPatch: stagedPatch})
	a.refreshAfterSCMMutation(ctx)
	a.showToast("success", "Hunk reverted", file.Path+" (Ctrl+Z restores)")
	a.refresh()

	return nil
}

// revertAddedHunk reverts one hunk of a staged new file. The worktree file
// goes first while `git diff HEAD` still reports it; the index follows.
func (a *App) revertAddedHunk(ctx context.Context, file *changedFile, uiHunk hunk) error {
	patch, ok := a.worktreeHunkPatch(ctx, file.Path, uiHunk.OldStart, uiHunk.OldCount)
	if !ok {
		a.showToast("info", "Hunk already reverted", "The working tree no longer has this hunk.")
		a.refresh()

		return nil
	}

	if _, err := gitBytesStdin(ctx, a.ws.Root, patch, "apply", "--reverse", "--whitespace=nowarn"); err != nil {
		a.showToast("error", "Hunk revert failed", err.Error())
		a.refresh()

		return nil
	}

	staged, hasStaged := a.stagedHunkPatch(ctx, file.Path, uiHunk.OldStart, uiHunk.OldCount)
	if hasStaged {
		if _, err := gitBytesStdin(ctx, a.ws.Root, staged, "apply", "--cached", "--reverse", "--whitespace=nowarn"); err != nil {
			a.showToast("error", "Hunk revert failed", err.Error())
			a.refreshAfterSCMMutation(ctx)
			a.refresh()

			return nil
		}
	}

	a.pushRevertUndo(revertUndoEntry{Path: file.Path, Status: file.Status, WorktreePatch: patch, StagedPatch: staged})
	a.refreshAfterSCMMutation(ctx)
	a.showToast("success", "Hunk reverted", file.Path+" (Ctrl+Z restores)")
	a.refresh()

	return nil
}

// stagedHunkPatch builds a cached reverse patch for the staged hunks whose
// HEAD range overlaps one combined hunk. It returns false when the index
// has no overlapping hunk.
func (a *App) stagedHunkPatch(ctx context.Context, path string, oldStart, oldCount int) (string, bool) {
	raw, err := gitBytes(ctx, a.ws.Root, "diff", "--cached", "--no-color", "--no-ext-diff", "-U3", "--", path)
	if err != nil || len(raw) == 0 {
		return "", false
	}

	patch := selectRawHunks(string(raw), oldStart, oldCount)
	if patch == "" {
		return "", false
	}

	return patch, true
}

// worktreeHunkPatch builds a worktree reverse patch for the combined hunks
// whose HEAD range overlaps one displayed hunk. Raw git text keeps tabs and
// line endings intact, which the parsed view expands.
func (a *App) worktreeHunkPatch(ctx context.Context, path string, oldStart, oldCount int) (string, bool) {
	raw, err := gitBytes(ctx, a.ws.Root, "diff", "HEAD", "--no-color", "--no-ext-diff", "-U3", "--", path)
	if err != nil || len(raw) == 0 {
		return "", false
	}

	patch := selectRawHunks(string(raw), oldStart, oldCount)
	if patch == "" {
		return "", false
	}

	return patch, true
}

// selectRawHunks keeps the file header plus every raw @@ block whose HEAD
// range overlaps the displayed hunk. Verbatim text preserves tabs so git
// apply sees the same context git produced.
func selectRawHunks(raw string, oldStart, oldCount int) string {
	lines := strings.Split(raw, "\n")

	diffIdx := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "diff --git ") {
			diffIdx = i

			break
		}
	}

	if diffIdx < 0 {
		return ""
	}

	firstHunk := -1
	for i := diffIdx; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "@@ ") {
			firstHunk = i

			break
		}
	}

	if firstHunk < 0 {
		return ""
	}

	header := lines[diffIdx:firstHunk]

	type block struct {
		start int
		end   int
		old   int
		count int
	}

	var blocks []block

	for i := firstHunk; i < len(lines); {
		line := lines[i]
		if strings.HasPrefix(line, "diff --git ") {
			break
		}

		if !strings.HasPrefix(line, "@@ ") {
			i++

			continue
		}

		parsed := parseHunkHeader(line)
		if parsed == nil {
			i++

			continue
		}

		j := i + 1
		for j < len(lines) && !strings.HasPrefix(lines[j], "@@ ") && !strings.HasPrefix(lines[j], "diff --git ") {
			j++
		}

		blocks = append(blocks, block{start: i, end: j, old: parsed.OldStart, count: parsed.OldCount})
		i = j
	}

	var out []string
	out = append(out, header...)

	matched := false
	for _, b := range blocks {
		if hunkRangesOverlap(oldStart, oldCount, b.old, b.count) {
			out = append(out, lines[b.start:b.end]...)
			matched = true
		}
	}

	if !matched {
		return ""
	}

	patch := strings.Join(out, "\n")
	if !strings.HasSuffix(patch, "\n") {
		patch += "\n"
	}

	return patch
}

// hunkRangesOverlap reports whether two HEAD ranges share a line. A zero
// count is an insertion point, which overlaps when it sits inside the other
// range or matches another insertion at the same line.
func hunkRangesOverlap(aStart, aCount, bStart, bCount int) bool {
	if aCount == 0 && bCount == 0 {
		return aStart == bStart
	}

	if aCount == 0 {
		return bStart <= aStart && aStart <= bStart+bCount
	}

	if bCount == 0 {
		return aStart <= bStart && bStart <= aStart+aCount
	}

	return bStart < aStart+aCount && aStart < bStart+bCount
}

// pushRevertUndo records one successful revert and clears the redo stack.
// It never touches the text edit stacks.
func (a *App) pushRevertUndo(e revertUndoEntry) {
	a.revertUndo = append(a.revertUndo, e)
	if len(a.revertUndo) > revertUndoLimit {
		a.revertUndo = append([]revertUndoEntry(nil), a.revertUndo[len(a.revertUndo)-revertUndoLimit:]...)
	}
	a.revertRedo = nil
}

// canUndoRevert reports whether a hunk restore is available. Sample
// workspaces have no git history to restore, and commit/branch views are
// read-only, so undo stays in the working tree.
func (a *App) canUndoRevert() bool {
	return !a.ws.Sample && len(a.revertUndo) > 0 && a.diffSource == "working"
}

// canRedoRevert reports whether a re-revert is available.
func (a *App) canRedoRevert() bool {
	return !a.ws.Sample && len(a.revertRedo) > 0 && a.diffSource == "working"
}

// undoRevertHunk restores the most recently reverted hunk by re-applying its
// forward patches: index first, then worktree. Untracked files are rewritten
// from the saved bytes.
func (a *App) undoRevertHunk(ctx context.Context) error {
	if len(a.revertUndo) == 0 {
		return nil
	}

	entry := a.revertUndo[len(a.revertUndo)-1]

	if entry.Status == "U" {
		full := filepath.Join(a.ws.Root, filepath.FromSlash(entry.Path))
		mode := entry.DeletedMode
		if mode == 0 {
			mode = workspaceFileMode
		}
		if err := os.MkdirAll(filepath.Dir(full), workspaceDirMode); err != nil {
			a.showToast("error", "Restore failed", err.Error())
			a.refresh()

			return nil
		}
		if err := os.WriteFile(full, entry.DeletedBytes, mode); err != nil {
			a.showToast("error", "Restore failed", err.Error())
			a.refresh()

			return nil
		}

		a.revertUndo = a.revertUndo[:len(a.revertUndo)-1]
		a.revertRedo = append(a.revertRedo, entry)
		a.forgetSCMDoc(entry.Path)
		a.refreshAfterSCMMutation(ctx)
		a.showToast("success", "Hunk restored", entry.Path)
		a.refresh()

		return nil
	}

	if entry.StagedPatch != "" {
		if _, err := gitBytesStdin(ctx, a.ws.Root, entry.StagedPatch, "apply", "--cached", "--whitespace=nowarn"); err != nil {
			a.showToast("error", "Restore failed", err.Error())
			a.refresh()

			return nil
		}
	}

	if entry.WorktreePatch != "" {
		if _, err := gitBytesStdin(ctx, a.ws.Root, entry.WorktreePatch, "apply", "--whitespace=nowarn"); err != nil {
			a.showToast("error", "Restore failed", err.Error())
			a.refreshAfterSCMMutation(ctx)
			a.refresh()

			return nil
		}
	}

	a.revertUndo = a.revertUndo[:len(a.revertUndo)-1]
	a.revertRedo = append(a.revertRedo, entry)
	a.forgetSCMDoc(entry.Path)
	a.refreshAfterSCMMutation(ctx)
	a.showToast("success", "Hunk restored", entry.Path)
	a.refresh()

	return nil
}

// redoRevertHunk re-applies the last undone revert in its original order:
// worktree first for added files (so `git diff HEAD` still reports the
// file), otherwise index first so a worktree failure stays retryable.
func (a *App) redoRevertHunk(ctx context.Context) error {
	if len(a.revertRedo) == 0 {
		return nil
	}

	entry := a.revertRedo[len(a.revertRedo)-1]

	applyReverse := func(patch, scope string) error {
		if patch == "" {
			return nil
		}
		args := []string{"apply", "--reverse", "--whitespace=nowarn"}
		if scope == "cached" {
			args = []string{"apply", "--cached", "--reverse", "--whitespace=nowarn"}
		}
		_, err := gitBytesStdin(ctx, a.ws.Root, patch, args...)
		return err
	}

	if entry.Status == "U" {
		full := filepath.Join(a.ws.Root, filepath.FromSlash(entry.Path))
		if err := os.Remove(full); err != nil && !os.IsNotExist(err) {
			a.showToast("error", "Re-revert failed", err.Error())
			a.refresh()

			return nil
		}

		a.revertRedo = a.revertRedo[:len(a.revertRedo)-1]
		a.revertUndo = append(a.revertUndo, entry)
		a.forgetSCMDoc(entry.Path)
		a.refreshAfterSCMMutation(ctx)
		a.showToast("success", "Hunk re-reverted", entry.Path)
		a.refresh()

		return nil
	}

	if entry.Status == "A" {
		if err := applyReverse(entry.WorktreePatch, "worktree"); err != nil {
			a.showToast("error", "Re-revert failed", err.Error())
			a.refresh()

			return nil
		}
		if err := applyReverse(entry.StagedPatch, "cached"); err != nil {
			a.showToast("error", "Re-revert failed", err.Error())
			a.refreshAfterSCMMutation(ctx)
			a.refresh()

			return nil
		}
	} else {
		if err := applyReverse(entry.StagedPatch, "cached"); err != nil {
			a.showToast("error", "Re-revert failed", err.Error())
			a.refresh()

			return nil
		}
		if err := applyReverse(entry.WorktreePatch, "worktree"); err != nil {
			a.showToast("error", "Re-revert failed", err.Error())
			a.refreshAfterSCMMutation(ctx)
			a.refresh()

			return nil
		}
	}

	a.revertRedo = a.revertRedo[:len(a.revertRedo)-1]
	a.revertUndo = append(a.revertUndo, entry)
	a.forgetSCMDoc(entry.Path)
	a.refreshAfterSCMMutation(ctx)
	a.showToast("success", "Hunk re-reverted", entry.Path)
	a.refresh()

	return nil
}

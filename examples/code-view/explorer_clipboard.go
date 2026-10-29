package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

// PendingCopy is one file staged by Copy for a later Paste. Task 4's paste
// backend consumes it; the explorer clipboard owns the definition so every
// workstream shares one shape.
type PendingCopy struct {
	SrcPath string
	IsCut   bool
}

// pasteStagedCopy is the Task 3 name for a staged file copy. It aliases
// PendingCopy so the App seam and this package share one shape.
type pasteStagedCopy = PendingCopy

// pendingCopy holds the file Copy staged. Copy keeps it for repeated pastes;
// a cut clears it after one paste.
var pendingCopy *PendingCopy

// stageCopy records src for a later Paste.
func stageCopy(src string, isCut bool) {
	pendingCopy = &PendingCopy{SrcPath: src, IsCut: isCut}
}

// stagedCopy returns the file Copy staged, or nil when nothing is staged.
func stagedCopy() *PendingCopy {
	return pendingCopy
}

// clearStagedCopy drops the staged file, for example between tests.
func clearStagedCopy() {
	pendingCopy = nil
}

// stageExplorerCopy is the Task 3 seam: the explorer Copy menu row stages
// here. It shares the package store so the menu, Ctrl+V, and Task 4's paste
// backend all see one staged file.
func (a *App) stageExplorerCopy(src string, isCut bool) {
	stageCopy(src, isCut)
	a.pasteStaged = pendingCopy
}

// dropExplorerCopy clears the staged file on the app and in the store.
func (a *App) dropExplorerCopy() {
	clearStagedCopy()
	a.pasteStaged = nil
}

// contextMenuForExplorer returns the explorer rows for a tree id. The second
// result reports whether the id belongs to the explorer, including empty
// space and stale paths that yield no rows. Only visible rows have boxes, so
// callers reach this through the id of the row under the pointer.
func (a *App) contextMenuForExplorer(id string) ([]contextMenuRow, bool) {
	switch id {
	case "tree-pane", "sidebar", "":
		return []contextMenuRow{
			{ID: contextMenuPasteID, Label: "Paste", Enabled: stagedCopy() != nil},
		}, true
	}

	if !strings.HasPrefix(id, "tree-") {
		return nil, false
	}

	rel := strings.TrimPrefix(id, "tree-")

	switch rel {
	case "scroll", "thumb", "up", "down", "pane":
		return nil, true
	}

	isDir, ok := a.lookupTreePath(rel)
	if !ok {
		return nil, true
	}

	if isDir {
		return []contextMenuRow{
			{ID: contextMenuCopyPathID + rel, Label: "Copy Path", Enabled: true},
			{ID: contextMenuPasteID + rel, Label: "Paste", Enabled: stagedCopy() != nil},
		}, true
	}

	return []contextMenuRow{
		{ID: contextMenuCopyPathID + rel, Label: "Copy Path", Enabled: true},
		{ID: contextMenuCopyFileID + rel, Label: "Copy File", Enabled: true},
	}, true
}

// lookupTreePath validates a slash-relative tree path and reports whether it
// is a directory. A path counts when the workspace lists it or the current
// windowed view shows it.
func (a *App) lookupTreePath(rel string) (bool, bool) {
	if rel == "" || rel == "." || filepath.IsAbs(rel) {
		return false, false
	}

	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(rel)))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return false, false
	}

	if node := findTreeNode(a.ws.Tree, clean); node != nil {
		return node.Dir, true
	}

	if meta := a.ws.ByPath[clean]; meta != nil {
		return false, true
	}

	for _, row := range a.view.Tree {
		if row.Path == clean {
			return row.Dir, true
		}
	}

	return false, false
}

// copyTreePath writes one slash-relative tree path to the clipboard. It
// mirrors copyActive(false).
func (a *App) copyTreePath(path string) {
	if _, ok := a.lookupTreePath(path); !ok {
		return
	}

	ownframe.Write(path)
	a.showToast("success", "Path copied", path)
	a.refresh()
}

// copyTreeFile writes one tree file to the clipboard and stages it for Paste.
// Text files go over as text; a binary file cannot, so the clipboard keeps
// the path as a fallback while the staged copy still pastes the bytes.
func (a *App) copyTreeFile(path string) {
	isDir, ok := a.lookupTreePath(path)
	if !ok || isDir {
		return
	}

	raw, err := a.ws.readFile(path)
	if err != nil {
		a.showToast("error", "Cannot copy", err.Error())
		a.refresh()

		return
	}

	if isBinary(raw) {
		a.stageExplorerCopy(path, false)
		ownframe.Write(path)
		a.showToast("success", "File copied", path+" (binary; path on clipboard)")
		a.refresh()

		return
	}

	a.stageExplorerCopy(path, false)
	ownframe.Write(string(raw))
	a.showToast("success", "File copied", path)
	a.refresh()
}

// pasteTreeCopy duplicates the staged file into destDir, or the workspace
// root when destDir is empty. Pasting onto a file id lands beside it. This is
// the minimal disk backend Task 4 replaces: it resolves the destination,
// picks a unique name, copies the bytes, refreshes the workspace, and
// repaints.
func (a *App) pasteTreeCopy(ctx context.Context, destDir string) error {
	staged := stagedCopy()
	if staged == nil || staged.SrcPath == "" {
		a.showToast("info", "Nothing to paste", "Copy a file first.")
		a.refresh()

		return a.redrawTreeMenu(ctx)
	}

	if a.ws.Sample {
		a.showToast("error", "Cannot paste", sampleReadOnlyMessage)
		a.refresh()

		return a.redrawTreeMenu(ctx)
	}

	src := staged.SrcPath

	raw, err := a.ws.readFile(src)
	if err != nil {
		a.showToast("error", "Cannot paste", err.Error())
		a.refresh()

		return a.redrawTreeMenu(ctx)
	}

	dest := destDir
	if dest != "" {
		if _, ok := a.lookupTreePath(dest); !ok {
			a.showToast("error", "Cannot paste", "The destination folder is not in the workspace.")
			a.refresh()

			return a.redrawTreeMenu(ctx)
		}

		if meta := a.ws.ByPath[dest]; meta != nil {
			dest = dirName(dest)
		}

		if dest != "" {
			node := findTreeNode(a.ws.Tree, dest)
			if node == nil || !node.Dir {
				a.showToast("error", "Cannot paste", "The destination folder is not in the workspace.")
				a.refresh()

				return a.redrawTreeMenu(ctx)
			}
		}
	}

	candidate := uniquePastePath(a.ws, dest, baseName(src))

	target := filepath.Join(a.ws.Root, filepath.FromSlash(candidate))
	if err := os.MkdirAll(filepath.Dir(target), workspaceDirMode); err != nil {
		a.showToast("error", "Cannot paste", err.Error())
		a.refresh()

		return a.redrawTreeMenu(ctx)
	}

	if err := os.WriteFile(target, raw, workspaceFileMode); err != nil {
		a.showToast("error", "Cannot paste", err.Error())
		a.refresh()

		return a.redrawTreeMenu(ctx)
	}

	a.refreshPastedFile(dest, candidate)

	if dest != "" {
		a.folders[dest] = true
	}

	if staged.IsCut {
		a.dropExplorerCopy()
	}

	a.showToast("success", "File pasted", candidate)
	a.refresh()

	return a.redrawTreeMenu(ctx)
}

// refreshPastedFile makes one pasted file visible. A full scan rescans and
// reinstalls the workspace; a lazy workspace loads the destination folder,
// falling back to a direct insert at the root, which has no folder node to
// load.
func (a *App) refreshPastedFile(dest, candidate string) {
	if a.ws.Lazy {
		if dest != "" {
			_ = a.ws.loadDir(dest)
		}

		if a.ws.ByPath[candidate] == nil {
			_ = a.ws.addDiskFile(candidate)
		}

		return
	}

	next, delta, err := rescanDiskWorkspace(a.ws.Root, a.ws, a.scanLimit)
	if err != nil {
		_ = a.ws.addDiskFile(candidate)

		return
	}

	a.installRescan(next, delta)
}

// uniquePastePath picks dest/base, or a " copy" suffixed name when it exists
// in the workspace or on disk.
func uniquePastePath(w *workspace, dest, base string) string {
	join := func(name string) string {
		if dest == "" {
			return name
		}

		return dest + "/" + name
	}

	if w.ByPath[join(base)] == nil && !pastePathExists(w.Root, join(base)) {
		return join(base)
	}

	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)

	for i := 1; ; i++ {
		name := fmt.Sprintf("%s copy%s", stem, ext)
		if i > 1 {
			name = fmt.Sprintf("%s copy %d%s", stem, i, ext)
		}

		if w.ByPath[join(name)] == nil && !pastePathExists(w.Root, join(name)) {
			return join(name)
		}
	}
}

// pastePathExists reports whether a slash-relative path exists on disk.
func pastePathExists(root, rel string) bool {
	_, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel)))

	return err == nil
}

// addDiskFile registers a file already written under the root: it adds the
// metadata and inserts the tree node under its parent. A missing parent
// folder is an error; loading a folder stays with loadDir.
func (w *workspace) addDiskFile(candidate string) error {
	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(candidate)))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") || filepath.IsAbs(clean) {
		return fmt.Errorf("path escapes workspace: %s", candidate)
	}

	info, err := os.Stat(filepath.Join(w.Root, filepath.FromSlash(clean)))
	if err != nil {
		return err
	}

	if info.IsDir() {
		return fmt.Errorf("%s is a directory", candidate)
	}

	if w.ByPath == nil {
		w.ByPath = map[string]*fileMeta{}
	}

	if w.ByPath[clean] == nil {
		w.addFile(clean, baseName(clean), info.Size())
	}

	meta := w.ByPath[clean]
	dir := dirName(clean)

	if dir == "" {
		if findChild(w.Tree, meta.Name) == nil {
			w.Tree = append(w.Tree, &fileNode{Path: clean, Name: meta.Name, Loaded: true, Meta: meta})
		} else {
			for _, child := range w.Tree {
				if child.Name == meta.Name {
					child.Meta = meta
				}
			}
		}

		sortTree(w.Tree)

		return nil
	}

	parent := findTreeNode(w.Tree, dir)
	if parent == nil || !parent.Dir {
		return fmt.Errorf("no folder %s", dir)
	}

	if findChild(parent.Children, meta.Name) == nil {
		parent.Children = append(parent.Children, &fileNode{Path: clean, Name: meta.Name, Loaded: true, Meta: meta})
	} else {
		for _, child := range parent.Children {
			if child.Name == meta.Name {
				child.Meta = meta
			}
		}
	}

	sortTree(parent.Children)

	return nil
}

// redrawTreeMenu repaints after an explorer menu action. Tab actions reset
// the scroll; explorer copy and paste keep it.
func (a *App) redrawTreeMenu(ctx context.Context) error {
	if a.page == nil {
		return nil
	}

	return a.page.Redraw(ctx)
}

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

// PendingCopy stages files for a later Paste. SrcPaths holds every staged
// workspace path; SrcPath is the first one for single-file callers. A cut
// clears after one paste; a copy stays staged for repeated pastes.
type PendingCopy struct {
	SrcPath  string
	SrcPaths []string
	IsCut    bool
}

// pasteStagedCopy is the Task 3 name for a staged file copy. It aliases
// PendingCopy so the App seam and this package share one shape.
type pasteStagedCopy = PendingCopy

// pendingCopy holds the files Copy staged. Copy keeps them for repeated
// pastes; a cut clears after one paste.
var pendingCopy *PendingCopy //nolint:gochecknoglobals // reason: single staged Ctrl+V copy shared by the clipboard staging helpers across explorer_clipboard.go and explorer_paste.go, mirrored onto App.pasteStaged

// stageCopy records one source for a later Paste.
func stageCopy(src string, isCut bool) {
	stageCopies([]string{src}, isCut)
}

// stageCopies records sources for a later Paste.
func stageCopies(srcs []string, isCut bool) {
	kept := make([]string, 0, len(srcs))
	for _, src := range srcs {
		if src != "" {
			kept = append(kept, src)
		}
	}

	first := ""
	if len(kept) > 0 {
		first = kept[0]
	}

	pendingCopy = &PendingCopy{SrcPath: first, SrcPaths: kept, IsCut: isCut}
}

// stagedCopy returns the file Copy staged, or nil when nothing is staged.
func stagedCopy() *PendingCopy {
	return pendingCopy
}

// clearStagedCopy drops the staged file, for example between tests.
func clearStagedCopy() {
	pendingCopy = nil
}

// stageExplorerCopy stages one workspace path for a later Paste.
func (a *App) stageExplorerCopy(src string, isCut bool) {
	a.stageExplorerCopies([]string{src}, isCut)
}

// stageExplorerCopies stages workspace paths for a later Paste. A cut
// clears after one paste; a copy stays staged for repeated pastes.
func (a *App) stageExplorerCopies(srcs []string, isCut bool) {
	stageCopies(srcs, isCut)
	a.pasteStaged = pendingCopy
}

// dropExplorerCopy clears the staged file on the app and in the store.
func (a *App) dropExplorerCopy() {
	clearStagedCopy()
	a.pasteStaged = nil
}

// shortcutHint labels a menu row with its keyboard shortcut, using the
// macOS symbol there and Ctrl elsewhere.
func shortcutHint(win string) string {
	if runtime.GOOS == "darwin" {
		switch win {
		case "Ctrl+X":
			return "\u2318X"
		case "Ctrl+C":
			return "\u2318C"
		case "Ctrl+V":
			return "\u2318V"
		}
	}

	return win
}

// contextMenuForExplorer returns the explorer rows for a tree id. The second
// result reports whether the id belongs to the explorer, including empty
// space and stale paths that yield no rows. Only visible rows have boxes, so
// callers reach this through the id of the row under the pointer.
// Right-clicking a row outside the multi-select replaces the selection with
// that row, so Copy and Cut act on what the menu shows.
func (a *App) contextMenuForExplorer(id string) ([]contextMenuRow, bool) {
	switch id {
	case "tree-pane", "sidebar", "":
		return []contextMenuRow{
			{ID: contextMenuPasteID, Label: "Paste", Hint: shortcutHint("Ctrl+V"), Enabled: stagedCopy() != nil},
		}, true
	}

	if !strings.HasPrefix(id, "tree-") {
		return nil, false
	}

	rel := strings.TrimPrefix(id, "tree-")

	switch rel {
	case "scroll", "thumb", "up", "down", "pane", "sticky":
		return nil, true
	}

	if _, ok := a.lookupTreePath(rel); !ok {
		return nil, true
	}

	if len(a.explSel) == 0 || !a.explSel[rel] {
		a.explSel = map[string]bool{rel: true}
		a.explAnchor = rel
	}

	pasteEnabled := stagedCopy() != nil

	return []contextMenuRow{
		{ID: contextMenuCutID + rel, Label: "Cut", Hint: shortcutHint("Ctrl+X"), Enabled: true},
		{ID: contextMenuCopyID + rel, Label: "Copy", Hint: shortcutHint("Ctrl+C"), Enabled: true},
		{ID: contextMenuCopyPathID + rel, Label: "Copy Path", Enabled: true},
		{ID: contextMenuCopyAbsID + rel, Label: "Copy Absolute Path", Enabled: true},
		{ID: contextMenuPasteID + rel, Label: "Paste", Hint: shortcutHint("Ctrl+V"), Enabled: pasteEnabled},
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

// copyTargets resolves what a Copy or Cut acts on: the multi-select when
// the row belongs to it, else the row alone.
func (a *App) copyTargets(rel string) []string {
	if len(a.explSel) > 0 && a.explSel[rel] {
		out := make([]string, 0, len(a.explSel))
		for path := range a.explSel {
			out = append(out, path)
		}

		sort.Strings(out)

		return out
	}

	return []string{rel}
}

// stageTreeCopy validates paths and stages them for Paste, writing the paths
// to the OS clipboard as text. Folders stage too: paste copies the whole
// tree. It reports how many files and folders staged.
func (a *App) stageTreeCopy(paths []string, isCut bool) {
	kept := make([]string, 0, len(paths))
	for _, path := range paths {
		if _, ok := a.lookupTreePath(path); !ok {
			continue
		}

		kept = append(kept, path)
	}

	if len(kept) == 0 {
		return
	}

	a.stageExplorerCopies(kept, isCut)
	ownframe.Write(strings.Join(kept, "\n"))

	verb := "copied"
	if isCut {
		verb = "cut"
	}

	if len(kept) == 1 {
		a.showToast("success", "File "+verb, kept[0])
	} else {
		a.showToast("success", fmt.Sprintf("%d files %s", len(kept), verb), strings.Join(kept, ", "))
	}

	a.refresh()
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

// explorerAbsPath resolves a tree path to a native absolute host path. A WSL
// checkout already roots under a Linux path such as /mnt/c or the guest
// filesystem, so the join is the WSL path. A Windows checkout rooted at a
// \\wsl$\ network share keeps the UNC form, which is the path Windows apps
// need to reach the same files. Cross-convention roots translate: a Windows
// drive path seen from WSL maps to /mnt/<drive>, and a /mnt/<drive> path
// seen from Windows maps back to the drive letter.
func (a *App) explorerAbsPath(rel string) string {
	abs := rel
	if !filepath.IsAbs(rel) {
		root := a.ws.Root
		if root == "" {
			root = "."
		}

		abs = filepath.Join(root, filepath.FromSlash(rel))
	}

	if isWSL() {
		if win, ok := windowsToWSLPath(abs); ok {
			return win
		}
	} else if runtime.GOOS == "windows" {
		if win, ok := wslMntToWindows(abs); ok {
			return win
		}
	}

	if out, err := filepath.Abs(abs); err == nil {
		return out
	}

	return abs
}

// copyTreeAbsPath writes the machine-native absolute path of one tree row
// to the clipboard.
func (a *App) copyTreeAbsPath(path string) {
	if _, ok := a.lookupTreePath(path); !ok {
		return
	}

	abs := a.explorerAbsPath(path)
	ownframe.Write(abs)
	a.showToast("success", "Absolute path copied", abs)
	a.refresh()
}

// toggleExplorerSel flips one row in the multi-select and anchors the range
// there. Ctrl+click never expands or opens the row.
func (a *App) toggleExplorerSel(path string) {
	if a.explSel == nil {
		a.explSel = map[string]bool{}
	}

	if a.explSel[path] {
		delete(a.explSel, path)
	} else {
		a.explSel[path] = true
	}

	a.explAnchor = path
}

// extendExplorerSel replaces the multi-select with the visible range between
// the anchor and the clicked row, like Shift+click in VS Code. A missing
// anchor selects just the row.
func (a *App) extendExplorerSel(path string) {
	order := a.explorerOrder()

	anchorIdx, pathIdx := -1, -1

	for i, p := range order {
		if p == a.explAnchor {
			anchorIdx = i
		}

		if p == path {
			pathIdx = i
		}
	}

	if pathIdx < 0 {
		return
	}

	if anchorIdx < 0 {
		a.explSel = map[string]bool{path: true}
		a.explAnchor = path

		return
	}

	lo, hi := anchorIdx, pathIdx
	if lo > hi {
		lo, hi = hi, lo
	}

	sel := make(map[string]bool, hi-lo+1)
	for _, p := range order[lo : hi+1] {
		sel[p] = true
	}

	a.explSel = sel
}

// clearExplorerSel drops the multi-select and its anchor.
func (a *App) clearExplorerSel() {
	a.explSel = nil
	a.explAnchor = ""
}

// explorerOrder lists the visible tree paths in display order for
// Shift+click ranges. Filtered mode lists the visible window; otherwise the
// full open-folders walk runs, which also warms lazy folders like a render.
func (a *App) explorerOrder() []string {
	if a.ws == nil {
		return nil
	}

	if a.filter != "" {
		out := make([]string, 0, len(a.view.Tree))
		for _, row := range a.view.Tree {
			out = append(out, row.Path)
		}

		return out
	}

	total := countTreeOpen(a.ws.Tree, a.folders)

	var rows []treeRow

	index := 0

	a.walkTree(a.ws.Tree, 0, false, &rows, &index, 0, total+1)

	out := make([]string, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.Path)
	}

	return out
}

// explorerCopyReady reports whether Ctrl+C/Ctrl+X should stage explorer
// files: no overlay or text field owns the keys, and the focus is a tree row
// or rows are already selected.
func (a *App) explorerCopyReady() bool {
	if !a.pasteTargetReady() {
		return false
	}

	if len(a.explSel) > 0 {
		return true
	}

	return isTreeRowID(a.page.Form().FocusID())
}

// copyExplorerSelection stages the multi-select, or the focused tree row
// when nothing is selected. It reports false when the keys belong to
// another control.
func (a *App) copyExplorerSelection(ctx context.Context, isCut bool) (bool, error) {
	var targets []string

	if len(a.explSel) > 0 {
		for path := range a.explSel {
			targets = append(targets, path)
		}

		sort.Strings(targets)
	} else if focus := a.page.Form().FocusID(); isTreeRowID(focus) {
		targets = []string{strings.TrimPrefix(focus, "tree-")}
	} else {
		return false, nil
	}

	a.stageTreeCopy(targets, isCut)

	return true, a.page.Redraw(ctx)
}

// clickTreeSelect handles Ctrl+click (toggle) and Shift+click (extend) on
// an explorer row without opening or expanding it.
func (a *App) clickTreeSelect(ctx context.Context, path string) error {
	if _, ok := a.lookupTreePath(path); !ok {
		return nil
	}

	if a.shift {
		a.extendExplorerSel(path)
	} else {
		a.toggleExplorerSel(path)
	}

	a.refresh()

	if err := a.page.Redraw(ctx); err != nil {
		return err
	}

	return a.page.Form().Focus(ctx, "tree-"+path)
}

// pasteTreeCopy pastes the staged files into destDir, or the workspace
// root when destDir is empty. Pasting onto a file id lands beside it. A cut
// moves every source; a copy leaves staging in place for repeated pastes.
func (a *App) pasteTreeCopy(ctx context.Context, destDir string) error {
	staged := stagedCopy()
	if staged == nil || (staged.SrcPath == "" && len(staged.SrcPaths) == 0) {
		a.showToast("info", "Nothing to paste", "Copy a file first.")
		a.refresh()

		return a.redrawTreeMenu(ctx)
	}

	if a.ws.Sample {
		a.showToast("error", "Cannot paste", sampleReadOnlyMessage)
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

	rels, err := a.PastePendingCopies(dest)
	if err != nil {
		a.showToast("error", "Cannot paste", err.Error())
		a.refresh()

		return a.redrawTreeMenu(ctx)
	}

	if dest != "" {
		a.folders[dest] = true
	}

	if len(rels) == 1 {
		a.showToast("success", "File pasted", rels[0])
	} else {
		a.showToast("success", fmt.Sprintf("%d files pasted", len(rels)), strings.Join(rels, ", "))
	}

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

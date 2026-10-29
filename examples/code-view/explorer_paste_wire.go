package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/chinmay-sawant/ownframe"
)

const (
	// pasteTextName is the fallback name for plain clipboard text.
	pasteTextName = "paste.txt"
	// pasteBinName is the fallback name when the clipboard text is not
	// valid UTF-8, such as image data pasted as text.
	pasteBinName = "paste.bin"
	// pasteMaxBytes caps one pasted host file so a huge drop cannot exhaust
	// memory. PNG, JPEG, and PDF copies fit comfortably under it.
	pasteMaxBytes = 64 << 20
)

// pasteTargetReady reports whether explorer paste may run: no overlay owns
// the keyboard and no text field owns the focus. The palette, search, find,
// filter, and browser inputs keep their keystrokes instead of pasting files.
func (a *App) pasteTargetReady() bool {
	if a.paletteOpen || a.branchOpen || a.searchOpen || a.settingsOpen || a.fileMenu || a.browseOpen {
		return false
	}

	switch a.page.Form().FocusID() {
	case "palette", "workspace-search", "find", "filter", "branch-filter", "project-path", "drawer-path":
		return false
	}

	return true
}

// pasteTargetDir resolves where Ctrl+V lands: the focused tree-<path> row,
// else the active file's folder, else the root. It validates through
// resolvePasteCandidate, the same check PastePendingCopy applies.
func (a *App) pasteTargetDir() string {
	ctxPath := ""
	if focus := a.page.Form().FocusID(); strings.HasPrefix(focus, "tree-") {
		ctxPath = strings.TrimPrefix(focus, "tree-")
	} else if doc := a.activeDoc(); doc != nil {
		ctxPath = doc.Path
	}

	if dir, ok := a.resolvePasteCandidate(ctxPath); ok {
		return dir
	}

	return ""
}

// writePastedBytes stores data as a new workspace file through the explorer
// paste backend, so the explorer lists it without a rescan.
func (a *App) writePastedBytes(dir, name string, data []byte) (string, error) {
	rel, err := uniquePasteCopyPath(a.ws, dir, name)
	if err != nil {
		return "", err
	}

	if err := a.ws.createFile(rel, data); err != nil {
		return "", err
	}

	if _, err := a.ws.ensureFile(rel); err != nil {
		return "", fmt.Errorf("paste %s: %w", rel, err)
	}

	a.refreshPastedFile(dir, rel)
	sortPasteFiles(a.ws)

	if dir != "" {
		a.folders[dir] = true
	}

	return rel, nil
}

// copyHostFileIntoWorkspace copies one existing host file into the paste
// target through PastePendingCopy: the file is staged transiently, pasted,
// and unstaged, so a later clipboard change cannot resurrect it. PNG, JPEG,
// and PDF bytes copy like text: the copy is a byte copy, so binary files
// land intact even though the viewer reads text only.
func (a *App) copyHostFileIntoWorkspace(src, dir string) (string, error) {
	info, err := os.Stat(src)
	if err != nil {
		return "", err
	}

	if info.IsDir() {
		return "", fmt.Errorf("paste %s: directories cannot be pasted", src)
	}

	if info.Size() > pasteMaxBytes {
		return "", fmt.Errorf("paste %s: file is larger than %s", src, humanSize(pasteMaxBytes))
	}

	SetPendingCopy(src, false)
	defer ClearPendingCopy()

	return a.PastePendingCopy(dir)
}

// pasteNameForText picks paste.txt for text and paste.bin for bytes that are
// not valid UTF-8, such as image data pasted as text.
func pasteNameForText(text string) string {
	if !utf8.ValidString(text) || strings.IndexByte(text, 0) >= 0 {
		return pasteBinName
	}

	return pasteTextName
}

// pasteFromClipboard pastes the OS clipboard into the explorer folder. The
// editing branch in onKeyDown returns before this runs, so plain Ctrl+V here
// is always the explorer path; Ctrl+Shift+V preview keeps its own branch.
//
// Serve (-web) exposes no clipboard route, so under -web this chord runs
// only when the shell forwards keys: explorer paste otherwise arrives from
// the Paste menu row, which reads the in-process clipboard like the edit
// menu does.
func (a *App) pasteFromClipboard(ctx context.Context) error {
	if a.editing || !a.pasteTargetReady() {
		return nil
	}

	text, err := ownframe.Read()
	if err != nil {
		a.showToast("info", "Clipboard is empty", "Copy a file, then press Ctrl+V in the explorer.")
		a.refresh()

		return a.page.Redraw(ctx)
	}

	return a.pasteClipboardText(ctx, text)
}

// pasteClipboardText pastes clipboard text into the explorer: a staged
// in-app copy first (the Task 3 pending copy), then the first existing file
// ParseClipboardFileList finds, then the raw text as a new file. It
// refreshes and redraws, reporting the new path in a toast.
func (a *App) pasteClipboardText(ctx context.Context, text string) error {
	if a.ws.Sample {
		a.showToast("info", "Sample workspace is read-only", "Open a disk folder with -root to paste files.")
		a.refresh()

		return a.page.Redraw(ctx)
	}

	if !a.pasteTargetReady() {
		return nil
	}

	dir := a.pasteTargetDir()

	if _, ok := GetPendingCopy(); ok {
		rels, err := a.PastePendingCopies(dir)
		if err != nil {
			a.showToast("error", "Paste failed", err.Error())
		} else if len(rels) == 1 {
			a.showToast("success", "File pasted", rels[0])
		} else {
			a.showToast("success", fmt.Sprintf("%d files pasted", len(rels)), strings.Join(rels, ", "))
		}

		a.refresh()

		return a.page.Redraw(ctx)
	}

	if sources := ParseClipboardFileList(text); len(sources) > 0 {
		rel, err := a.copyHostFileIntoWorkspace(sources[0], dir)
		if err != nil {
			a.showToast("error", "Paste failed", err.Error())
		} else {
			a.showToast("success", "Pasted", rel)
		}

		a.refresh()

		return a.page.Redraw(ctx)
	}

	if strings.TrimSpace(text) == "" {
		a.showToast("info", "Clipboard is empty", "Copy a file, then press Ctrl+V in the explorer.")
		a.refresh()

		return a.page.Redraw(ctx)
	}

	rel, err := a.writePastedBytes(dir, pasteNameForText(text), []byte(text))
	if err != nil {
		a.showToast("error", "Paste failed", err.Error())
	} else {
		a.showToast("success", "Pasted", rel)
	}

	a.refresh()

	return a.page.Redraw(ctx)
}

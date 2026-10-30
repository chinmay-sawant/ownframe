package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// Explorer paste backend: the Copy/Paste menu and Ctrl+V land here.
//
// The staged copy lives in explorer_clipboard.go: PendingCopy carries the
// source in SrcPath (the task wording calls it SrcAbs; a staged in-app copy
// holds a workspace-relative path while a clipboard file list stages an
// absolute host path) and SetPendingCopy, GetPendingCopy, and
// ClearPendingCopy below wrap that shared store, so the menu, Ctrl+V, and
// this backend see one staged file.
//
// OS clipboard gap, deliberately out of scope: internal/clipboard moves text
// only (X11 UTF8_STRING, Windows CF_UNICODETEXT, macOS
// NSPasteboardTypeString). There is no CF_HDROP, text/uri-list, or image
// backend, so file drops from other apps arrive as text when the source
// offers it, and ParseClipboardFileList parses that text. WSL translation
// reuses wsl_path.go, so no new internal/filepick import is needed and
// scripts/example-internal-allowlist.txt stays untouched.
//
// ParseClipboardFileList is the validated variant: the Ctrl+V text flow in
// explorer_paste_wire.go keeps its own raw parseClipboardFileList, which
// returns unfiltered candidate lines, while this one returns only existing
// files. The two stay separate until the flows converge.

// SetPendingCopy stages src for a later PastePendingCopy: a
// workspace-relative path from the Copy menu, or an absolute host path from
// an OS file list. A cut stays staged for one paste; a copy stays staged
// for repeated pastes.
func SetPendingCopy(src string, isCut bool) {
	stageCopy(src, isCut)
}

// GetPendingCopy returns a copy of the staged paste, if any. The copy keeps
// callers from mutating the shared staged state.
func GetPendingCopy() (*PendingCopy, bool) {
	staged := stagedCopy()
	if staged == nil {
		return nil, false
	}

	dup := *staged

	return &dup, true
}

// ClearPendingCopy drops any staged paste, for example between tests.
func ClearPendingCopy() {
	clearStagedCopy()
}

// ResolvePasteDir picks the workspace-relative directory a paste lands in.
// An explicit ctxPath wins (a directory stays, a file maps to its parent),
// then the focused tree-<path> row decides the same way, then the active
// document's directory. Empty means the workspace root. Anything outside the
// root (absolute paths, ".." escapes, unknown names) falls through to the
// next candidate, and "" comes back when nothing qualifies.
func (a *App) ResolvePasteDir(ctxPath string) string {
	if strings.TrimSpace(ctxPath) != "" {
		if dir, ok := a.resolvePasteCandidate(ctxPath); ok {
			return dir
		}
	}

	if a.page != nil {
		if id := a.page.Form().FocusID(); strings.HasPrefix(id, "tree-") {
			if path := strings.TrimSpace(strings.TrimPrefix(id, "tree-")); path != "" {
				if dir, ok := a.resolvePasteCandidate(path); ok {
					return dir
				}
			}
		}
	}

	if doc := strings.TrimSpace(a.activeDocPath()); doc != "" {
		if dir, ok := a.resolvePasteCandidate(dirName(doc)); ok {
			return dir
		}
	}

	return ""
}

// resolvePasteCandidate maps one workspace path to its paste directory:
// directories stay, files map to their parent, and "" is the root. ok is
// false when the candidate escapes the workspace or names nothing usable, so
// the caller falls through to the next candidate.
func (a *App) resolvePasteCandidate(path string) (string, bool) {
	clean, ok := cleanPasteCandidate(path)
	if !ok || a.ws == nil {
		return "", false
	}

	if clean == "" {
		return "", true
	}

	if isDir, found := a.lookupTreePath(clean); found {
		if isDir {
			return clean, true
		}

		return a.resolvePasteCandidate(dirName(clean))
	}

	// A lazy workspace only lists opened folders: accept a real directory on
	// disk even when the tree has not scanned it yet.
	if a.ws.Sample || a.ws.Root == "" {
		return "", false
	}

	if info, err := os.Stat(filepath.Join(a.ws.Root, filepath.FromSlash(clean))); err == nil && info.IsDir() {
		return clean, true
	}

	return "", false
}

// cleanPasteCandidate normalizes one paste target: "" is the root, anything
// absolute or escaping with ".." is rejected.
func cleanPasteCandidate(path string) (string, bool) {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return "", true
	}

	if filepath.IsAbs(trimmed) {
		return "", false
	}

	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(trimmed)))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", false
	}

	return clean, true
}

// PastePendingCopy copies the staged file into dstDir and registers it with
// the workspace, returning the new workspace-relative slash path. dstDir is
// a paste directory as ResolvePasteDir returns it ("" is the root, files map
// to their parent). The copy keeps its bytes verbatim: PNG, JPEG, PDF, and
// any other binary land intact with no decode. A missing absolute source
// pastes an empty file under the source name; any other read error aborts.
// A cut moves the source: after a successful copy the workspace-relative
// source is removed and the tree rescanned. Errors name the problem for a
// toast; nothing here touches the UI, so tests drive it directly.
func (a *App) PastePendingCopy(dstDir string) (string, error) {
	rels, err := a.PastePendingCopies(dstDir)
	if err != nil {
		return "", err
	}

	if len(rels) == 0 {
		return "", fmt.Errorf("nothing to paste: copy a file first")
	}

	return rels[0], nil
}

// PastePendingCopies pastes every staged source into dstDir and returns the
// new workspace-relative slash paths in staged order. Sources stage through
// stageCopies, so one call moves a whole multi-select.
func (a *App) PastePendingCopies(dstDir string) ([]string, error) {
	staged, ok := GetPendingCopy()
	if !ok || len(staged.SrcPaths) == 0 {
		if !ok || staged.SrcPath == "" {
			return nil, fmt.Errorf("nothing to paste: copy a file first")
		}

		staged.SrcPaths = []string{staged.SrcPath}
	}

	ws := a.ws
	if ws == nil {
		return nil, fmt.Errorf("nothing to paste into: no workspace")
	}

	if !ws.writable() {
		return nil, errSampleReadOnly()
	}

	dir, valid := a.resolvePasteCandidate(dstDir)
	if !valid {
		return nil, fmt.Errorf("paste target %q is not a folder in this workspace", dstDir)
	}

	var rels []string

	var cut []string

	for _, src := range staged.SrcPaths {
		rel, err := a.pasteOneSource(dir, src)
		if err != nil {
			if len(rels) == 0 {
				return nil, err
			}

			break
		}

		rels = append(rels, rel)

		if staged.IsCut && !filepath.IsAbs(src) {
			cut = append(cut, src)
		}
	}

	if len(cut) > 0 {
		a.removeCutSources(cut)
	}

	if staged.IsCut {
		a.dropExplorerCopy()
		a.clearExplorerSel()
	}

	return rels, nil
}

// pasteOneSource copies one staged source (file or folder) into dir and
// registers it, returning the new workspace-relative slash path. Folders
// copy recursively with os.CopyFS; files copy byte for byte.
func (a *App) pasteOneSource(dir, src string) (string, error) {
	ws := a.ws

	if isPasteSourceDir(ws, src) {
		rel, err := uniquePasteCopyPath(ws, dir, pasteBaseName(src))
		if err != nil {
			return "", err
		}

		if rel == "" {
			return "", fmt.Errorf("cannot paste %q: no folder name", src)
		}

		dst := filepath.Join(ws.Root, filepath.FromSlash(rel))
		from := src
		if !filepath.IsAbs(from) {
			from = filepath.Join(ws.Root, filepath.FromSlash(from))
		}

		if err := os.MkdirAll(dst, workspaceDirMode); err != nil {
			return "", fmt.Errorf("paste %s: %w", src, err)
		}

		if err := os.CopyFS(dst, os.DirFS(from)); err != nil {
			return "", fmt.Errorf("paste %s: %w", src, err)
		}

		// A folder brings its own tree: rescan rather than inserting nodes
		// by hand, the same path refreshPastedFile takes for full scans.
		if ws.Lazy {
			if dir != "" {
				_ = ws.loadDir(dir)
			}

			_ = ws.loadDir(rel)
		} else {
			ws.finish()
		}

		return rel, nil
	}

	data, err := a.readPasteSource(src)
	if err != nil {
		return "", err
	}

	rel, err := uniquePasteCopyPath(ws, dir, pasteBaseName(src))
	if err != nil {
		return "", err
	}

	if rel == "" {
		return "", fmt.Errorf("cannot paste %q: no file name", src)
	}

	if err := ws.createFile(rel, data); err != nil {
		return "", err
	}

	if _, err := ws.ensureFile(rel); err != nil {
		return "", fmt.Errorf("paste %s: %w", rel, err)
	}

	if ws.Lazy {
		if dir != "" {
			_ = ws.loadDir(dir)
		}

		// Best effort: the bytes and the ByPath entry above are the paste;
		// the node only decides whether the row shows without a rescan.
		_ = ws.addDiskFile(rel)

		sortPasteFiles(ws)
	} else if findTreeNode(ws.Tree, rel) == nil {
		ws.finish()
	}

	return rel, nil
}

// isPasteSourceDir reports whether a staged source is a folder on disk.
func isPasteSourceDir(ws *workspace, src string) bool {
	path := src
	if !filepath.IsAbs(path) {
		if ws == nil {
			return false
		}

		if node := findTreeNode(ws.Tree, src); node != nil {
			return node.Dir
		}

		path = filepath.Join(ws.Root, filepath.FromSlash(src))
	}

	info, err := os.Stat(path)

	return err == nil && info.IsDir()
}

// removeCutSources deletes moved workspace sources and rescans their
// parents. Absolute host sources are never cut, so only relative paths
// arrive here.
func (a *App) removeCutSources(srcs []string) {
	ws := a.ws
	if ws == nil {
		return
	}

	parents := map[string]bool{}

	for _, src := range srcs {
		full := filepath.Join(ws.Root, filepath.FromSlash(src))
		if info, err := os.Stat(full); err == nil && info.IsDir() {
			_ = os.RemoveAll(full)
		} else {
			_ = os.Remove(full)
		}

		parents[dirName(src)] = true
	}

	if ws.Lazy {
		for parent := range parents {
			if parent != "" {
				_ = ws.loadDir(parent)
			}
		}

		return
	}

	next, delta, err := rescanDiskWorkspace(ws.Root, ws, a.scanLimit)
	if err != nil {
		return
	}

	a.installRescan(next, delta)
}

// readPasteSource reads a staged copy source. Absolute paths come from the
// OS (clipboard file lists may name files outside the workspace) and a
// missing one yields empty bytes under the source name; any other read error
// aborts. Workspace-relative paths read through the workspace.
func (a *App) readPasteSource(src string) ([]byte, error) {
	if filepath.IsAbs(src) {
		raw, err := os.ReadFile(src) //nolint:gosec // src is the user's own staged copy source, confirmed at paste time
		if err != nil {
			if os.IsNotExist(err) {
				return nil, nil
			}

			return nil, fmt.Errorf("paste %s: %w", src, err)
		}

		return raw, nil
	}

	raw, err := a.ws.readFile(src)
	if err != nil {
		return nil, fmt.Errorf("paste %s: %w", src, err)
	}

	return raw, nil
}

// pasteBaseName takes the file name of a staged source. Workspace-relative
// and absolute POSIX paths split on "/", and a stray Windows separator is
// tolerated so a pasted C:\ path still names its file.
func pasteBaseName(src string) string {
	base := baseName(filepath.ToSlash(src))
	if i := strings.LastIndexByte(base, '\\'); i >= 0 {
		base = base[i+1:]
	}

	return base
}

const (
	// firstPasteCopySuffix names the first duplicate: "notes (copy).md".
	firstPasteCopySuffix = " (copy)"
	// pasteCopyStart numbers later duplicates from "notes (copy 2).md".
	pasteCopyStart = 2
	// maxPasteNameTries caps unique-name generation so a crowded folder
	// fails instead of looping.
	maxPasteNameTries = 1000
)

// pastePathExists reports whether a slash-relative path exists on disk.
func pastePathExists(root, rel string) bool {
	_, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel)))

	return err == nil
}

// uniquePasteCopyPath picks dir/base, then "stem (copy).ext", "stem (copy
// 2).ext", and so on. Both the workspace map and the disk count as taken, so
// a paste never overwrites an untracked file.
func uniquePasteCopyPath(ws *workspace, dir, base string) (string, error) {
	join := func(name string) string {
		if dir == "" {
			return name
		}

		return dir + "/" + name
	}

	if ws.ByPath[join(base)] == nil && !pastePathExists(ws.Root, join(base)) {
		return join(base), nil
	}

	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)

	if candidate := join(stem + firstPasteCopySuffix + ext); ws.ByPath[candidate] == nil && !pastePathExists(ws.Root, candidate) {
		return candidate, nil
	}

	for i := pasteCopyStart; i <= maxPasteNameTries; i++ {
		candidate := join(fmt.Sprintf("%s (copy %d)%s", stem, i, ext))
		if ws.ByPath[candidate] == nil && !pastePathExists(ws.Root, candidate) {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("no free name for %s in %q", base, dir)
}

// sortPasteFiles keeps Files ordered like every other scan path after a lazy
// paste appends out of order.
func sortPasteFiles(ws *workspace) {
	sort.Slice(ws.Files, func(i, j int) bool {
		return strings.ToLower(ws.Files[i].Path) < strings.ToLower(ws.Files[j].Path)
	})
}

const (
	// maxPasteSources caps parsed clipboard file lists so a huge drop cannot
	// exhaust memory statting paths.
	maxPasteSources = 128
)

// ParseClipboardFileList reads clipboard text as a file list: text/uri-list
// (file:// URIs with %-decoding, "#" comments skipped), Windows C:\ paths
// translated through windowsToWSLPath when isWSL, WSL \\wsl$\ paths, and
// plain newline-separated absolute paths. Only existing regular files come
// back, in first-seen order; relative paths, ".." traversal, and missing
// files are dropped. Extension policy is permissive on purpose: PNG, JPEG,
// and PDF paste as bytes like any other binary (no decode), so every
// existing file is accepted for now.
func ParseClipboardFileList(text string) []string {
	var out []string

	seen := map[string]bool{}

	for _, line := range strings.Split(text, "\n") {
		if len(out) >= maxPasteSources {
			break
		}

		path, ok := translateClipboardPath(line)
		if !ok || seen[path] {
			continue
		}

		seen[path] = true

		info, err := os.Stat(path)
		if err != nil || !info.Mode().IsRegular() {
			continue
		}

		out = append(out, path)
	}

	return out
}

// translateClipboardPath turns one clipboard line into an absolute local
// path. It reports false for blanks, text/uri-list comments, relative
// paths, ".." traversal, and anything that cannot map to this machine.
func translateClipboardPath(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return "", false
	}

	if path, ok := clipboardFileURIToPath(trimmed); ok {
		trimmed = path
	} else if strings.HasPrefix(trimmed, "file:") {
		return "", false
	}

	if len(trimmed) >= 2 && isWSLSep(trimmed[0]) && isWSLSep(trimmed[1]) {
		return wslClipboardPath(trimmed)
	}

	if isClipboardDrivePath(trimmed) {
		if isWSL() {
			translated, ok := windowsToWSLPath(trimmed)
			if !ok {
				return "", false
			}

			return cleanClipboardAbsPath(translated)
		}

		if runtime.GOOS == "windows" {
			return cleanClipboardAbsPath(filepath.Clean(trimmed))
		}

		return "", false
	}

	return cleanClipboardAbsPath(trimmed)
}

// clipboardFileURIToPath decodes a file:// URI to a local path. Only empty
// and localhost hosts map locally; a remote file server is not a paste
// source. A "/C:/..." drive form is returned without its leading slash so
// the drive handling below sees it.
func clipboardFileURIToPath(uri string) (string, bool) {
	const scheme = "file://"

	if len(uri) < len(scheme) || !strings.EqualFold(uri[:len(scheme)], scheme) {
		return "", false
	}

	rest := uri[len(scheme):]

	var path string

	if strings.HasPrefix(rest, "/") {
		path = rest
	} else {
		i := strings.IndexByte(rest, '/')
		if i < 0 {
			return "", false
		}

		if host := rest[:i]; host != "" && !strings.EqualFold(host, "localhost") {
			return "", false
		}

		path = rest[i:]
	}

	decoded, err := url.PathUnescape(path)
	if err != nil {
		return "", false
	}

	if len(decoded) >= drivePrefixLen && decoded[0] == '/' && isDriveLetter(decoded[1]) && decoded[2] == ':' {
		decoded = decoded[1:]
	}

	return decoded, true
}

// isClipboardDrivePath reports a Windows drive-letter path such as C:\x or
// C:/x, reusing the drivePrefixLen shape from wsl_path.go.
func isClipboardDrivePath(path string) bool {
	return len(path) >= drivePrefixLen && isDriveLetter(path[0]) && path[1] == ':' && isWSLSep(path[2])
}

// wslClipboardPath maps \\wsl$\<distro>\... to the local path. The distro
// must be this guest's (WSL_DISTRO_NAME) when known; a foreign distro is not
// a local file. An unknown guest keeps the path and lets the existence
// filter decide.
func wslClipboardPath(unc string) (string, bool) {
	distro, wslPath, ok := parseWSLRoot(unc)
	if !ok {
		return "", false
	}

	if own := os.Getenv("WSL_DISTRO_NAME"); own != "" && !strings.EqualFold(own, distro) {
		return "", false
	}

	return cleanClipboardAbsPath(wslPath)
}

// cleanClipboardAbsPath accepts only absolute paths without NUL bytes or
// ".." segments.
func cleanClipboardAbsPath(path string) (string, bool) {
	if path == "" || strings.ContainsRune(path, 0) {
		return "", false
	}

	for _, seg := range strings.Split(filepath.ToSlash(path), "/") {
		if seg == ".." {
			return "", false
		}
	}

	if !filepath.IsAbs(path) {
		return "", false
	}

	return filepath.Clean(path), true
}

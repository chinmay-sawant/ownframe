package main

import (
	"path/filepath"
	"sort"
	"strings"
)

// workspaceDelta is what a full rescan found: paths that appeared, disappeared,
// or kept their path but changed size or language. Same counts the files that
// did not move.
type workspaceDelta struct {
	Added   []string
	Removed []string
	Changed []string
	Same    int
}

// diffWorkspace compares two scans by path. A file counts as changed when its
// size or language moved; contents are not read, so the check stays cheap.
func diffWorkspace(old, next *workspace) workspaceDelta {
	var delta workspaceDelta

	oldFiles := workspaceFilesByPath(old)
	nextFiles := workspaceFilesByPath(next)

	for path, meta := range nextFiles {
		prev, ok := oldFiles[path]
		switch {
		case !ok:
			delta.Added = append(delta.Added, path)
		case prev.Size != meta.Size || prev.Lang != meta.Lang:
			delta.Changed = append(delta.Changed, path)
		default:
			delta.Same++
		}
	}

	for path := range oldFiles {
		if _, ok := nextFiles[path]; !ok {
			delta.Removed = append(delta.Removed, path)
		}
	}

	sort.Strings(delta.Added)
	sort.Strings(delta.Removed)
	sort.Strings(delta.Changed)

	return delta
}

// workspaceFilesByPath returns the scan's file map, rebuilding it from Files
// when a caller handed over a workspace without one. A nil workspace is empty.
func workspaceFilesByPath(w *workspace) map[string]*fileMeta {
	files := map[string]*fileMeta{}

	if w == nil {
		return files
	}

	if len(w.ByPath) > 0 {
		for path, meta := range w.ByPath {
			if meta != nil {
				files[path] = meta
			}
		}

		return files
	}

	for _, meta := range w.Files {
		if meta != nil {
			files[meta.Path] = meta
		}
	}

	return files
}

// rescanDiskWorkspace scans root again and reports how it moved from old. A
// limit of zero or less means maxWorkspaceFiles.
func rescanDiskWorkspace(root string, old *workspace, limit int) (*workspace, workspaceDelta, error) {
	next, err := loadDiskWorkspaceWithLimit(root, limit)
	if err != nil {
		return nil, workspaceDelta{}, err
	}
	if old != nil && old.Lazy {
		next.Lazy = true
		next.Truncated = false
		next.TruncatedNote = ""
	}

	return next, diffWorkspace(old, next), nil
}

// isGitMetadataChange reports whether path is a git ref, index, or log file
// that can move the branch name or working-tree diff. Object churn under
// .git/objects is not a branch change and stays out.
func isGitMetadataChange(path string) bool {
	slash := "/" + strings.TrimPrefix(filepath.ToSlash(path), "/")

	const marker = "/.git/"

	i := strings.Index(slash, marker)
	if i < 0 {
		return false
	}

	rest := slash[i+len(marker):]

	switch {
	case rest == "HEAD", rest == "packed-refs", rest == "index":
		return true
	case strings.HasPrefix(rest, "refs/"), strings.HasPrefix(rest, "logs/"):
		return true
	}

	return false
}

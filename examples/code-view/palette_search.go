package main

import (
	"bufio"
	"bytes"
	"context"
	"io/fs"
	"os/exec"
	"path/filepath"
	"strings"
)

// diskFileSearcher is an optional hook for testing disk file search.
var diskFileSearcher func(ctx context.Context, root string) ([]string, error) //nolint:gochecknoglobals // test hook for the disk fallback, nil in production

// diskSearchFiles runs a fast disk search across root using ripgrep or git ls-files.
func diskSearchFiles(ctx context.Context, root string) ([]string, error) {
	if diskFileSearcher != nil {
		return diskFileSearcher(ctx, root)
	}

	return defaultDiskSearchFiles(ctx, root)
}

// defaultDiskSearchFiles tries ripgrep, then git ls-files, then a directory walk.
func defaultDiskSearchFiles(ctx context.Context, root string) ([]string, error) {
	if root == "" || root == "sample" {
		return nil, nil
	}

	// 1. Try ripgrep if available.
	if rgPath, err := exec.LookPath("rg"); err == nil {
		cmd := exec.CommandContext(ctx, rgPath, "--files") //nolint:gosec // rg path from LookPath with a fixed --files flag, Dir pinned to the scanned root
		cmd.Dir = root
		out, err := cmd.Output()
		if err == nil {
			return parseFileLines(out), nil
		}
	}

	// 2. Fall back to git ls-files.
	if gitPath, err := exec.LookPath("git"); err == nil {
		cmd := exec.CommandContext(ctx, gitPath, "ls-files", "--cached", "--others", "--exclude-standard") //nolint:gosec // git path from LookPath with fixed ls-files flags, Dir pinned to the scanned root
		cmd.Dir = root
		out, err := cmd.Output()
		if err == nil {
			return parseFileLines(out), nil
		}
	}

	// 3. Fall back to bounded directory walk if neither command succeeded.
	var paths []string
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || ctx.Err() != nil {
			return fs.SkipAll
		}
		if path == root {
			return nil
		}
		if d.IsDir() {
			if skipDir(d.Name()) {
				return fs.SkipDir
			}
			return nil
		}
		if skipExt(d.Name()) {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err == nil {
			paths = append(paths, filepath.ToSlash(rel))
		}
		if len(paths) >= maxWorkspaceFiles {
			return fs.SkipAll
		}
		return nil
	})

	return paths, nil
}

// parseFileLines parses newline-delimited file paths from command output.
func parseFileLines(out []byte) []string {
	var paths []string
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			paths = append(paths, filepath.ToSlash(line))
		}
	}
	return paths
}

// makePaletteRow builds one suggestion from a path.
func makePaletteRow(path string) paletteRow {
	dir := ""
	name := path
	if idx := strings.LastIndexByte(path, '/'); idx >= 0 {
		dir = path[:idx]
		name = path[idx+1:]
	}
	lang := languageOf(path)
	return paletteRow{
		Path:  path,
		Name:  name,
		Dir:   dir,
		Lang:  lang,
		Badge: languageInfo(lang).Badge,
	}
}

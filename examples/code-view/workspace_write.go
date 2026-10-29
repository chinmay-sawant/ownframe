package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// workspaceFileMode is the mode for a saved file.
	workspaceFileMode = 0o644
	// workspaceDirMode creates a missing parent directory for a saved file.
	workspaceDirMode = 0o750
	// sampleReadOnlyMessage is the error text for a write to the embedded
	// demo workspace.
	sampleReadOnlyMessage = "the sample workspace is read-only"
)

// errSampleReadOnly builds the error for a write to the embedded sample. It
// returns a fresh value so the package keeps no global error sentinel.
func errSampleReadOnly() error {
	return errors.New(sampleReadOnlyMessage)
}

// writable reports whether the workspace lives on disk.
func (w *workspace) writable() bool {
	return !w.Sample
}

// createFile writes a brand-new file under the root. It refuses the sample
// workspace, paths outside the root, and paths the workspace or the disk
// already holds. Registration (ByPath, tree) stays with the caller: paste
// calls ensureFile and rescans the target dir, so lazy trees stay
// incremental.
func (w *workspace) createFile(path string, data []byte) error {
	if !w.writable() {
		return errSampleReadOnly()
	}

	clean := filepath.Clean(filepath.FromSlash(path))
	if clean == "." || filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return fmt.Errorf("path escapes workspace: %s", path)
	}

	rel := filepath.ToSlash(clean)
	if w.ByPath[rel] != nil {
		return fmt.Errorf("file exists %s", path)
	}

	target := filepath.Join(w.Root, clean)
	if _, err := os.Stat(target); err == nil {
		return fmt.Errorf("file exists %s", path)
	}

	if err := os.MkdirAll(filepath.Dir(target), workspaceDirMode); err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}

	if err := os.WriteFile(target, data, workspaceFileMode); err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}

	return nil
}

// writeFile saves one scanned file under the root and refreshes its meta. It
// refuses the sample, paths outside the root, and files the scan never
// listed. It does not stage or commit anything in git.
func (w *workspace) writeFile(path string, data []byte) error {
	if !w.writable() {
		return errSampleReadOnly()
	}

	clean := filepath.Clean(filepath.FromSlash(path))
	if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return fmt.Errorf("path escapes workspace: %s", path)
	}

	meta := w.ByPath[filepath.ToSlash(clean)]
	if meta == nil {
		return fmt.Errorf("no file %s", path)
	}

	target := filepath.Join(w.Root, clean)
	if err := os.MkdirAll(filepath.Dir(target), workspaceDirMode); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	if err := os.WriteFile(target, data, workspaceFileMode); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	size := int64(len(data))
	w.Bytes += size - meta.Size
	meta.Size = size
	meta.Big = size > maxFileBytes

	return nil
}

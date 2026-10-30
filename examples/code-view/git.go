package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	gitTimeout = 3 * time.Second
	// untrackedLimit is the largest untracked file the diff view synthesizes.
	untrackedLimit = 256 << 10
)

// loadLiveChanges reads the working tree diff for -root. It keeps the
// combined HEAD diff in Files and splits Staged (`git diff --cached`) from
// Unstaged (`git diff` plus untracked files). Untracked files are added as
// synthetic all-added hunks. Writes live in scm_ops.go.
func loadLiveChanges(ctx context.Context, root string) (changeSet, error) {
	ctx, cancel := context.WithTimeout(ctx, gitTimeout)
	defer cancel()

	branch, _ := gitText(ctx, root, "rev-parse", "--abbrev-ref", "HEAD")
	cs := changeSet{Branch: branch}
	if cs.Branch == "" {
		cs.Branch = "detached"
	}

	raw, err := gitBytes(ctx, root, "diff", "--no-color", "--no-ext-diff", "HEAD")
	if err != nil {
		// A repository without a first commit has no HEAD. Show the two sides
		// git can still produce instead of failing the whole panel.
		work, workErr := gitBytes(ctx, root, "diff", "--no-color", "--no-ext-diff")
		cached, cachedErr := gitBytes(ctx, root, "diff", "--cached", "--no-color", "--no-ext-diff")
		if workErr != nil && cachedErr != nil {
			return cs, err
		}

		cs.Staged = parseDiff(string(cached))
		cs.Unstaged = parseDiff(string(work))
	} else {
		cached, cachedErr := gitBytes(ctx, root, "diff", "--cached", "--no-color", "--no-ext-diff")
		if cachedErr == nil {
			cs.Staged = parseDiff(string(cached))
		}

		work, workErr := gitBytes(ctx, root, "diff", "--no-color", "--no-ext-diff")
		if workErr == nil {
			cs.Unstaged = parseDiff(string(work))
		} else {
			cs.Unstaged = parseDiff(string(raw))
			cs.Staged = nil
		}

		cs.Files = parseDiff(string(raw))
	}

	untracked, err := gitBytes(ctx, root, "ls-files", "--others", "--exclude-standard", "-z")
	if err == nil {
		for _, path := range strings.Split(string(untracked), "\x00") {
			if path == "" {
				continue
			}

			if file, ok := untrackedChange(root, path); ok {
				cs.Files = append(cs.Files, file)
				cs.Unstaged = append(cs.Unstaged, file)
			}
		}
	}

	sortChangeFiles(cs.Files)
	sortChangeFiles(cs.Staged)
	sortChangeFiles(cs.Unstaged)

	if len(cs.Files) == 0 {
		cs.Note = "The working tree is clean. Make a change and refresh."
	}

	return cs, nil
}

// sortChangeFiles orders files case-insensitively by path.
func sortChangeFiles(files []changedFile) {
	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i].Path) < strings.ToLower(files[j].Path)
	})
}

// untrackedChange turns one untracked file into an all-added change.
func untrackedChange(root, path string) (changedFile, bool) {
	full := filepath.Join(root, filepath.FromSlash(path))

	info, err := os.Stat(full)
	if err != nil || info.Size() > untrackedLimit {
		return changedFile{}, false
	}

	raw, err := os.ReadFile(full) //nolint:gosec // path came from git ls-files under the scanned root
	if err != nil {
		return changedFile{}, false
	}

	file := changedFile{Path: path, Status: "U"}

	if isBinary(raw) {
		file.Binary = true

		return file, true
	}

	text := strings.ReplaceAll(string(raw), "\r\n", "\n")
	text = strings.TrimSuffix(text, "\n")
	lines := strings.Split(text, "\n")

	h := hunk{NewStart: 1, NewCount: len(lines), oldNext: 1, newNext: 1}
	h.OldStart = 0

	for i, line := range lines {
		h.Lines = append(h.Lines, diffLine{Kind: '+', NewNo: i + 1, Text: expandTabs(line)})
	}

	file.Add = len(lines)
	file.Hunks = []hunk{h}
	measureDiff(&file)

	return file, true
}

// gitText runs one git command and returns its trimmed stdout.
func gitText(ctx context.Context, root string, args ...string) (string, error) {
	out, err := gitBytes(ctx, root, args...)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

// cappedBuffer keeps at most limit bytes and drains the rest without an EPIPE.
type cappedBuffer struct {
	buf       []byte
	limit     int
	truncated bool
}

func (b *cappedBuffer) Write(p []byte) (int, error) {
	if room := b.limit - len(b.buf); room > 0 {
		if len(p) <= room {
			b.buf = append(b.buf, p...)
		} else {
			b.buf = append(b.buf, p[:room]...)
			b.truncated = true
		}
	} else if len(p) > 0 {
		b.truncated = true
	}

	return len(p), nil
}

// gitBytesCapped runs one git command and caps its stdout.
func gitBytesCapped(ctx context.Context, root string, limit int, args ...string) ([]byte, bool, error) {
	full := append([]string{"-C", root}, args...)
	cmd := exec.CommandContext(ctx, "git", full...) //nolint:gosec // fixed binary, fixed subcommand, operator-selected -C
	cmd.Env = append(os.Environ(), "GIT_PAGER=cat", "LC_ALL=C", "GIT_TERMINAL_PROMPT=0", "GIT_OPTIONAL_LOCKS=0")
	hideConsoleWindow(cmd)

	buf := &cappedBuffer{limit: limit}
	cmd.Stdout = buf

	if err := cmd.Run(); err != nil {
		return nil, false, fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}

	return buf.buf, buf.truncated, nil
}

// gitBytes runs one git command. The arguments are fixed by the caller and
// -C carries the operator-selected workspace root.
func gitBytes(ctx context.Context, root string, args ...string) ([]byte, error) {
	full := append([]string{"-C", root}, args...)
	cmd := exec.CommandContext(ctx, "git", full...) //nolint:gosec // fixed binary, fixed subcommand, operator-selected -C
	cmd.Env = append(os.Environ(), "GIT_PAGER=cat", "LC_ALL=C", "GIT_TERMINAL_PROMPT=0", "GIT_OPTIONAL_LOCKS=0")
	hideConsoleWindow(cmd)

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}

	return out, nil
}

// gitBytesStdin runs one git command with input on stdin. It is used for
// check-ignore batches, where one process answers for every scanned directory.
func gitBytesStdin(ctx context.Context, root, input string, args ...string) ([]byte, error) {
	full := append([]string{"-C", root}, args...)
	cmd := exec.CommandContext(ctx, "git", full...) //nolint:gosec // fixed binary, fixed subcommand, operator-selected -C
	cmd.Env = append(os.Environ(), "GIT_PAGER=cat", "LC_ALL=C", "GIT_TERMINAL_PROMPT=0", "GIT_OPTIONAL_LOCKS=0")
	cmd.Stdin = strings.NewReader(input)
	hideConsoleWindow(cmd)

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}

	return out, nil
}

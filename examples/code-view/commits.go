package main

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	maxCommits          = 20
	maxHistoryCommits   = 500
	maxCommitFiles      = 20
	maxCommitBody       = 2000
	maxCommitPatchCache = 24
	maxCommitPatchBytes = 512 << 10
	commitHashLen       = 40
	commitFields        = 8
	commitDateLen       = 10
	commitFileSeeds     = 8
	commitParts         = 3
	decimalBase         = 10
)

// commitLog is the recent history of the checked-out branch.
type commitLog struct {
	Branch      string
	Ref         string
	Head        string
	Commits     []gitCommit
	Note        string
	Unborn      bool
	Unavailable bool
	Sample      bool
}

// gitCommit is one commit from git log, with its file stats already parsed.
type gitCommit struct {
	Hash    string
	Short   string
	Parents []string
	Subject string
	Body    string
	Author  string
	Date    string
	Files   []commitFile
	Add     int
	Del     int
	More    int
}

// commitFile is one file changed by a commit.
type commitFile struct {
	Path    string
	OldPath string
	Status  string
	Add     int
	Del     int
	Binary  bool
}

// loadCommitLog reads the recent commits on ref in one git process. An empty
// ref means HEAD and follows the checked-out branch.
func loadCommitLog(ctx context.Context, root, ref string) (commitLog, error) {
	return loadCommitLogLimit(ctx, root, ref, maxCommits)
}

// loadCommitLogLimit reads up to limit recent commits on ref in one git
// process. A non-positive limit falls back to maxCommits.
func loadCommitLogLimit(ctx context.Context, root, ref string, limit int) (commitLog, error) {
	if limit < 1 {
		limit = maxCommits
	}

	ctx, cancel := context.WithTimeout(ctx, gitTimeout)
	defer cancel()

	log := commitLog{Branch: "HEAD", Ref: ref}

	revision := ref
	if revision == "" {
		revision = "HEAD"

		branch, err := gitText(ctx, root, "symbolic-ref", "--quiet", "--short", "HEAD")
		switch {
		case err == nil && branch != "":
			log.Branch = branch
		case err != nil:
			if short, shortErr := gitText(ctx, root, "rev-parse", "--short", "HEAD"); shortErr == nil && short != "" {
				log.Branch = "HEAD detached at " + short
			}
		}
	} else {
		log.Branch = branchLabel(ref)
	}

	head, err := gitText(ctx, root, "rev-parse", "--verify", revision)
	if err != nil {
		if ref != "" {
			return log, err
		}

		log.Unborn = true
		log.Note = "No commits on this branch yet."

		return log, nil
	}

	log.Head = head

	raw, err := gitBytes(ctx, root,
		"log", "-n", fmt.Sprint(limit), "--topo-order",
		"--no-color", "--no-decorate", "--no-notes", "--no-show-signature",
		"--name-status", "--numstat",
		"--format=%x1e%H%x00%h%x00%P%x00%an%x00%aI%x00%s%x00%b%x00",
		revision, "--",
	)
	if err != nil {
		return log, err
	}

	log.Commits = parseCommitLog(string(raw))
	if len(log.Commits) == 0 {
		log.Note = "No non-merge commits on this branch yet."
	}

	return log, nil
}

// parseCommitLog turns the NUL and record separated git log output into rows.
func parseCommitLog(raw string) []gitCommit {
	records := strings.Split(raw, "\x1e")
	commits := make([]gitCommit, 0, len(records))

	for _, record := range records {
		if strings.TrimSpace(record) == "" {
			continue
		}

		commit, ok := parseCommitRecord(record)
		if ok {
			commits = append(commits, commit)
		}
	}

	return commits
}

func parseCommitRecord(record string) (gitCommit, bool) {
	fields := strings.SplitN(record, "\x00", commitFields)
	if len(fields) < commitFields {
		return gitCommit{}, false
	}

	subject := strings.TrimSpace(fields[5])
	body := strings.TrimSpace(fields[6])
	if len([]rune(body)) > maxCommitBody {
		body = string([]rune(body)[:maxCommitBody])
	}

	commit := gitCommit{
		Hash:    strings.TrimSpace(fields[0]),
		Short:   strings.TrimSpace(fields[1]),
		Parents: parseCommitParents(fields[2]),
		Author:  strings.TrimSpace(fields[3]),
		Date:    formatCommitDate(strings.TrimSpace(fields[4])),
		Subject: subject,
		Body:    body,
	}

	if len(commit.Hash) != commitHashLen {
		return gitCommit{}, false
	}

	commit.Files = parseCommitFiles(fields[7])

	for _, file := range commit.Files {
		commit.Add += file.Add
		commit.Del += file.Del
	}

	if len(commit.Files) > maxCommitFiles {
		commit.More = len(commit.Files) - maxCommitFiles
		commit.Files = commit.Files[:maxCommitFiles]
	}

	return commit, true
}

// parseCommitParents splits the space-separated %P field and keeps only the
// 40-character hex hashes. A root commit has no parents and yields nil.
func parseCommitParents(raw string) []string {
	var parents []string

	for _, field := range strings.Fields(raw) {
		if validCommitHash(field) {
			parents = append(parents, field)
		}
	}

	return parents
}

func formatCommitDate(raw string) string {
	if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
		return parsed.Format("2006-01-02")
	}

	if len(raw) > commitDateLen {
		return raw[:10]
	}

	return raw
}

// parseCommitFiles reads the --name-status and --numstat blocks.
func parseCommitFiles(block string) []commitFile {
	byPath := map[string]*commitFile{}
	order := make([]string, 0, commitFileSeeds)

	for _, line := range strings.Split(block, "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}

		if file, ok := parseNameStatus(line); ok {
			copy := file
			if _, seen := byPath[file.Path]; !seen {
				order = append(order, file.Path)
			}

			byPath[file.Path] = &copy

			continue
		}

		if add, del, path, binary, ok := parseNumstat(line); ok {
			file := byPath[path]
			if file == nil {
				file = &commitFile{Path: path, Status: "M"}
				byPath[path] = file
				order = append(order, path)
			}

			file.Add, file.Del, file.Binary = add, del, binary
		}
	}

	files := make([]commitFile, 0, len(order))

	for _, path := range order {
		if file := byPath[path]; file != nil {
			files = append(files, *file)
		}
	}

	return files
}

func parseNameStatus(line string) (commitFile, bool) {
	tab := strings.IndexByte(line, '\t')
	if tab < 1 {
		return commitFile{}, false
	}

	status := line[:tab]
	code := status[0]

	switch code {
	case 'M', 'A', 'D', 'T', 'U', 'X':
		return commitFile{Path: line[tab+1:], Status: string(code)}, true
	case 'R', 'C':
		parts := strings.Split(line, "\t")
		if len(parts) < commitParts {
			return commitFile{}, false
		}

		return commitFile{Path: parts[2], OldPath: parts[1], Status: string(code)}, true
	default:
		return commitFile{}, false
	}
}

func parseNumstat(line string) (int, int, string, bool, bool) {
	parts := strings.Split(line, "\t")
	if len(parts) < commitParts {
		return 0, 0, "", false, false
	}

	path := parts[len(parts)-1]
	if path == "" {
		return 0, 0, "", false, false
	}

	if arrow := strings.Index(path, " => "); arrow >= 0 {
		path = path[arrow+4:]
	}

	add, addOK := parseStat(parts[0])
	del, delOK := parseStat(parts[1])
	binary := parts[0] == "-" || parts[1] == "-"
	if !addOK && !delOK && !binary {
		return 0, 0, "", false, false
	}

	return add, del, path, binary, true
}

func parseStat(raw string) (int, bool) {
	if raw == "-" || raw == "" {
		return 0, false
	}

	n := 0
	for _, r := range raw {
		if r < '0' || r > '9' {
			return 0, false
		}

		n = n*decimalBase + int(r-'0')
	}

	return n, true
}

// loadCommitFile reads one file patch from a commit. The hash is validated
// before it reaches git's argument list.
func loadCommitFile(ctx context.Context, root, hash, path string) (*changedFile, error) {
	ctx, cancel := context.WithTimeout(ctx, gitTimeout)
	defer cancel()

	if !validCommitHash(hash) {
		return nil, errors.New("invalid commit hash")
	}

	if path == "" {
		return nil, errors.New("empty commit path")
	}

	raw, _, err := gitBytesCapped(ctx, root, maxCommitPatchBytes,
		"show", "--format=", "--no-color", "--no-ext-diff", "--no-textconv",
		"--no-show-signature", "--diff-merges=first-parent", "-M",
		hash, "--", path,
	)
	if err != nil {
		return nil, err
	}

	files := parseDiff(string(raw))

	for i := range files {
		if files[i].Path == path || files[i].OldPath == path {
			return &files[i], nil
		}
	}

	return nil, fmt.Errorf("no diff for %s in %s", path, hash[:7])
}

func validCommitHash(hash string) bool {
	if len(hash) != commitHashLen {
		return false
	}

	_, err := hex.DecodeString(hash)

	return err == nil
}

// commitDetailKey keys one parsed file patch in the commit cache.
func commitDetailKey(hash, path string) string { return hash + "\x00" + path }

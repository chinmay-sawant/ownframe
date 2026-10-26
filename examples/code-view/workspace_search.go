package main

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"unicode"
)

const (
	workspaceSearchLimit      = 200
	workspaceSearchFileHits   = 8
	workspaceSearchTextRunes  = 240
	workspaceSearchCancelNote = "Search cancelled."
)

// workspaceHit is one occurrence of a workspace query. Line and Col are
// 1-based, and Col counts runes in the display line, so a tab is four spaces.
// Before, Match, and After split Text around the occurrence.
type workspaceHit struct {
	Path   string
	Line   int
	Col    int
	Text   string
	Before string
	Match  string
	After  string
}

// workspaceSearchResult is one pass over the workspace. Files counts the
// distinct paths in Hits, not every file that was scanned. Note explains a
// truncated or cancelled pass.
type workspaceSearchResult struct {
	Query     string
	Hits      []workspaceHit
	Files     int
	Truncated bool
	Note      string
}

// searchWorkspace scans the text files of ws for query. An empty query has no
// hits. A file larger than maxFileBytes or one that looks binary is skipped
// before its bytes are read. Matches are non-overlapping, at most
// workspaceSearchFileHits per file, and the total is capped at limit, or
// workspaceSearchLimit when limit is not positive.
func searchWorkspace(ctx context.Context, ws *workspace, query string, caseSensitive bool, limit int) workspaceSearchResult {
	result := workspaceSearchResult{Query: query}
	if ws == nil || query == "" {
		return result
	}

	if limit <= 0 {
		limit = workspaceSearchLimit
	}

	queryRunes := []rune(query)
	limitStopped := false
	perFileStopped := false

search:
	for _, meta := range ws.Files {
		if meta == nil {
			continue
		}

		if ctx.Err() != nil {
			result.Truncated = perFileStopped
			result.Note = workspaceSearchCancelNote

			return finishWorkspaceSearch(result)
		}

		if meta.Big || meta.Size > maxFileBytes {
			continue
		}

		raw, err := ws.readFile(meta.Path)
		if err != nil || len(raw) > maxFileBytes || isBinary(raw) {
			continue
		}

		text := strings.TrimPrefix(string(raw), "\ufeff")
		fileHits := 0

	fileLines:
		for lineIndex, rawLine := range strings.Split(text, "\n") {
			if ctx.Err() != nil {
				result.Truncated = perFileStopped
				result.Note = workspaceSearchCancelNote

				return finishWorkspaceSearch(result)
			}

			line := []rune(expandTabs(rawLine))
			if len(line) > workspaceSearchTextRunes {
				line = line[:workspaceSearchTextRunes]
			}

			for start := 0; start+len(queryRunes) <= len(line); {
				if !workspaceMatchAt(line[start:start+len(queryRunes)], queryRunes, caseSensitive) {
					start++

					continue
				}

				if len(result.Hits) >= limit {
					limitStopped = true

					break search
				}

				if fileHits >= workspaceSearchFileHits {
					perFileStopped = true

					break fileLines
				}

				end := start + len(queryRunes)
				result.Hits = append(result.Hits, workspaceHit{
					Path:   meta.Path,
					Line:   lineIndex + 1,
					Col:    start + 1,
					Text:   string(line),
					Before: string(line[:start]),
					Match:  string(line[start:end]),
					After:  string(line[end:]),
				})
				fileHits++
				start = end
			}
		}
	}

	result.Truncated = limitStopped || perFileStopped
	result.Note = workspaceSearchNote(limitStopped, perFileStopped, limit)

	return finishWorkspaceSearch(result)
}

// workspaceMatchAt reports whether candidate starts with query.
func workspaceMatchAt(candidate, query []rune, caseSensitive bool) bool {
	for i, r := range query {
		if caseSensitive {
			if candidate[i] != r {
				return false
			}

			continue
		}

		if !workspaceRuneEqualFold(candidate[i], r) {
			return false
		}
	}

	return true
}

// workspaceRuneEqualFold compares one rune under Unicode simple folding.
func workspaceRuneEqualFold(a, b rune) bool {
	if a == b {
		return true
	}

	for r := unicode.SimpleFold(a); r != a; r = unicode.SimpleFold(r) {
		if r == b {
			return true
		}
	}

	return false
}

// workspaceSearchNote describes why a pass stopped before its last match.
func workspaceSearchNote(limitStopped, perFileStopped bool, limit int) string {
	var notes []string

	if limitStopped {
		notes = append(notes, fmt.Sprintf("Showing the first %d matches; more were found.", limit))
	}

	if perFileStopped {
		notes = append(notes, fmt.Sprintf("Some files have more than %d matches.", workspaceSearchFileHits))
	}

	return strings.Join(notes, " ")
}

// finishWorkspaceSearch orders the hits and counts the files they came from.
func finishWorkspaceSearch(result workspaceSearchResult) workspaceSearchResult {
	sort.Slice(result.Hits, func(i, j int) bool {
		a, b := result.Hits[i], result.Hits[j]
		if a.Path != b.Path {
			return a.Path < b.Path
		}

		if a.Line != b.Line {
			return a.Line < b.Line
		}

		return a.Col < b.Col
	})

	for i, hit := range result.Hits {
		if i == 0 || hit.Path != result.Hits[i-1].Path {
			result.Files++
		}
	}

	return result
}

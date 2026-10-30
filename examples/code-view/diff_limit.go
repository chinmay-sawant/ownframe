package main

import "fmt"

// The diff pane applies the same text budget as the editor. A tracked
// deletion or a rewrite can carry a whole minified bundle into the working
// tree diff even though opening that file would be refused by maxFileBytes.
const (
	maxDiffBytes = maxFileBytes
	maxDiffLines = maxFileLines
)

// measureDiff records the text size the diff pane would render and marks a
// file over the budget. It runs before word emphasis so an oversized diff
// does not pay for pairing that no row will use.
func measureDiff(file *changedFile) {
	bytes, lines, longest := 0, 0, 0

	for _, h := range file.Hunks {
		bytes += len(h.Header) + 1

		for _, line := range h.Lines {
			bytes += len(line.Text) + 1
			lines++
			if len(line.Text) > longest {
				longest = len(line.Text)
			}
		}
	}

	file.DiffBytes = bytes
	file.DiffLines = lines
	file.DiffMaxLine = longest
	file.Oversized = bytes > maxDiffBytes || lines > maxDiffLines || longest > maxLineBytes
}

// diffLimitNote is the row shown instead of a diff over the budget. The
// wording names the measured size so the limit is not a mystery.
func diffLimitNote(file changedFile) string {
	if file.DiffTruncated {
		return fmt.Sprintf(
			"Diff skipped: the patch stopped at the %d KB read limit.",
			maxCommitPatchBytes/byteKB,
		)
	}

	if file.DiffMaxLine > maxLineBytes {
		return fmt.Sprintf(
			"Diff skipped: one line is %s, over the %d KB line diff limit.",
			humanSize(int64(file.DiffMaxLine)), maxLineKB,
		)
	}

	return fmt.Sprintf(
		"Diff skipped: %s across %d lines, over the %d KB and %d line diff limit.",
		humanSize(int64(file.DiffBytes)), file.DiffLines, maxFileBytesKB, maxDiffLines,
	)
}

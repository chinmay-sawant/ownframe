package main

import (
	"strconv"
	"strings"
	"unicode"
)

// changeSet is the pending diff the Source Control view shows.
type changeSet struct {
	Files      []changedFile
	Branch     string
	Sample     bool
	Note       string
	EmptyTitle string
}

// changedFile is one file in a diff, with its hunks already parsed.
type changedFile struct {
	Path       string
	OldPath    string
	Status     string
	Add        int
	Del        int
	Hunks      []hunk
	Binary     bool
	TotalLines int
}

// hunk is one `@@` block from a unified diff.
type hunk struct {
	OldStart int
	OldCount int
	NewStart int
	NewCount int
	Header   string
	Lines    []diffLine
	oldNext  int
	newNext  int
}

// diffLine is one context, added, or deleted line.
type diffLine struct {
	Kind   byte
	OldNo  int
	NewNo  int
	Text   string
	Emph   bool
	EmphAt int
	EmphTo int
}

// diffRow is one rendered row of the diff, either unified or split.
type diffRow struct {
	Kind      string
	Hunk      string
	OldNo     string
	NewNo     string
	LeftNo    string
	RightNo   string
	Left      []segment
	Right     []segment
	LeftKind  string
	RightKind string
	LeftPad   bool
	RightPad  bool
	Cont      bool
	LeftCont  bool
	RightCont bool
	Expand    *diffExpandRow
	ScrollX   int
}

// clipDiffRows trims long diff text so a cell cannot paint into its neighbour.
func clipDiffRows(rows []diffRow, unified, split int) {
	for i := range rows {
		row := &rows[i]
		if row.Kind == rowExpand {
			continue
		}

		cols := unified

		if row.Kind == "split" {
			cols = split
			row.Left = clipSegments(row.Left, cols)
			row.Right = clipSegments(row.Right, cols)

			continue
		}

		row.Left = clipSegments(row.Left, cols)
		row.Right = clipSegments(row.Right, cols)

		if row.Hunk != "" {
			runes := []rune(row.Hunk)
			if len(runes) > cols {
				row.Hunk = string(runes[:cols]) + "…"
			}
		}
	}
}

// clipSegments keeps at most max runes and marks the cut with an ellipsis.
func clipSegments(segs []segment, max int) []segment {
	if max < 1 {
		return nil
	}

	var (
		out   []segment
		count int
	)

	for _, seg := range segs {
		runes := []rune(seg.Text)
		if count+len(runes) <= max {
			out = append(out, seg)
			count += len(runes)

			continue
		}

		take := max - count
		if take > 0 {
			out = append(out, segment{Text: string(runes[:take]) + "…", Class: seg.Class})
		} else if len(out) > 0 {
			out[len(out)-1].Text += "…"
		}

		return out
	}

	return out
}

const (
	diffLinesPerFileEstimate = 8
	diffRowsPerHunkEstimate  = 8
	markRangeParts           = 2

	rowHunk   = "hunk"
	rowCtx    = "ctx"
	rowAdd    = "add"
	rowDel    = "del"
	rowMeta   = "meta"
	rowExpand = "expand"
)

// parseDiff turns unified diff text into files and hunks. It accepts the
// output of `git diff`, including new, deleted, renamed, and binary files.
func parseDiff(text string) []changedFile {
	lines := strings.Split(text, "\n")
	files := make([]changedFile, 0, len(lines)/diffLinesPerFileEstimate)
	var current *changedFile
	var active *hunk

	flush := func() {
		if current == nil {
			return
		}

		captureHunkEnd(active, current)
		files = append(files, *current)
		current = nil
		active = nil
	}

	for li, line := range lines {
		if li == len(lines)-1 && line == "" {
			break
		}

		switch {
		case strings.HasPrefix(line, "diff --git "):
			flush()

			oldPath, newPath := parseDiffGitPaths(strings.TrimPrefix(line, "diff --git "))
			current = &changedFile{Path: newPath, OldPath: oldPath, Status: "M"}
		case current == nil:
			continue
		case strings.HasPrefix(line, "new file mode"):
			current.Status = "A"
		case strings.HasPrefix(line, "deleted file mode"):
			current.Status = "D"
		case strings.HasPrefix(line, "rename from "):
			current.OldPath = strings.TrimPrefix(line, "rename from ")
			current.Status = "R"
		case strings.HasPrefix(line, "rename to "):
			current.Path = strings.TrimPrefix(line, "rename to ")
			current.Status = "R"
		case strings.HasPrefix(line, "Binary files ") || strings.HasPrefix(line, "GIT binary patch"):
			current.Binary = true
		case strings.HasPrefix(line, "--- "):
			path := parseDiffPath(strings.TrimPrefix(line, "--- "))
			if path == "" {
				continue
			}

			if current.OldPath == "" {
				current.OldPath = path
			}
		case strings.HasPrefix(line, "+++ "):
			path := parseDiffPath(strings.TrimPrefix(line, "+++ "))
			if path != "" && current.Status != "D" {
				current.Path = path
			}
		case strings.HasPrefix(line, "@@ "):
			captureHunkEnd(active, current)
			active = parseHunkHeader(line)
		case active != nil && strings.HasPrefix(line, "\\"):
			continue
		case active != nil && len(line) > 0 && (line[0] == '+' || line[0] == '-' || line[0] == ' '):
			addDiffLine(active, line[0], line[1:])
		}
	}

	flush()

	for i := range files {
		countHunkLines(&files[i])
		applyEmphasis(&files[i])
	}

	return files
}

// captureHunkEnd appends the open hunk to its file.
func captureHunkEnd(active *hunk, file *changedFile) {
	if active == nil || file == nil {
		return
	}

	file.Hunks = append(file.Hunks, *active)
}

// parseDiffGitPaths splits `a/path b/path` into a plain pair.
func parseDiffGitPaths(rest string) (string, string) {
	rest = strings.TrimSpace(rest)

	if i := strings.Index(rest, " b/"); i >= 0 {
		oldPath := strings.TrimPrefix(rest[:i], "a/")
		newPath := strings.TrimPrefix(rest[i+1:], "b/")

		return oldPath, newPath
	}

	if i := strings.Index(rest, " "); i >= 0 {
		return parseDiffPath(rest[:i]), parseDiffPath(rest[i+1:])
	}

	return "", parseDiffPath(rest)
}

// parseDiffPath strips a git prefix and rejects the null side.
func parseDiffPath(path string) string {
	path = strings.TrimSpace(path)

	if path == "" || path == "/dev/null" {
		return ""
	}

	if i := strings.IndexByte(path, '\t'); i >= 0 {
		path = path[:i]
	}

	path = strings.Trim(path, "\"")

	if strings.HasPrefix(path, "a/") || strings.HasPrefix(path, "b/") {
		path = path[2:]
	}

	return path
}

// parseHunkHeader reads `@@ -1,2 +3,4 @@ func name`.
func parseHunkHeader(line string) *hunk {
	rest, ok := strings.CutPrefix(line, "@@ -")
	if !ok {
		return nil
	}

	h := &hunk{OldCount: 1, NewCount: 1}

	h.OldStart, rest = parseHunkNumber(rest)

	if strings.HasPrefix(rest, ",") {
		h.OldCount, rest = parseHunkNumber(rest[1:])
	}

	rest, ok = strings.CutPrefix(rest, " +")
	if !ok {
		return h
	}

	h.NewStart, rest = parseHunkNumber(rest)

	if strings.HasPrefix(rest, ",") {
		h.NewCount, rest = parseHunkNumber(rest[1:])
	}

	if tail, ok := strings.CutPrefix(rest, " @@"); ok {
		h.Header = strings.TrimSpace(tail)
	}

	h.oldNext, h.newNext = h.OldStart, h.NewStart

	return h
}

// parseHunkNumber reads one decimal and returns the remainder.
func parseHunkNumber(s string) (int, string) {
	i := 0

	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}

	n, err := strconv.Atoi(s[:i])
	if err != nil {
		return 0, s[i:]
	}

	return n, s[i:]
}

// addDiffLine appends one hunk line and advances the old and new counters.
func addDiffLine(h *hunk, kind byte, text string) {
	oldNo, newNo := h.oldNext, h.newNext

	switch kind {
	case '+':
		h.Lines = append(h.Lines, diffLine{Kind: '+', NewNo: newNo, Text: expandTabs(text)})
		h.newNext++
	case '-':
		h.Lines = append(h.Lines, diffLine{Kind: '-', OldNo: oldNo, Text: expandTabs(text)})
		h.oldNext++
	default:
		h.Lines = append(h.Lines, diffLine{Kind: ' ', OldNo: oldNo, NewNo: newNo, Text: expandTabs(text)})
		h.oldNext++
		h.newNext++
	}
}

// countHunkLines fills Add and Del for one file.
func countHunkLines(file *changedFile) {
	for _, h := range file.Hunks {
		for _, line := range h.Lines {
			switch line.Kind {
			case '+':
				file.Add++
			case '-':
				file.Del++
			}
		}
	}
}

// applyEmphasis pairs each run of deletions with the following additions and
// marks the changed middle, which is the word-level highlight.
func applyEmphasis(file *changedFile) {
	for hi := range file.Hunks {
		h := &file.Hunks[hi]
		i := 0

		for i < len(h.Lines) {
			if h.Lines[i].Kind != '-' {
				i++

				continue
			}

			delStart := i
			for i < len(h.Lines) && h.Lines[i].Kind == '-' {
				i++
			}

			addStart := i
			for i < len(h.Lines) && h.Lines[i].Kind == '+' {
				i++
			}

			pairs := min(addStart-delStart, i-addStart)
			for k := range pairs {
				emphasizePair(&h.Lines[delStart+k], &h.Lines[addStart+k])
			}
		}
	}
}

// emphasizePair marks the differing middle of one deleted and one added line.
func emphasizePair(del, add *diffLine) {
	a := []rune(del.Text)
	b := []rune(add.Text)

	pre := 0
	for pre < len(a) && pre < len(b) && a[pre] == b[pre] {
		pre++
	}

	suf := 0
	for suf < len(a)-pre && suf < len(b)-pre && a[len(a)-1-suf] == b[len(b)-1-suf] {
		suf++
	}

	startA, endA := pre, len(a)-suf
	startB, endB := pre, len(b)-suf

	if startA == endA && startB == endB {
		return
	}

	if startA < len(a) && startA > 0 && isWordRune(a[startA-1]) && isWordRune(a[startA]) {
		for startA > 0 && isWordRune(a[startA-1]) {
			startA--
		}

		startB = startA
	}

	for endA < len(a) && endA > startA && isWordRune(a[endA-1]) && isWordRune(a[endA]) {
		endA++
	}

	for endB < len(b) && endB > startB && isWordRune(b[endB-1]) && isWordRune(b[endB]) {
		endB++
	}

	if endA > startA {
		del.Emph, del.EmphAt, del.EmphTo = true, startA, endA
	}

	if endB > startB {
		add.Emph, add.EmphAt, add.EmphTo = true, startB, endB
	}
}

// isWordRune reports whether a rune continues an identifier-like word.
func isWordRune(r rune) bool {
	return r == '_' || r == '$' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

// expandTabs replaces tabs with four spaces and strips a trailing CR. Tabs
// make column alignment impossible, so the viewer never prints one.
func expandTabs(text string) string {
	text = strings.TrimSuffix(text, "\r")

	if !strings.ContainsRune(text, '\t') {
		return text
	}

	var b strings.Builder
	b.Grow(len(text) + tabWidth)

	for _, r := range text {
		if r == '\t' {
			b.WriteString(tabSpaces)

			continue
		}

		b.WriteRune(r)
	}

	return b.String()
}

// diffRowsUnified builds the one-column row list.
func diffRowsUnified(file changedFile, lang string) []diffRow {
	rows := make([]diffRow, 0, len(file.Hunks)*diffRowsPerHunkEstimate)

	if file.Binary {
		return []diffRow{{Kind: rowMeta, Hunk: "Binary file. The viewer shows two lines of metadata instead of a diff."}}
	}

	if len(file.Hunks) == 0 {
		return []diffRow{{Kind: rowMeta, Hunk: "No textual changes in this file."}}
	}

	gaps := identifyGaps(file)
	var gapBefore *diffGap
	var gapAfter *diffGap
	gapBetween := make(map[int]diffGap)

	for _, g := range gaps {
		if g.Before {
			copy := g
			gapBefore = &copy
		} else if g.After {
			copy := g
			gapAfter = &copy
		} else {
			gapBetween[g.HunkIdx] = g
		}
	}

	if gapBefore != nil {
		rows = append(rows, newDiffExpandRow(*gapBefore))
	}

	for i, h := range file.Hunks {
		rows = append(rows, diffRow{Kind: rowHunk, Hunk: hunkLabel(h)})

		for _, line := range h.Lines {
			row := diffRow{Kind: rowCtx, OldNo: lineNo(line.OldNo), NewNo: lineNo(line.NewNo)}
			row.Left = diffSegments(line, lang)

			switch line.Kind {
			case '+':
				row.Kind = rowAdd
				row.OldNo = ""
			case '-':
				row.Kind = rowDel
				row.NewNo = ""
			}

			rows = append(rows, row)
		}

		if g, ok := gapBetween[i]; ok {
			rows = append(rows, newDiffExpandRow(g))
		}
	}

	if gapAfter != nil {
		rows = append(rows, newDiffExpandRow(*gapAfter))
	}

	return rows
}

// diffRowsSplit builds the two-column row list.
func diffRowsSplit(file changedFile, lang string) []diffRow {
	rows := make([]diffRow, 0, len(file.Hunks)*diffRowsPerHunkEstimate)

	if file.Binary {
		return []diffRow{{Kind: rowMeta, Hunk: "Binary file. The viewer shows two lines of metadata instead of a diff."}}
	}

	if len(file.Hunks) == 0 {
		return []diffRow{{Kind: rowMeta, Hunk: "No textual changes in this file."}}
	}

	gaps := identifyGaps(file)
	var gapBefore *diffGap
	var gapAfter *diffGap
	gapBetween := make(map[int]diffGap)

	for _, g := range gaps {
		if g.Before {
			copy := g
			gapBefore = &copy
		} else if g.After {
			copy := g
			gapAfter = &copy
		} else {
			gapBetween[g.HunkIdx] = g
		}
	}

	if gapBefore != nil {
		rows = append(rows, newDiffExpandRow(*gapBefore))
	}

	for hunkIdx, h := range file.Hunks {
		rows = append(rows, diffRow{Kind: rowHunk, Hunk: hunkLabel(h)})

		i := 0
		for i < len(h.Lines) {
			if h.Lines[i].Kind == ' ' {
				line := h.Lines[i]
				rows = append(rows, diffRow{
					Kind: "split", LeftNo: lineNo(line.OldNo), RightNo: lineNo(line.NewNo),
					Left: diffSegments(line, lang), Right: diffSegments(line, lang),
					LeftKind: rowCtx, RightKind: rowCtx,
				})
				i++

				continue
			}

			delStart := i
			for i < len(h.Lines) && h.Lines[i].Kind == '-' {
				i++
			}

			addStart := i
			for i < len(h.Lines) && h.Lines[i].Kind == '+' {
				i++
			}

			dels, adds := delStart, addStart
			pairs := max(i-addStart, addStart-delStart)

			for k := range pairs {
				row := diffRow{Kind: "split", LeftKind: rowDel, RightKind: rowAdd}

				if dels+k < addStart {
					line := h.Lines[dels+k]
					row.LeftNo = lineNo(line.OldNo)
					row.Left = diffSegments(line, lang)
				} else {
					row.LeftPad = true
				}

				if adds+k < i {
					line := h.Lines[adds+k]
					row.RightNo = lineNo(line.NewNo)
					row.Right = diffSegments(line, lang)
				} else {
					row.RightPad = true
				}

				rows = append(rows, row)
			}
		}

		if g, ok := gapBetween[hunkIdx]; ok {
			rows = append(rows, newDiffExpandRow(g))
		}
	}

	if gapAfter != nil {
		rows = append(rows, newDiffExpandRow(*gapAfter))
	}

	return rows
}

// diffSegments highlights one diff line and applies the word emphasis.
func diffSegments(line diffLine, lang string) []segment {
	if line.Kind == '+' || line.Kind == '-' {
		if line.Emph {
			segs := highlightLine(lang, line.Text)
			class := "em-d"
			if line.Kind == '+' {
				class = "em-a"
			}

			return shiftAll(markRange(segs, line.EmphAt, line.EmphTo, class))
		}
	}

	return highlightLine(lang, line.Text)
}

// markRange splits segments so the rune range [start,end) carries an extra
// class. It never rewrites text, so the template keeps escaping it.
func markRange(segs []segment, start, end int, extra string) []segment {
	if start >= end {
		return segs
	}

	out := make([]segment, 0, len(segs)+markRangeParts)
	pos := 0

	for _, seg := range segs {
		runes := []rune(seg.Text)
		from, to := pos, pos+len(runes)
		pos = to

		if to <= start || from >= end {
			out = append(out, seg)

			continue
		}

		cut := func(a, b int) int {
			return max(a, min(b, len(runes)))
		}

		before := cut(0, start-from)
		middleEnd := cut(before, end-from)

		if before > 0 {
			out = append(out, segment{Text: string(runes[:before]), Class: seg.Class})
		}

		if middleEnd > before {
			out = append(out, segment{Text: string(runes[before:middleEnd]), Class: seg.Class + " " + extra})
		}

		if middleEnd < len(runes) {
			out = append(out, segment{Text: string(runes[middleEnd:]), Class: seg.Class})
		}
	}

	return out
}

// hunkLabel renders a standard hunk range, with the enclosing symbol when
// git supplied one.
func hunkLabel(h hunk) string {
	old := strconv.Itoa(h.OldStart)
	if h.OldCount != 1 {
		old += "," + strconv.Itoa(h.OldCount)
	}

	fresh := strconv.Itoa(h.NewStart)
	if h.NewCount != 1 {
		fresh += "," + strconv.Itoa(h.NewCount)
	}

	label := "@@ -" + old + " +" + fresh + " @@"
	if h.Header != "" {
		label += " " + h.Header
	}

	return label
}

// lineNo formats a line number, with zero shown as a blank gutter cell.
func lineNo(n int) string {
	if n <= 0 {
		return ""
	}

	return strconv.Itoa(n)
}

// totals sums the additions and deletions across a change set.
func (c changeSet) totals() (int, int) {
	add, del := 0, 0

	for _, file := range c.Files {
		add += file.Add
		del += file.Del
	}

	return add, del
}

// findFile returns the changed file with the given path.
func findFile(files []changedFile, path string) *changedFile {
	for i := range files {
		if files[i].Path == path {
			return &files[i]
		}
	}

	return nil
}

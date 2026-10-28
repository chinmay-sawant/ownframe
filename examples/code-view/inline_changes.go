package main

import "strings"

// inlineChanges maps each changed 1-based line of the new file to its change.
// The map key is the line the editor shows, so an added or modified line sits
// on its own new line, and a deletion hangs on the next new line. Binary files
// and files with no hunks yield nil. Context lines are not stored.
//
// A deletion pairs one-for-one with the additions that immediately follow its
// run in the same hunk, which becomes inlineModified. An unpaired addition is
// inlineAdded. An unpaired deletion anchors to the next new line in the hunk,
// or to the hunk's NewStart when no later new line exists; several deleted
// lines sharing an anchor keep their text joined with newlines.
func inlineChanges(file changedFile) map[int]inlineLine {
	if file.Binary || len(file.Hunks) == 0 {
		return nil
	}

	changes := make(map[int]inlineLine)

	for hi := range file.Hunks {
		h := &file.Hunks[hi]

		for i := 0; i < len(h.Lines); {
			switch h.Lines[i].Kind {
			case '+':
				addInlineLine(changes, h.Lines[i])
				i++
			case '-':
				i = pairInlineDeletionRun(changes, h, i)
			default:
				i++
			}
		}
	}

	return changes
}

// inlineChangeAt returns the change at a 1-based line of the new file.
func inlineChangeAt(changes map[int]inlineLine, line int) (inlineLine, bool) {
	change, ok := changes[line]
	return change, ok
}

// inlineStats counts added and modified lines of the new file and deleted
// lines of the old file. A deleted entry can hold several old lines, and each
// one counts.
func inlineStats(changes map[int]inlineLine) (added, modified, deleted int) {
	for _, change := range changes {
		switch change.Kind {
		case inlineAdded:
			added++
		case inlineModified:
			modified++
		case inlineDeleted:
			deleted += strings.Count(change.OldText, "\n") + 1
		}
	}

	return added, modified, deleted
}

// addInlineLine stores one addition at its new-file line.
func addInlineLine(changes map[int]inlineLine, line diffLine) {
	if line.NewNo < 1 {
		return
	}

	changes[line.NewNo] = inlineLine{Kind: inlineAdded}
}

// pairInlineDeletionRun handles one run of deletions and the additions that
// immediately follow it, returning the index after that run. The first
// min(deletions, additions) pairs become inlineModified; the rest are added or
// deleted on their own.
func pairInlineDeletionRun(changes map[int]inlineLine, h *hunk, start int) int {
	delStart := start
	for start < len(h.Lines) && h.Lines[start].Kind == '-' {
		start++
	}

	addStart := start
	for start < len(h.Lines) && h.Lines[start].Kind == '+' {
		start++
	}

	dels := h.Lines[delStart:addStart]
	adds := h.Lines[addStart:start]
	pairs := min(len(dels), len(adds))

	for k := 0; k < pairs; k++ {
		if adds[k].NewNo < 1 {
			continue
		}

		changes[adds[k].NewNo] = inlineLine{Kind: inlineModified, OldText: dels[k].Text}
	}

	for k := pairs; k < len(adds); k++ {
		addInlineLine(changes, adds[k])
	}

	if pairs < len(dels) {
		anchor := inlineDeletionAnchor(changes, h, start)
		mergeInlineDeletions(changes, anchor, dels[pairs:])
	}

	return start
}

// inlineDeletionAnchor returns the new line a deletion with no paired addition
// hangs on: the first later line in the hunk that exists in the new file, or
// NewStart when the deletion ends the hunk. When NewStart already holds an
// earlier change from the same hunk, the anchor shifts one past the hunk's last
// new line so no marker is displaced.
func inlineDeletionAnchor(changes map[int]inlineLine, h *hunk, from int) int {
	for i := from; i < len(h.Lines); i++ {
		if h.Lines[i].NewNo > 0 {
			return h.Lines[i].NewNo
		}
	}

	anchor := h.NewStart
	if anchor < 1 {
		return 0
	}

	if existing, ok := changes[anchor]; ok && existing.Kind != inlineDeleted {
		if last := lastInlineNewLine(h); last >= anchor {
			return last + 1
		}
	}

	return anchor
}

// lastInlineNewLine returns the highest new-file line in one hunk, or zero
// when the hunk only deletes.
func lastInlineNewLine(h *hunk) int {
	for i := len(h.Lines) - 1; i >= 0; i-- {
		if h.Lines[i].NewNo > 0 {
			return h.Lines[i].NewNo
		}
	}

	return 0
}

// mergeInlineDeletions stores the old text of a deletion group at anchor. A
// group already there appends rather than replaces, and a marker of another
// kind is left alone.
func mergeInlineDeletions(changes map[int]inlineLine, anchor int, dels []diffLine) {
	if anchor < 1 || len(dels) == 0 {
		return
	}

	texts := make([]string, 0, len(dels))
	for _, line := range dels {
		texts = append(texts, line.Text)
	}

	text := strings.Join(texts, "\n")

	existing, ok := changes[anchor]
	if ok && existing.Kind != inlineDeleted {
		return
	}

	if ok {
		existing.OldText += "\n" + text
		changes[anchor] = existing

		return
	}

	changes[anchor] = inlineLine{Kind: inlineDeleted, OldText: text}
}

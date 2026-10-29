package main

import (
	"context"
	"strconv"
	"strings"
)

// revertClickedBlock resolves the line control back to its contiguous source
// block. Display rows may be paired or wrapped; source line numbers stay stable.
func (a *App) revertClickedBlock(ctx context.Context, index int, id string) error {
	file := findFile(a.changes.Files, a.diffPath)
	if file == nil || index < 0 || index >= len(file.Hunks) {
		return nil
	}
	_, coordinates, ok := strings.Cut(strings.TrimPrefix(id, "revert-line-"), "-")
	if !ok {
		return nil
	}
	oldText, newText, ok := strings.Cut(coordinates, "-")
	if !ok {
		return nil
	}
	oldNo, _ := strconv.Atoi(oldText)
	newNo, _ := strconv.Atoi(newText)
	source := file.Hunks[index]
	oldCursor := source.OldStart
	for i := 0; i < len(source.Lines); {
		if source.Lines[i].Kind == ' ' {
			oldCursor++
			i++
			continue
		}
		start := oldCursor
		count := 0
		selected := false
		for i < len(source.Lines) && source.Lines[i].Kind != ' ' {
			line := source.Lines[i]
			selected = selected || (oldNo > 0 && line.OldNo == oldNo) || (newNo > 0 && line.NewNo == newNo)
			if line.Kind == '-' {
				count++
				oldCursor++
			}
			i++
		}
		if selected {
			if count == 0 {
				start--
			}
			return a.revertHunk(ctx, index, hunk{OldStart: start, OldCount: count})
		}
	}
	return nil
}

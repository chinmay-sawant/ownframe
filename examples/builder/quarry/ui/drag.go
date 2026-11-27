package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

// Split limits in CSS pixels.
const (
	minSidebarW, maxSidebarW     = 180, 440
	minEditorH, maxEditorH       = 140, 520
	minInspectorW, maxInspectorW = 240, 560
	minColW, maxColW             = 48, 600
)

func (a *App) onDragStart(ctx context.Context, box ownframe.Box) (bool, error) {
	if a.pressedEditor && a.modal == "" && (strings.HasPrefix(box.ID, "ed-line-") || box.ID == "editor-body" || box.ID == "editor-wrap") {
		a.edDrag = true
		return true, nil
	}
	if key, ok := a.thumbKey(box.ID); ok {
		a.beginThumbDrag(key)
		return true, nil
	}
	if a.modal != "" {
		return false, nil
	}
	if strings.HasPrefix(box.ID, "split-") || strings.HasPrefix(box.ID, "colresize-") {
		a.splitDragging = box.ID
		return true, nil
	}
	return false, nil
}

func (a *App) onDragMove(ctx context.Context, dx, dy float64) error {
	if a.edDrag {
		return a.editorDrag(ctx, dx, dy)
	}
	if a.thumbDrag != "" {
		return a.dragThumb(ctx, dx, dy)
	}
	id := a.splitDragging
	switch {
	case id == "split-sidebar":
		a.sidebarW = clampInt(a.sidebarW+int(dx), minSidebarW, maxSidebarW)
	case id == "split-editor":
		a.editorH = clampInt(a.editorH+int(dy), minEditorH, maxEditorH)
	case id == "split-inspector":
		a.inspectorW = clampInt(a.inspectorW-int(dx), minInspectorW, maxInspectorW)
	case strings.HasPrefix(id, "colresize-"):
		var col int
		if _, err := fmt.Sscanf(id, "colresize-%d", &col); err == nil {
			if t := a.activeTabPtr(); t != nil && t.grid != nil && col < len(t.grid.widths) {
				t.grid.widths[col] = clampInt(t.grid.widths[col]+int(dx), minColW, maxColW)
			}
		}
	default:
		return nil
	}
	return a.redraw(ctx)
}

func (a *App) onDragEnd(ctx context.Context) error {
	if a.edDrag {
		a.edDrag, a.pressedEditor = false, false
		return nil
	}
	if a.thumbDrag != "" {
		a.endThumbDrag()
		return nil
	}
	id := a.splitDragging
	a.splitDragging = ""
	switch id {
	case "split-sidebar":
		a.savePref("split.sidebar", fmt.Sprint(a.sidebarW))
	case "split-editor":
		a.savePref("split.editor", fmt.Sprint(a.editorH))
	case "split-inspector":
		a.savePref("split.inspector", fmt.Sprint(a.inspectorW))
	}
	return nil
}

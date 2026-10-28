package main

import (
	"context"
	"testing"
)

func TestEditorScrollbarInsidePaneAndDrags(t *testing.T) {
	a := newTestApp(t)
	a.openFile("backend/probe.go")
	ctx := context.Background()
	a.page.SetSize(minWidth, minHeight)
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	thumb, ok := boxByID(a.page.Boxes(), "editor-scroll-thumb")
	if !ok {
		t.Fatal("missing editor scrollbar")
	}
	body, ok := boxByID(a.page.Boxes(), "code-body")
	if !ok || thumb.X < body.X || thumb.X+thumb.W > body.X+body.W+1 || thumb.Y < body.Y || thumb.Y+thumb.H > body.Y+body.H+1 {
		t.Fatalf("thumb %+v outside editor %+v", thumb, body)
	}
	claimed, err := a.page.Testing().BeginDrag(ctx, thumb.X+thumb.W/2, thumb.Y+thumb.H/2)
	if err != nil || !claimed {
		t.Fatalf("drag start: %v, %v", claimed, err)
	}
	if err := a.page.Testing().MoveDrag(ctx, thumb.X+thumb.W/2, body.Y+body.H-1); err != nil {
		t.Fatal(err)
	}
	if err := a.page.Testing().EndDrag(ctx); err != nil {
		t.Fatal(err)
	}
	if a.scrollY == 0 {
		t.Fatal("drag did not scroll editor")
	}
}

func TestSplitEditorScrollbarsTrackTheirOwnPanes(t *testing.T) {
	a := newTestApp(t)
	a.openFile("backend/probe.go")
	a.toggleSplit()
	a.openFileIn(1, "backend/probe.go")
	ctx := context.Background()
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	for _, pair := range [][2]string{{"editor-scroll-thumb", "code-body"}, {"editor-scroll-thumb-2", "code-body-2"}} {
		thumb, ok := boxByID(a.page.Boxes(), pair[0])
		if !ok {
			t.Fatalf("missing %s", pair[0])
		}
		body, ok := boxByID(a.page.Boxes(), pair[1])
		if !ok || thumb.X < body.X || thumb.X+thumb.W > body.X+body.W+1 {
			t.Fatalf("%s outside %s", pair[0], pair[1])
		}
	}
	a.scrollGroupTo(1, 200)
	if a.scrollGroup[0] != 0 || a.scrollGroup[1] == 0 {
		t.Fatalf("group offsets: %v", a.scrollGroup)
	}
}

func TestEditorScrollbarKeepsMovingAcrossDragEvents(t *testing.T) {
	a := newTestApp(t)
	a.openFile("backend/probe.go")
	if !a.startEditorScrollDrag(0) {
		t.Fatal("scrollbar did not start")
	}
	a.moveEditorScrollDrag(20)
	first := a.scrollY
	a.moveEditorScrollDrag(20)
	if first <= 0 || a.scrollY <= first {
		t.Fatalf("drag offsets: first %d, second %d", first, a.scrollY)
	}
}

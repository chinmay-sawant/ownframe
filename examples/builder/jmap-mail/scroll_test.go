package main

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/pagetest"
)

func TestMailScrollbarDragPagingAndKeyboard(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	a.findMessage(a.selectedID).Body = strings.Repeat("Long reader paragraph with enough content to scroll. ", 100)
	if err := a.refresh(ctx); err != nil {
		t.Fatal(err)
	}
	if err := a.page.Testing().Tick(ctx); err != nil {
		t.Fatal(err)
	}
	thumb, ok := pagetest.BoxByID(a.page, "mail-reader-thumb")
	if !ok {
		t.Fatal("missing reader thumb")
	}
	x, y := thumb.X+thumb.W/2, thumb.Y+thumb.H/2
	claimed, err := a.page.Testing().BeginDrag(ctx, x, y)
	if err != nil || !claimed {
		t.Fatalf("begin drag: claimed=%v err=%v", claimed, err)
	}
	if err := a.page.Testing().MoveDrag(ctx, x, y+30); err != nil {
		t.Fatal(err)
	}
	if err := a.page.Testing().EndDrag(ctx); err != nil {
		t.Fatal(err)
	}
	if a.readerScroll.Offset() == 0 || a.listScroll.Offset() != 0 {
		t.Fatal("drag did not isolate reader state")
	}
	a.readerScroll.ScrollTo(0)
	if err := a.refresh(ctx); err != nil {
		t.Fatal(err)
	}
	viewport := a.readerScroll.Metrics().Viewport
	if err := pagetest.Click(ctx, a.page, "mail-reader-scrollbar-after"); err != nil {
		t.Fatal(err)
	}
	if a.readerScroll.Offset() != viewport {
		t.Fatalf("paging moved %d instead of %d", a.readerScroll.Offset(), viewport)
	}
	if a.page.Form().FocusID() != "mail-reader-thumb" {
		t.Fatal("paging did not focus scrollbar")
	}
	if err := a.page.Edit().KeyDown(ctx, "end"); err != nil {
		t.Fatal(err)
	}
	if a.readerScroll.Offset() != a.readerScroll.Metrics().MaxOffset() {
		t.Fatal("End did not reach content end")
	}
	a.settings = true
	before := a.readerScroll.Offset()
	if _, err := a.handleWheel(ctx, x, y, 0, 1); err != nil {
		t.Fatal(err)
	}
	if a.readerScroll.Offset() != before {
		t.Fatal("settings allowed background wheel")
	}
}

func TestMailRemeasuresReaderAfterBatchedSelection(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	a.findMessage("m-102").Body = strings.Repeat("The next selected message has a much longer body. ", 100)
	if err := pagetest.ClickAction(ctx, a.page, "message:m-102"); err != nil {
		t.Fatal(err)
	}
	if err := a.page.Testing().Tick(ctx); err != nil {
		t.Fatal(err)
	}
	if a.readerScroll.Metrics().MaxOffset() == 0 {
		t.Fatal("reader range used the old selected message")
	}
	if _, ok := pagetest.BoxByID(a.page, "mail-reader-thumb"); !ok {
		t.Fatal("committed selection did not publish a scrollbar")
	}
}

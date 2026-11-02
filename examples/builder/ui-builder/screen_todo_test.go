package main

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/pagetest"
	"github.com/chinmay-sawant/ownframe/ui"
)

func TestTodoRowsDeclareShippedTransitions(t *testing.T) {
	t.Parallel()
	a, err := newBuilderApp()
	if err != nil {
		t.Fatal(err)
	}
	fragment, err := ui.Render(a.todoPanel())
	if err != nil {
		t.Fatalf("render todo panel: %v", err)
	}
	for _, want := range []string{
		"transition:background-color 180ms ease",
		`id="todo-row-0"`,
		`id="todo-toggle-0"`,
	} {
		if !strings.Contains(fragment, want) {
			t.Errorf("todo panel missing %q", want)
		}
	}
}

func TestTodoToggleBindsTransition(t *testing.T) {
	t.Parallel()
	a, err := newBuilderApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.ClickAction(ctx, a.page, "tab:todo"); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.ClickAction(ctx, a.page, "todo:toggle:0"); err != nil {
		t.Fatal(err)
	}
	if !a.todos.Done[0] {
		t.Fatal("toggle did not mark task 0 done")
	}
	stats := a.page.Probe().Motion()
	if stats.TransitionsBound == 0 && stats.TransitionsStarted == 0 {
		t.Fatalf("toggle bound no transition (skip %q, error %q, noID %d, noOld %d)",
			stats.TransitionSkipReason, stats.TransitionError, stats.TransitionsNoID, stats.TransitionsNoOld)
	}
}

func TestTodoAllDoneBannerAnimates(t *testing.T) {
	t.Parallel()
	a, err := newBuilderApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.ClickAction(ctx, a.page, "tab:todo"); err != nil {
		t.Fatal(err)
	}
	for i := range a.todos.Items {
		if err := pagetest.ClickAction(ctx, a.page, fmt.Sprintf("todo:toggle:%d", i)); err != nil {
			t.Fatalf("toggle %d: %v", i, err)
		}
	}
	html := a.page.HTML()
	for _, want := range []string{
		`id="todo-all-done"`,
		"todo-done-pop",
		"animation-name:todo-done-pop",
		"0 of 3 done",
	} {
		// The progress line moves as rows complete; only the final count matters.
		if want == "0 of 3 done" {
			want = "3 of 3 done"
		}
		if !strings.Contains(html, want) {
			t.Errorf("all-done HTML missing %q", want)
		}
	}
	stats := a.page.Probe().Motion()
	if stats.Bound == 0 {
		t.Fatalf("all-done banner bound no animation (error %q)", stats.LastError)
	}
	if stats.LastError != "" {
		t.Fatalf("LastError = %q", stats.LastError)
	}
}

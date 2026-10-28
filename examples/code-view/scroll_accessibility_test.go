package main

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// scrollAccessibilityLongList is long enough that its proportional thumb falls
// below treeThumbMin, so the rail has to clamp it.
const scrollAccessibilityLongList = 240

// scrollAccessibilityMediumList still overflows its pane, but leaves a taller
// proportional thumb, so both the clamp and the ratio are covered.
const scrollAccessibilityMediumList = 48

// TestScrollAccessibilityGeometry checks the four rail boxes a pointer can hit
// and the thumb geometry each rail publishes. Both a change-heavy and a
// history-heavy fixture run, so the thumb sizes come from two different totals.
func TestScrollAccessibilityGeometry(t *testing.T) {
	cases := []struct {
		name    string
		files   int
		commits int
	}{
		{"changes-heavy", scrollAccessibilityLongList, scrollAccessibilityMediumList},
		{"history-heavy", scrollAccessibilityMediumList, scrollAccessibilityLongList},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := newScrollAccessibilityApp(t, tc.files, tc.commits)
			boxes := app.Page().Boxes()

			sidebar, ok := boxByID(boxes, "sidebar")
			if !ok {
				t.Fatal("sidebar box is missing")
			}

			statusbar, ok := boxByID(boxes, "statusbar")
			if !ok {
				t.Fatal("statusbar box is missing")
			}

			view := app.View()
			if !view.Diff.Scroll {
				t.Fatalf("changes do not scroll: files=%d visible=%d", len(app.changes.Files), app.changeVisibleRows())
			}

			if !view.History.Scroll {
				t.Fatalf("history does not scroll: commits=%d visible=%d", len(app.commits.Commits), app.historyVisibleRows())
			}

			rails := []struct {
				name    string
				railID  string
				thumbID string
			}{
				{"changes", "changes-scroll", "changes-thumb"},
				{"history", "history-scroll", "history-thumb"},
			}

			for _, railCase := range rails {
				t.Run(railCase.name, func(t *testing.T) {
					rail, ok := boxByID(boxes, railCase.railID)
					if !ok {
						t.Fatalf("%s box is missing", railCase.railID)
					}

					thumb, ok := boxByID(boxes, railCase.thumbID)
					if !ok {
						t.Fatalf("%s box is missing from Page.Boxes()", railCase.thumbID)
					}

					t.Logf("%s rail x=%.1f y=%.1f w=%.1f h=%.1f; thumb x=%.1f y=%.1f w=%.1f h=%.1f; sidebar x=%.1f..%.1f; statusbar top=%.1f",
						railCase.name, rail.X, rail.Y, rail.W, rail.H,
						thumb.X, thumb.Y, thumb.W, thumb.H,
						sidebar.X, sidebar.X+sidebar.W, statusbar.Y)

					if thumb.H < float64(treeThumbMin) {
						t.Errorf("%s height = %.1f, want at least treeThumbMin %d", railCase.thumbID, thumb.H, treeThumbMin)
					}

					if thumb.H >= rail.H {
						t.Errorf("%s height = %.1f, want less than the track %.1f", railCase.thumbID, thumb.H, rail.H)
					}

					if thumb.X < sidebar.X || thumb.X+thumb.W > sidebar.X+sidebar.W {
						t.Errorf("%s spans x=%.1f..%.1f, want it inside the sidebar x=%.1f..%.1f",
							railCase.thumbID, thumb.X, thumb.X+thumb.W, sidebar.X, sidebar.X+sidebar.W)
					}

					if thumb.Y+thumb.H > statusbar.Y {
						t.Errorf("%s bottom = %.1f, want it above the status bar top %.1f",
							railCase.thumbID, thumb.Y+thumb.H, statusbar.Y)
					}

					if thumb.Y < rail.Y || thumb.Y+thumb.H > rail.Y+rail.H {
						t.Errorf("%s spans y=%.1f..%.1f, want it inside its rail y=%.1f..%.1f",
							railCase.thumbID, thumb.Y, thumb.Y+thumb.H, rail.Y, rail.Y+rail.H)
					}
				})
			}
		})
	}
}

// TestScrollAccessibilityControls walks the accessibility tree and checks each
// scroll rail is either named for assistive technology or left out of the Tab
// order. A control that is unnamed and still in the Tab order fails with the
// node IDs and names the tree exposed.
func TestScrollAccessibilityControls(t *testing.T) {
	app := newScrollAccessibilityApp(t, scrollAccessibilityLongList, scrollAccessibilityLongList)

	tree, err := app.Page().Access().AccessibilityTree()
	if err != nil {
		t.Fatalf("accessibility tree: %v", err)
	}

	if len(tree.Nodes) == 0 {
		t.Fatal("accessibility tree is empty")
	}

	order := scrollAccessibilityTabOrder(t, app)

	scrollIDs := []string{"changes-scroll", "history-scroll"}
	found := map[string]ownframe.AccessibilityNode{}
	report := make([]string, 0, len(scrollIDs))

	walkAccessibilityNodes(tree, func(node ownframe.AccessibilityNode) {
		if !slices.Contains(scrollIDs, node.ElementID) {
			return
		}

		found[node.ElementID] = node
		report = append(report, fmt.Sprintf("%s=%q name=%q focusable=%v",
			node.ElementID, node.ID, node.Name, node.Focusable))
	})

	for _, id := range scrollIDs {
		node, ok := found[id]
		named := ok && strings.TrimSpace(node.Name) != ""
		inTabOrder := slices.Contains(order, id)

		if ok {
			t.Logf("%s: node=%q name=%q focusable=%v inTabOrder=%v", id, node.ID, node.Name, node.Focusable, inTabOrder)
		} else {
			t.Logf("%s: absent from the accessibility tree, inTabOrder=%v", id, inTabOrder)
		}

		if !named && inTabOrder {
			t.Fatalf("%s is unnamed and still in the Tab order; scroll nodes: %s",
				id, strings.Join(report, "; "))
		}
	}
}

// newScrollAccessibilityApp builds an app in changes mode with synthetic change
// and history lists, scrolls both to the middle, and redraws once.
func newScrollAccessibilityApp(t *testing.T, files, commits int) *App {
	t.Helper()

	app := newTestApp(t)
	app.mode = "changes"
	app.changes = changeSet{Branch: "feature/scroll-accessibility", Note: "Synthetic scroll accessibility fixture."}

	for i := 0; i < files; i++ {
		app.changes.Files = append(app.changes.Files, changedFile{
			Path:   fmt.Sprintf("changed/file-%02d.go", i),
			Status: "M",
			Add:    i + 1,
			Del:    i / 2,
		})
	}

	app.statuses = statusMap(app.changes)
	app.selectFirstDiff()
	app.commits = scrollAccessibilityHistory(commits)
	app.historyLabel = app.commits.Branch
	app.headBranch = app.commits.Branch
	app.historyScroll.Limit = commits

	changeRows := app.changeVisibleRows()
	if files <= changeRows {
		t.Fatalf("change fixture does not scroll: files=%d visible=%d", files, changeRows)
	}

	app.changesScroll.SetOffset((files-changeRows)/2, files, changeRows)

	historyRows := app.historyVisibleRows()
	if commits <= historyRows {
		t.Fatalf("history fixture does not scroll: commits=%d visible=%d", commits, historyRows)
	}

	app.historyScroll.SetOffset((commits-historyRows)/2, commits, historyRows)
	app.refresh()

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatalf("redraw: %v", err)
	}

	return app
}

// scrollAccessibilityHistory builds a newest-first linear history with one lane.
func scrollAccessibilityHistory(count int) commitLog {
	commits := make([]gitCommit, count)

	for i := range commits {
		hash := fmt.Sprintf("%040x", i+1)
		commits[i] = gitCommit{
			Hash:    hash,
			Short:   hash[:7],
			Subject: fmt.Sprintf("Accessibility commit %02d", i+1),
			Author:  "Test Author",
			Date:    "2026-10-27",
		}

		if i+1 < count {
			commits[i].Parents = []string{fmt.Sprintf("%040x", i+2)}
		}
	}

	return commitLog{
		Branch:  "feature/scroll-accessibility",
		Head:    commits[0].Hash,
		Commits: commits,
		Note:    "Synthetic linear history for the scroll accessibility test.",
		Sample:  true,
	}
}

// scrollAccessibilityTabOrder walks the page's real Tab order once and returns
// every control id it visits. FocusNext skips disabled controls and tabindex=-1.
func scrollAccessibilityTabOrder(t *testing.T, app *App) []string {
	t.Helper()

	ctx := context.Background()
	page := app.Page()

	if err := page.Form().Focus(ctx, ""); err != nil {
		t.Fatalf("clear focus: %v", err)
	}

	boxCount := len(page.Boxes())
	order := make([]string, 0, boxCount)

	for i := 0; i <= boxCount; i++ {
		if err := page.Form().FocusNext(ctx); err != nil {
			t.Fatalf("focus next: %v", err)
		}

		id := page.Form().FocusID()
		if id == "" {
			break
		}

		if len(order) > 0 && id == order[0] {
			break
		}

		order = append(order, id)
	}

	return order
}

// walkAccessibilityNodes visits the tree from its root in child order.
func walkAccessibilityNodes(tree ownframe.AccessibilityTree, visit func(ownframe.AccessibilityNode)) {
	byID := make(map[ownframe.AccessibilityNodeID]ownframe.AccessibilityNode, len(tree.Nodes))
	for _, node := range tree.Nodes {
		byID[node.ID] = node
	}

	visited := make(map[ownframe.AccessibilityNodeID]bool, len(tree.Nodes))

	var walk func(ownframe.AccessibilityNodeID)
	walk = func(id ownframe.AccessibilityNodeID) {
		node, ok := byID[id]
		if !ok || visited[id] {
			return
		}

		visited[id] = true
		visit(node)

		for _, child := range node.Children {
			walk(child)
		}
	}

	walk(tree.Root)
}

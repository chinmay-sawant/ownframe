package main

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"
)

// graphHash builds a stable 40-character hex hash from a small seed.
func graphHash(seed int) string {
	return fmt.Sprintf("%040x", seed)
}

// graphCommit builds one synthetic commit with parent seeds.
func graphCommit(seed int, parents ...int) gitCommit {
	commit := gitCommit{Hash: graphHash(seed)}
	for _, parent := range parents {
		commit.Parents = append(commit.Parents, graphHash(parent))
	}

	return commit
}

// graphRow returns a commit's row padded to the graph width.
func graphRow(t *testing.T, graph commitGraph, seed int) []graphCell {
	t.Helper()

	row, ok := graph.Rows[graphHash(seed)]
	if !ok {
		t.Fatalf("commit %x has no graph row", seed)
	}

	return graphCells(row, graph.Lanes)
}

func graphKindAt(row []graphCell, lane int) string {
	if lane < 0 || lane >= len(row) {
		return graphSpace
	}

	return row[lane].Kind
}

func TestCommitParentsParsed(t *testing.T) {
	t.Run("record", func(t *testing.T) {
		merge := graphHash(2)
		first := graphHash(3)
		second := graphHash(4)
		short := graphHash(5)

		record := "\x1e" + merge + "\x00" + short[:7] + "\x00" +
			first + " " + second + " abc1234 " + strings.Repeat("z", commitHashLen) + "\x00" +
			"Ada Lovelace\x00" + "2026-10-01T10:00:00+02:00\x00" +
			"Merge the probe branch\x00" + "Probe retries land.\x00" +
			"\nM\tbackend/probe.go\n"

		commits := parseCommitLog(record)
		if len(commits) != 1 {
			t.Fatalf("parsed %d commits, want 1", len(commits))
		}

		commit := commits[0]
		if !slices.Equal(commit.Parents, []string{first, second}) {
			t.Fatalf("parents = %v, want %v", commit.Parents, []string{first, second})
		}

		if commit.Subject != "Merge the probe branch" {
			t.Fatalf("subject = %q", commit.Subject)
		}

		if len(commit.Files) != 1 || commit.Files[0].Path != "backend/probe.go" {
			t.Fatalf("files = %+v", commit.Files)
		}

		if parents := parseCommitParents("   "); parents != nil {
			t.Fatalf("empty parents = %v, want nil", parents)
		}
	})

	t.Run("git log", func(t *testing.T) {
		dir, hashes := historyRepo(t)

		log, err := loadCommitLogLimit(context.Background(), dir, "", 2)
		if err != nil {
			t.Fatal(err)
		}

		if len(log.Commits) != 2 {
			t.Fatalf("commits = %d, want 2", len(log.Commits))
		}

		want := [][]string{{hashes[1]}, {hashes[2]}}
		for i, commit := range log.Commits {
			if !slices.Equal(commit.Parents, want[i]) {
				t.Fatalf("commit %d parents = %v, want %v", i, commit.Parents, want[i])
			}
		}

		full, err := loadCommitLog(context.Background(), dir, "")
		if err != nil {
			t.Fatal(err)
		}

		if len(full.Commits) != len(hashes) {
			t.Fatalf("commits = %d, want %d", len(full.Commits), len(hashes))
		}

		if root := full.Commits[len(full.Commits)-1]; len(root.Parents) != 0 {
			t.Fatalf("root parents = %v, want none", root.Parents)
		}
	})
}

func TestCommitGraphFirstParentSameLane(t *testing.T) {
	graph := buildCommitGraph([]gitCommit{
		graphCommit(3, 2),
		graphCommit(2, 1),
		graphCommit(1),
	})

	if graph.Lanes != 1 {
		t.Fatalf("lanes = %d, want 1", graph.Lanes)
	}

	for seed := 1; seed <= 3; seed++ {
		row := graphRow(t, graph, seed)
		if len(row) != 1 {
			t.Fatalf("commit %x row = %v, want one cell", seed, row)
		}

		if row[0].Kind != graphNode || row[0].Lane != 0 {
			t.Fatalf("commit %x cell = %+v, want node in lane 0", seed, row[0])
		}
	}
}

func TestCommitGraphBranchLane(t *testing.T) {
	graph := buildCommitGraph([]gitCommit{
		graphCommit(6, 5, 4),
		graphCommit(5, 2),
		graphCommit(4, 3),
		graphCommit(3, 1),
		graphCommit(2, 1),
		graphCommit(1),
	})

	if graph.Lanes != 2 {
		t.Fatalf("lanes = %d, want 2", graph.Lanes)
	}

	mergeRow := graphRow(t, graph, 6)
	if got := graphKindAt(mergeRow, 0); got != graphNode {
		t.Fatalf("merge node = %q, want node", got)
	}

	if got := graphKindAt(mergeRow, 1); got != graphBranch {
		t.Fatalf("extra parent lane = %q, want branch", got)
	}

	branchRow := graphRow(t, graph, 5)
	if got := graphKindAt(branchRow, 0); got != graphNode {
		t.Fatalf("first parent node = %q, want node", got)
	}

	if got := graphKindAt(branchRow, 1); got != graphVert {
		t.Fatalf("open lane = %q, want vert", got)
	}

	sideRow := graphRow(t, graph, 4)
	if got := graphKindAt(sideRow, 1); got != graphNode {
		t.Fatalf("side commit = %q, want node in lane 1", got)
	}

	if got := graphKindAt(sideRow, 0); got != graphVert {
		t.Fatalf("mainline lane = %q, want vert", got)
	}

	if got := graphKindAt(graphRow(t, graph, 1), 1); got != graphMerge {
		t.Fatalf("shared parent lane = %q, want merge", got)
	}
}

func TestCommitGraphEdgesFollowParents(t *testing.T) {
	graph := buildCommitGraph([]gitCommit{
		graphCommit(6, 5, 4), graphCommit(5, 2), graphCommit(4, 3),
		graphCommit(3, 1), graphCommit(2, 1), graphCommit(1),
	})
	want := map[int][]graphEdge{
		6: {graphLine(0, 21, 42), graphCurve(0, 1, false)},
		5: {graphLine(0, 0, 21), graphLine(1, 0, 21), graphLine(0, 21, 42), graphLine(1, 21, 42)},
		1: {graphLine(0, 0, 21), graphCurve(1, 0, true)},
	}
	for commit, edges := range want {
		if got := graph.Edges[graphHash(commit)]; !slices.Equal(got, edges) {
			t.Errorf("commit %d edges = %+v, want %+v", commit, got, edges)
		}
	}
}

func TestCommitGraphMaxLanes(t *testing.T) {
	commits := []gitCommit{
		graphCommit(100, 99),
		graphCommit(99, 1, 2, 3, 4, 5, 6, 7, 8),
	}

	for seed := 1; seed <= 8; seed++ {
		commits = append(commits, graphCommit(seed))
	}

	graph := buildCommitGraph(commits)

	if graph.Lanes != maxCommitLanes {
		t.Fatalf("lanes = %d, want %d", graph.Lanes, maxCommitLanes)
	}

	if len(graph.Rows) != len(commits) {
		t.Fatalf("rows = %d, want %d", len(graph.Rows), len(commits))
	}

	for _, commit := range commits {
		row, ok := graph.Rows[commit.Hash]
		if !ok {
			t.Fatalf("commit %s has no row", commit.Hash)
		}

		padded := graphCells(row, graph.Lanes)
		if len(padded) != maxCommitLanes {
			t.Fatalf("commit %s row = %d lanes, want %d", commit.Hash, len(padded), maxCommitLanes)
		}

		nodes := 0

		for lane, cell := range padded {
			if cell.Lane != lane {
				t.Fatalf("commit %s cell %d lane = %d", commit.Hash, lane, cell.Lane)
			}

			if cell.Kind == graphNode {
				nodes++
			}
		}

		if nodes != 1 {
			t.Fatalf("commit %s nodes = %d, want 1", commit.Hash, nodes)
		}
	}

	// The newest row predates the octopus merge, so it pads out to the final
	// width instead of carrying six cells of its own.
	top := graphRow(t, graph, 100)
	if got := graphKindAt(top, 0); got != graphNode {
		t.Fatalf("top commit = %q, want node", got)
	}

	for lane := 1; lane < maxCommitLanes; lane++ {
		if got := graphKindAt(top, lane); got != graphSpace {
			t.Fatalf("top commit lane %d = %q, want space", lane, got)
		}
	}

	again := buildCommitGraph(commits)
	if again.Lanes != graph.Lanes || len(again.Rows) != len(graph.Rows) {
		t.Fatalf("rebuild = %+v, want %+v", again, graph)
	}

	for hash, row := range graph.Rows {
		if !slices.Equal(row, again.Rows[hash]) {
			t.Fatalf("rebuild row %s = %v, want %v", hash, again.Rows[hash], row)
		}
	}
}

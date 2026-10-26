package main

import "testing"

func branchFilterRows() []branchRow {
	return []branchRow{
		{Name: "origin/main", Ref: "refs/remotes/origin/main", Remote: true},
		{Name: "zeta", Ref: "refs/heads/zeta", Local: true},
		{Name: "main", Ref: "refs/heads/main", Local: true, CheckedOut: true},
		{Name: "alpha", Ref: "refs/heads/alpha", Local: true},
		{Name: "upstream", Ref: "refs/remotes/team/upstream", Remote: true},
	}
}

func branchNames(rows []branchRow) []string {
	names := make([]string, 0, len(rows))

	for _, row := range rows {
		names = append(names, row.Name)
	}

	return names
}

func TestFilterBranchesOrderAndLimit(t *testing.T) {
	visible, more := filterBranches(branchFilterRows(), "", 0)
	if more != 0 {
		t.Fatalf("more = %d", more)
	}

	want := []string{"main", "zeta", "alpha", "origin/main", "upstream"}
	got := branchNames(visible)

	if len(got) != len(want) {
		t.Fatalf("visible = %v, want %v", got, want)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("visible = %v, want %v", got, want)
		}
	}
}

func TestFilterBranchesQueryMatchesNameAndRef(t *testing.T) {
	visible, more := filterBranches(branchFilterRows(), "  TEAM/ ", 0)
	if len(visible) != 1 || visible[0].Name != "upstream" || more != 0 {
		t.Fatalf("ref-only match = %v more %d", branchNames(visible), more)
	}

	visible, _ = filterBranches(branchFilterRows(), "MAIN", 0)
	if len(visible) != 2 || visible[0].Name != "main" || visible[1].Name != "origin/main" {
		t.Fatalf("name match = %v", branchNames(visible))
	}

	visible, more = filterBranches(branchFilterRows(), "nothing-here", 0)
	if len(visible) != 0 || more != 0 {
		t.Fatalf("no matches = %v more %d", branchNames(visible), more)
	}
}

func TestFilterBranchesTrimsToLimit(t *testing.T) {
	visible, more := filterBranches(branchFilterRows(), "", 2)
	if len(visible) != 2 || more != 3 {
		t.Fatalf("limit 2 = %v more %d", branchNames(visible), more)
	}

	if visible[0].Name != "main" || visible[1].Name != "zeta" {
		t.Fatalf("limit 2 rows = %v", branchNames(visible))
	}
}

func TestFilterBranchesDefaultLimit(t *testing.T) {
	rows := make([]branchRow, 0, branchDefaultLimit+5)
	for i := 0; i < branchDefaultLimit+5; i++ {
		name := string(rune('a' + i%26))
		rows = append(rows, branchRow{Name: name, Ref: "refs/heads/" + name, Local: true})
	}

	for _, limit := range []int{0, -4} {
		visible, more := filterBranches(rows, "", limit)
		if len(visible) != branchDefaultLimit || more != 5 {
			t.Fatalf("limit %d = %d rows, more %d", limit, len(visible), more)
		}
	}
}

func TestFilterBranchesKeepsInputOrder(t *testing.T) {
	rows := []branchRow{
		{Name: "delta", Ref: "refs/heads/delta", Local: true},
		{Name: "charlie", Ref: "refs/heads/charlie", Local: true},
		{Name: "bravo", Ref: "refs/heads/bravo", Local: true},
	}

	visible, _ := filterBranches(rows, "", 0)
	got := branchNames(visible)

	for i, want := range []string{"delta", "charlie", "bravo"} {
		if got[i] != want {
			t.Fatalf("visible = %v", got)
		}
	}
}

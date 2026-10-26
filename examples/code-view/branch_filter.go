package main

import (
	"sort"
	"strings"
)

const (
	// branchDefaultLimit is the picker cap when the caller passes no limit.
	branchDefaultLimit = 20

	branchRankCheckedOut = 0
	branchRankLocal      = 1
	branchRankRemote     = 2
	branchRankOther      = 3
)

// filterBranches narrows the branch rows to a case-insensitive substring of the
// short name or the full ref. A checked-out row sorts first, then locals, then
// remotes, and each group keeps the order loadBranches gave it. A limit of zero
// or less shows branchDefaultLimit rows; more counts the matching rows left out.
func filterBranches(rows []branchRow, query string, limit int) (visible []branchRow, more int) {
	if limit <= 0 {
		limit = branchDefaultLimit
	}

	needle := strings.ToLower(strings.TrimSpace(query))
	matches := make([]branchRow, 0, len(rows))

	for _, row := range rows {
		if needle != "" &&
			!strings.Contains(strings.ToLower(row.Name), needle) &&
			!strings.Contains(strings.ToLower(row.Ref), needle) {
			continue
		}

		matches = append(matches, row)
	}

	sort.SliceStable(matches, func(i, j int) bool {
		return branchRank(matches[i]) < branchRank(matches[j])
	})

	if len(matches) > limit {
		return matches[:limit], len(matches) - limit
	}

	return matches, 0
}

// branchRank orders the picker: the checked-out branch, then locals, then
// remotes, then anything else.
func branchRank(row branchRow) int {
	switch {
	case row.CheckedOut:
		return branchRankCheckedOut
	case row.Local:
		return branchRankLocal
	case row.Remote:
		return branchRankRemote
	default:
		return branchRankOther
	}
}

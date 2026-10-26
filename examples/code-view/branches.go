package main

import (
	"context"
	"sort"
	"strings"
)

const branchSeed = 16

// branchRow is one local or remote ref in the history picker.
type branchRow struct {
	Name       string
	Ref        string
	Hash       string
	Upstream   string
	Track      string
	Local      bool
	Remote     bool
	CheckedOut bool
	Unborn     bool
}

// loadBranches lists local and remote refs without changing the checkout.
func loadBranches(ctx context.Context, root string) ([]branchRow, string, error) {
	ctx, cancel := context.WithTimeout(ctx, gitTimeout)
	defer cancel()

	raw, err := gitBytes(ctx, root,
		"for-each-ref",
		"--format=%(refname)%00%(refname:short)%00%(objectname)%00%(HEAD)%00%(upstream:short)%00%(upstream:track)%00%(symref)",
		"refs/heads", "refs/remotes",
	)
	if err != nil {
		return nil, "", err
	}

	rows := make([]branchRow, 0, branchSeed)

	for _, line := range strings.Split(string(raw), "\n") {
		fields := strings.Split(line, "\x00")
		if len(fields) < 7 || fields[0] == "" || fields[6] != "" {
			continue
		}

		row := branchRow{
			Ref:      fields[0],
			Name:     fields[1],
			Hash:     fields[2],
			Upstream: fields[4],
			Track:    fields[5],
			Local:    strings.HasPrefix(fields[0], "refs/heads/"),
			Remote:   strings.HasPrefix(fields[0], "refs/remotes/"),
		}

		rows = append(rows, row)
	}

	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Local != rows[j].Local {
			return rows[i].Local
		}

		return rows[i].Name < rows[j].Name
	})

	head, err := gitText(ctx, root, "symbolic-ref", "--quiet", "--short", "HEAD")
	if err == nil && head != "" {
		for i := range rows {
			if rows[i].Local && rows[i].Name == head {
				rows[i].CheckedOut = true
			}
		}

		sort.SliceStable(rows, func(i, j int) bool {
			if rows[i].CheckedOut != rows[j].CheckedOut {
				return rows[i].CheckedOut
			}

			return rows[i].Name < rows[j].Name
		})

		return rows, head, nil
	}

	short, shortErr := gitText(ctx, root, "rev-parse", "--short", "HEAD")
	if shortErr != nil || short == "" {
		return rows, "HEAD", nil
	}

	label := "HEAD detached at " + short
	for i := range rows {
		rows[i].CheckedOut = false
	}

	return rows, label, nil
}

// branchByRef finds a branch row by full ref.
func branchByRef(rows []branchRow, ref string) (branchRow, bool) {
	for _, row := range rows {
		if row.Ref == ref {
			return row, true
		}
	}

	return branchRow{}, false
}

// branchLabel is the short display name for a full ref.
func branchLabel(ref string) string {
	ref = strings.TrimPrefix(ref, "refs/heads/")
	ref = strings.TrimPrefix(ref, "refs/remotes/")

	return ref
}

package main

import (
	"context"
	"fmt"
)

func defaultBranchRef(ctx context.Context, root string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, gitTimeout)
	defer cancel()
	ref, err := gitText(ctx, root, "symbolic-ref", "--quiet", "refs/remotes/origin/HEAD")
	if err == nil && ref != "" {
		return ref, nil
	}
	var found []string
	for _, candidate := range []string{"refs/heads/main", "refs/heads/master"} {
		if _, err := gitText(ctx, root, "show-ref", "--verify", "--quiet", candidate); err == nil {
			found = append(found, candidate)
		}
	}
	if len(found) == 1 {
		return found[0], nil
	}
	return "", fmt.Errorf("cannot determine the default branch: set origin/HEAD or keep one local main or master branch")
}

// loadBranchChanges compares the branch tip with its merge base against base.
func loadBranchChanges(ctx context.Context, root, base, branch string) (changeSet, error) {
	ctx, cancel := context.WithTimeout(ctx, gitTimeout)
	defer cancel()

	baseCommit, err := gitText(ctx, root, "rev-parse", "--verify", "--end-of-options", base+"^{commit}")
	if err != nil {
		return changeSet{}, fmt.Errorf("resolve base %s: %w", base, err)
	}
	tip, err := gitText(ctx, root, "rev-parse", "--verify", "--end-of-options", branch+"^{commit}")
	if err != nil {
		return changeSet{}, fmt.Errorf("resolve branch %s: %w", branch, err)
	}
	ancestor, err := gitText(ctx, root, "merge-base", baseCommit, tip)
	if err != nil {
		return changeSet{}, fmt.Errorf("find common ancestor: %w", err)
	}
	raw, err := gitBytes(ctx, root, "diff", "--no-color", "--no-ext-diff", ancestor, tip, "--")
	if err != nil {
		return changeSet{}, err
	}
	changes := changeSet{Branch: branchLabel(branch), Files: parseDiff(string(raw))}
	if len(changes.Files) == 0 {
		changes.EmptyTitle = "No branch changes"
	}
	changes.Note = fmt.Sprintf("Committed changes in %s since its common ancestor with %s.",
		branchLabel(branch), branchLabel(base))
	return changes, nil
}

package main

import (
	"fmt"
	"strings"
)

// noChanges is the label a clean working tree shows.
const noChanges = "No changes"

// changeSummary counts a change set by status and by line. The status bar and
// the floating file navigator read it instead of walking the file list.
type changeSummary struct {
	Added     int
	Modified  int
	Deleted   int
	Renamed   int
	Untracked int
	Other     int
	Total     int
	Add       int
	Del       int
}

// summarizeChanges buckets every file in cs once by its status letter and sums
// the added and deleted lines.
func summarizeChanges(cs changeSet) changeSummary {
	var summary changeSummary

	for _, file := range cs.Files {
		summary.Total++

		switch strings.ToUpper(file.Status) {
		case "A":
			summary.Added++
		case "M":
			summary.Modified++
		case "D":
			summary.Deleted++
		case "R":
			summary.Renamed++
		case "U":
			summary.Untracked++
		default:
			summary.Other++
		}
	}

	summary.Add, summary.Del = cs.totals()

	return summary
}

// Label is the long form shown in the status bar, such as
// "7 changed · 3 modified · +562 −109". Zero buckets are left out and a clean
// tree reads "No changes". The line totals use a minus sign, like the diff
// stats in the editor.
func (s changeSummary) Label() string {
	if s.Total == 0 {
		return noChanges
	}

	var parts []string
	if s.Total == 1 {
		parts = append(parts, "1 change")
	} else {
		parts = append(parts, fmt.Sprintf("%d changed", s.Total))
	}

	parts = appendCount(parts, s.Modified, "modified")
	parts = appendCount(parts, s.Added, "added")
	parts = appendCount(parts, s.Deleted, "deleted")
	parts = appendCount(parts, s.Renamed, "renamed")
	parts = appendCount(parts, s.Untracked, "untracked")
	parts = appendCount(parts, s.Other, "other")

	return strings.Join(parts, " · ") + fmt.Sprintf(" · +%d −%d", s.Add, s.Del)
}

// Short is the compact form shown beside the file navigator, such as
// "3M 2A 1D". Zero buckets are left out and a clean tree is empty.
func (s changeSummary) Short() string {
	var parts []string

	for _, bucket := range []struct {
		count int
		code  byte
	}{
		{s.Modified, 'M'},
		{s.Added, 'A'},
		{s.Deleted, 'D'},
		{s.Renamed, 'R'},
		{s.Untracked, 'U'},
		{s.Other, '?'},
	} {
		if bucket.count > 0 {
			parts = append(parts, fmt.Sprintf("%d%c", bucket.count, bucket.code))
		}
	}

	return strings.Join(parts, " ")
}

// appendCount adds a "N word" part unless the bucket is empty.
func appendCount(parts []string, count int, word string) []string {
	if count == 0 {
		return parts
	}

	return append(parts, fmt.Sprintf("%d %s", count, word))
}

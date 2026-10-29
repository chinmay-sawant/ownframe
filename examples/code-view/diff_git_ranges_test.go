package main

import (
	"strings"
	"testing"
)

func TestGitHunkRangesBothViews(t *testing.T) {
	for _, tc := range []struct{ name, before, after string }{
		{"replace", "one\ntwo\nthree\n", "one\nTWO\nthree\n"},
		{"insert", "one\ntwo\n", "zero\none\ntwo\nthree\n"},
		{"delete", "one\ntwo\nthree\n", "two\n"},
		{"empty before", "", "one\ntwo\n"},
		{"empty after", "one\ntwo\n", ""},
		{"no final newline", "one\ntwo", "one\nTWO"},
		{"separate hunks", "1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12\n", "ONE\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\nTWELVE\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir, _ := historyRepo(t)
			writeFile(t, dir, "ranges.txt", tc.before)
			runGit(t, dir, "add", "ranges.txt")
			runGit(t, dir, "commit", "-qm", "range base")
			writeFile(t, dir, "ranges.txt", tc.after)
			for _, contextLines := range []string{"-U0", "-U3"} {
				raw := runGit(t, dir, "diff", "HEAD", "--no-color", "--no-ext-diff", contextLines, "--", "ranges.txt")
				files := parseDiff(raw)
				if len(files) != 1 {
					t.Fatalf("%s: files=%d", contextLines, len(files))
				}
				headers := []string{}
				for _, line := range strings.Split(raw, "\n") {
					if strings.HasPrefix(line, "@@ ") {
						headers = append(headers, line)
					}
				}
				file := files[0]
				if len(file.Hunks) != len(headers) {
					t.Fatal("hunk count differs from Git")
				}
				for i, h := range file.Hunks {
					if got := hunkLabel(h); got != headers[i] {
						t.Fatalf("header=%q Git=%q", got, headers[i])
					}
					oldCount, newCount := 0, 0
					oldNext, newNext := h.OldStart, h.NewStart
					for _, line := range h.Lines {
						if line.Kind != '+' {
							if line.OldNo != oldNext {
								t.Fatal("old line offset mismatch")
							}
							oldNext++
							oldCount++
						}
						if line.Kind != '-' {
							if line.NewNo != newNext {
								t.Fatal("new line offset mismatch")
							}
							newNext++
							newCount++
						}
					}
					if oldCount != h.OldCount || newCount != h.NewCount {
						t.Fatalf("counts=%d,%d Git=%d,%d", oldCount, newCount, h.OldCount, h.NewCount)
					}
				}
				for _, rows := range [][]diffRow{diffRowsUnified(file, "text"), diffRowsSplit(file, "text")} {
					got := []string{}
					for _, row := range rows {
						if row.Kind == rowHunk {
							got = append(got, row.Hunk)
						}
					}
					if strings.Join(got, "\n") != strings.Join(headers, "\n") {
						t.Fatalf("view headers=%v Git=%v", got, headers)
					}
				}
			}
		})
	}
}

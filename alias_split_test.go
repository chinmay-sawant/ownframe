package ownframe

import "strings"

func splitAliasLines(s string) []string {
	var out []string
	for _, n := range strings.Split(s, "\n") {
		n = strings.TrimSpace(n)
		if n != "" {
			out = append(out, n)
		}
	}
	return out
}

func aliasFiles(allow map[string]bool) map[string]bool {
	files := make(map[string]bool)
	for k := range allow {
		file, _, ok := strings.Cut(k, ":")
		if ok {
			files[file] = true
		}
	}
	return files
}

package page

import "strings"

func splitComma(s string) []string {
	var out []string
	for _, n := range strings.Split(s, ",") {
		n = strings.TrimSpace(n)
		if n != "" {
			out = append(out, n)
		}
	}
	return out
}

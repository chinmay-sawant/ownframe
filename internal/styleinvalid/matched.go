package styleinvalid

import "strings"

func touchedSet(touched any) map[string]bool {
	set := map[string]bool{}
	switch v := touched.(type) {
	case nil:
	case string:
		set[normalizeKey(v)] = true
	case []string:
		for _, s := range v {
			set[normalizeKey(s)] = true
		}
	case []Change:
		for _, c := range v {
			if c.BoxID != "" {
				set[normalizeKey(c.BoxID)] = true
			}
			if c.Path != "" {
				set[normalizeKey(c.Path)] = true
			}
		}
	case map[string]bool:
		for k, ok := range v {
			if ok {
				set[normalizeKey(k)] = true
			}
		}
	case map[string]struct{}:
		for k := range v {
			set[normalizeKey(k)] = true
		}
	}
	return set
}

func normalizeKey(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "#")
	s = strings.TrimPrefix(s, ".")
	return s
}

func skipUnmodifiedSelectors(selectors []string, touched any) []string {
	set := touchedSet(touched)
	var out []string
	for _, sel := range selectors {
		if set[normalizeKey(sel)] {
			out = append(out, sel)
		}
	}
	if out == nil {
		out = []string{}
	}
	return out
}

// SkipUnmodifiedSelectors keeps only selectors affected by the touched set.
func SkipUnmodifiedSelectors(selectors []string, touched any) []string {
	return skipUnmodifiedSelectors(selectors, touched)
}

// SkipUnmodified is an alias for SkipUnmodifiedSelectors.
func SkipUnmodified(selectors []string, touched any) []string {
	return skipUnmodifiedSelectors(selectors, touched)
}

// SkipUnmodifiedMatchedSelectors is an alias.
func SkipUnmodifiedMatchedSelectors(selectors []string, touched any) []string {
	return skipUnmodifiedSelectors(selectors, touched)
}

// SelectorsModified is true when at least one selector is touched.
func SelectorsModified(selectors []string, touched any) bool {
	return len(skipUnmodifiedSelectors(selectors, touched)) > 0
}

package benchscore

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
)

// Class is the budget family a benchmark belongs to.
type Class string

const (
	ClassIdle        Class = "Idle"
	ClassInteraction Class = "Interaction"
	ClassRebuild     Class = "Rebuild"
	ClassResize      Class = "Resize"
	ClassTyping      Class = "Typing"
	ClassHeap        Class = "Heap"
	ClassOther       Class = "Other"
)

// Manifest maps a benchmark name, base name, or full sub-benchmark name to a
// class. It overrides the built-in name rules. Accepted values are idle,
// interaction (or click), rebuild, resize, typing, heap, other, and same-data.
type Manifest map[string]string

// ParseManifest decodes a JSON object such as
// {"BenchmarkInvoiceColdBuild":"rebuild"} and validates every class value.
func ParseManifest(r io.Reader) (Manifest, error) {
	var raw map[string]string
	if err := json.NewDecoder(r).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode manifest: %w", err)
	}
	manifest := make(Manifest, len(raw))
	for name, value := range raw {
		name = strings.TrimSpace(name)
		if name == "" {
			return nil, errors.New("manifest: benchmark name must not be empty")
		}
		if _, _, ok := parseClassValue(value); !ok {
			return nil, fmt.Errorf("manifest: benchmark %q: unknown class %q", name, value)
		}
		manifest[name] = strings.ToLower(strings.TrimSpace(value))
	}
	return manifest, nil
}

// LoadManifest reads a manifest from disk. An empty path returns no manifest.
func LoadManifest(path string) (Manifest, error) {
	if strings.TrimSpace(path) == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path) //nolint:gosec // reason: the path is an explicit CLI input.
	if err != nil {
		return nil, fmt.Errorf("read manifest %q: %w", path, err)
	}
	manifest, err := ParseManifest(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return manifest, nil
}

// Classify returns the class for a benchmark name and whether the row must
// hold the stricter same-data budget. A manifest entry wins over the name
// rules.
func Classify(name string, manifest Manifest) (Class, bool) {
	if raw, ok := manifest.lookup(name); ok {
		if class, sameData, valid := parseClassValue(raw); valid {
			return class, sameData
		}
	}
	return classifyName(name)
}

func (m Manifest) lookup(name string) (string, bool) {
	for _, candidate := range manifestNames(name) {
		if value, ok := m[candidate]; ok {
			return value, true
		}
	}
	return "", false
}

func manifestNames(name string) []string {
	var names []string
	add := func(candidate string) {
		for _, existing := range names {
			if existing == candidate {
				return
			}
		}
		names = append(names, candidate)
	}
	for _, candidate := range []string{name, strings.TrimPrefix(name, benchmarkPrefix)} {
		add(candidate)
		add(stripGOMAXPROCS(candidate))
		base := candidate
		if slash := strings.IndexByte(base, '/'); slash >= 0 {
			base = base[:slash]
		}
		add(base)
		add(stripGOMAXPROCS(base))
	}
	return names
}

func stripGOMAXPROCS(name string) string {
	dash := strings.LastIndexByte(name, '-')
	if dash < 0 || dash == len(name)-1 {
		return name
	}
	for _, r := range name[dash+1:] {
		if r < '0' || r > '9' {
			return name
		}
	}
	return name[:dash]
}

func parseClassValue(raw string) (Class, bool, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "idle":
		return ClassIdle, false, true
	case "same-data", "samedata", "same_data":
		return ClassIdle, true, true
	case "interaction", "click", "tap":
		return ClassInteraction, false, true
	case "rebuild", "build":
		return ClassRebuild, false, true
	case "resize":
		return ClassResize, false, true
	case "typing":
		return ClassTyping, false, true
	case "heap", "retained":
		return ClassHeap, false, true
	case "other":
		return ClassOther, false, true
	default:
		return "", false, false
	}
}

func classifyName(name string) (Class, bool) {
	tokens := nameTokens(name)
	switch {
	case hasToken(tokens, "idle"):
		return ClassIdle, false
	case hasToken(tokens, "samedata") || hasSequence(tokens, "same", "data"):
		return ClassIdle, true
	case hasToken(tokens, "heap", "retained", "retention", "allocs", "allocation"):
		return ClassHeap, false
	case hasToken(tokens, "click", "tap", "interaction"):
		return ClassInteraction, false
	case hasToken(tokens, "rebuild", "relayout", "reflow") || hasSequence(tokens, "cold", "build"):
		return ClassRebuild, false
	case hasToken(tokens, "resize", "resizing"):
		return ClassResize, false
	case hasToken(tokens, "typing", "keystroke", "keypress"):
		return ClassTyping, false
	default:
		return ClassOther, false
	}
}

func nameTokens(name string) []string {
	var tokens []string
	var current []rune
	flush := func() {
		if len(current) == 0 {
			return
		}
		tokens = append(tokens, string(current))
		current = current[:0]
	}
	var previous rune
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			flush()
			previous = 0
			continue
		}
		if len(current) > 0 && unicode.IsUpper(r) && (unicode.IsLower(previous) || unicode.IsDigit(previous)) {
			flush()
		}
		current = append(current, unicode.ToLower(r))
		previous = r
	}
	flush()
	return tokens
}

func hasToken(tokens []string, wanted ...string) bool {
	for _, token := range tokens {
		for _, want := range wanted {
			if token == want {
				return true
			}
		}
	}
	return false
}

func hasSequence(tokens []string, wanted ...string) bool {
	if len(wanted) == 0 || len(tokens) < len(wanted) {
		return false
	}
	for start := 0; start+len(wanted) <= len(tokens); start++ {
		matched := true
		for offset, want := range wanted {
			if tokens[start+offset] != want {
				matched = false
				break
			}
		}
		if matched {
			return true
		}
	}
	return false
}

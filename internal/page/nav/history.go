// Package nav holds template history and named HTML routes for package
// page. It does not import page.
package nav

import "strings"

// FNV-1a 64-bit parameters, matching hash/fnv New64a. Hash folds them by
// hand so one pass reports the hash with no heap state.
const (
	fnvOffset64 = 14695981039346656037
	fnvPrime64  = 1099511628211
)

// Hash returns the FNV-1a hash of s.
func Hash(s string) uint64 {
	hash := uint64(fnvOffset64)
	for i := 0; i < len(s); i++ {
		hash ^= uint64(s[i])
		hash *= fnvPrime64
	}
	return hash
}

// Entry records a source and how to render it. The delimiter verdict is
// stored so that revisiting the entry does not require another scan.
type Entry struct {
	HTML    string
	Hash    uint64
	Static  bool
	Execute bool
}

// Source prepares an entry for either literal HTML or a Go template.
func Source(html string, static bool) Entry {
	return Entry{
		HTML:    html,
		Hash:    Hash(html),
		Static:  static,
		Execute: !static && strings.Contains(html, "{{"),
	}
}

// SameSource compares the cheap mode and hash checks before the exact bytes.
func (s Entry) SameSource(other Entry) bool {
	return s.Static == other.Static && s.Hash == other.Hash && s.HTML == other.HTML
}

// History owns the visited sources and the named HTML routes.
type History struct {
	past   []Entry
	at     int
	routes map[string]string
}

// New starts history at source.
func New(source Entry) History { return History{past: []Entry{source}} }

// Current is the history entry at the visible index.
func (h *History) Current() Entry { return h.past[h.at] }

// Doc is the visible history index.
func (h *History) Doc() int { return h.at }

// Push records a source and drops any later entries.
func (h *History) Push(source Entry) {
	h.past = append(h.past[:h.at+1], source)
	h.at = len(h.past) - 1
}

// GoTo sets the visible index.
func (h *History) GoTo(at int) { h.at = at }

// Retarget replaces the current entry without moving the index.
func (h *History) Retarget(source Entry) { h.past[h.at] = source }

// At returns the entry at the given index.
func (h *History) At(at int) Entry { return h.past[at] }

// CanBack reports whether a previous entry exists.
func (h *History) CanBack() bool { return h.at > 0 }

// CanForward reports whether a later entry exists.
func (h *History) CanForward() bool { return h.at+1 < len(h.past) }

// BackIndex is the previous history index.
func (h *History) BackIndex() int { return h.at - 1 }

// ForwardIndex is the next history index.
func (h *History) ForwardIndex() int { return h.at + 1 }

// Route maps a data-action value to HTML for Click to load.
func (h *History) Route(action, html string) {
	if h.routes == nil {
		h.routes = map[string]string{}
	}
	h.routes[action] = html
}

// Lookup returns the HTML registered for action.
func (h *History) Lookup(action string) (string, bool) {
	html, ok := h.routes[action]
	return html, ok
}

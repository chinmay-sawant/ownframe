// Package nav holds template history and named HTML routes for package
// page. It does not import page.
package nav

// FNV-1a 64-bit parameters, matching hash/fnv New64a. Scan folds them by
// hand so one pass reports the hash and the template delimiter together
// with no heap state.
const (
	fnvOffset64 = 14695981039346656037
	fnvPrime64  = 1099511628211
)

// Scan returns the FNV-1a hash of s and whether s holds a Go template
// delimiter. History stores both beside each source, so the nav seam
// compares stored hashes instead of rehashing both sides and the template
// probe reuses the stored verdict instead of rescanning.
func Scan(s string) (uint64, bool) {
	hash := uint64(fnvOffset64)
	delim := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '{' && i+1 < len(s) && s[i+1] == '{' {
			delim = true
		}
		hash ^= uint64(c)
		hash *= fnvPrime64
	}
	return hash, delim
}

// Hash returns the FNV-1a hash of s: the same value Scan reports. Prefer
// Scan when the delimiter verdict is also needed.
func Hash(s string) uint64 {
	hash, _ := Scan(s)
	return hash
}

// History owns template history and routes behind narrow methods.
// past holds every loaded source; hashes and delims hold the Scan verdicts
// stored when each source was recorded, so identical-source checks compare
// without rehashing and the template probe never rescans. at is the visible
// index. routes maps a data-action value to HTML for Click to load.
type History struct {
	past   []string
	hashes []uint64
	delims []bool
	at     int
	routes map[string]string
}

// New starts history at source.
func New(source string) History {
	hash, delim := Scan(source)
	return History{past: []string{source}, hashes: []uint64{hash}, delims: []bool{delim}}
}

// Current is the template source at the visible index.
func (h *History) Current() string { return h.past[h.at] }

// CurrentHash is the stored hash of the visible source: comparing it never
// rehashes the stored side.
func (h *History) CurrentHash() uint64 { return h.hashes[h.at] }

// CurrentDelim reports whether the visible source holds a template
// delimiter, as stored when the source was recorded.
func (h *History) CurrentDelim() bool { return h.delims[h.at] }

// Doc is the visible history index.
func (h *History) Doc() int { return h.at }

// Push records html and drops any later entries. The Scan verdicts ride
// beside the source so later compares never rehash it.
func (h *History) Push(html string) {
	hash, delim := Scan(html)
	h.past = append(h.past[:h.at+1], html)
	h.hashes = append(h.hashes[:h.at+1], hash)
	h.delims = append(h.delims[:h.at+1], delim)
	h.at = len(h.past) - 1
}

// GoTo sets the visible index.
func (h *History) GoTo(at int) { h.at = at }

// Retarget replaces the current source without moving the index, storing
// the fresh Scan verdicts with it.
func (h *History) Retarget(src string) {
	hash, delim := Scan(src)
	h.past[h.at] = src
	h.hashes[h.at] = hash
	h.delims[h.at] = delim
}

// SourceAt returns the source at at.
func (h *History) SourceAt(at int) string { return h.past[at] }

// HashAt returns the stored hash of the source at at: Back and Forward
// compare entries without hashing either side.
func (h *History) HashAt(at int) uint64 { return h.hashes[at] }

// DelimAt returns the stored delimiter verdict for the source at at, so
// Back and Forward never rescan the target.
func (h *History) DelimAt(at int) bool { return h.delims[at] }

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

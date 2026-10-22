// Package nav holds template history and named HTML routes for package
// page. It does not import page.
package nav

// History owns template history and routes behind narrow methods.
// past holds every loaded source; at is the visible index. routes maps
// a data-action value to HTML for Click to load.
type History struct {
	past   []string
	at     int
	routes map[string]string
}

// New starts history at source.
func New(source string) History {
	return History{past: []string{source}}
}

// Current is the template source at the visible index.
func (h *History) Current() string { return h.past[h.at] }

// Doc is the visible history index.
func (h *History) Doc() int { return h.at }

// Push records html and drops any later entries.
func (h *History) Push(html string) {
	h.past = append(h.past[:h.at+1], html)
	h.at = len(h.past) - 1
}

// GoTo sets the visible index.
func (h *History) GoTo(at int) { h.at = at }

// Retarget replaces the current source without moving the index.
func (h *History) Retarget(src string) { h.past[h.at] = src }

// SourceAt returns the source at at.
func (h *History) SourceAt(at int) string { return h.past[at] }

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

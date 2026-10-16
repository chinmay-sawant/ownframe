package page

// navHistory owns template history and routes behind narrow methods.
// past holds every loaded source; at is the visible index. routes maps
// a data-action value to HTML for Click to load.
type navHistory struct {
	past   []string
	at     int
	routes map[string]string
}

func newNavHistory(source string) navHistory {
	return navHistory{past: []string{source}}
}

func (h *navHistory) current() string { return h.past[h.at] }

func (h *navHistory) doc() int { return h.at }

func (h *navHistory) push(html string) {
	h.past = append(h.past[:h.at+1], html)
	h.at = len(h.past) - 1
}

func (h *navHistory) goTo(at int) { h.at = at }

func (h *navHistory) retarget(src string) { h.past[h.at] = src }

func (h *navHistory) sourceAt(at int) string { return h.past[at] }

func (h *navHistory) canBack() bool { return h.at > 0 }

func (h *navHistory) canForward() bool { return h.at+1 < len(h.past) }

func (h *navHistory) backIndex() int { return h.at - 1 }

func (h *navHistory) forwardIndex() int { return h.at + 1 }

func (h *navHistory) route(action, html string) {
	if h.routes == nil {
		h.routes = map[string]string{}
	}
	h.routes[action] = html
}

func (h *navHistory) lookup(action string) (string, bool) {
	html, ok := h.routes[action]
	return html, ok
}

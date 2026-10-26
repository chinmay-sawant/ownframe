package web

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// maxKeyName bounds the key field before it is lowercased and matched.
const maxKeyName = 32

// maxKeyPrefixes bounds the shift+, ctrl+, and alt+ chain.
const maxKeyPrefixes = 3

// namedKey reports a key name the screen understands. Printable characters
// use /type; these are the special names a browser key event maps to.
func namedKey(name string) bool {
	switch name {
	case "arrowup", "arrowdown", "arrowleft", "arrowright",
		"pageup", "pagedown", "home", "end",
		"enter", "escape", "space", "tab", "backspace", "delete",
		"control", "shift", "meta", "alt":
		return true
	}

	return false
}

// caretKey reports a key that may carry modifier prefixes, the way the
// native window names shift+arrowleft and ctrl+arrowright.
func caretKey(name string) bool {
	switch name {
	case "arrowleft", "arrowright", "home", "end":
		return true
	}

	return false
}

// normalizeKey lowercases and validates one key name. Unknown names and
// names longer than maxKeyName are rejected.
func normalizeKey(raw string) (string, bool) {
	if len(raw) == 0 || len(raw) > maxKeyName {
		return "", false
	}
	key := strings.ToLower(strings.TrimSpace(raw))
	if key == "" {
		return "", false
	}
	if namedKey(key) {
		return key, true
	}
	// A printable key reaches /key only when a modifier is held: the browser
	// sends plain text through /type. The native window sends the base key
	// (for example "s" after Control) and the page tracks the modifier.
	if len(key) == 1 && key[0] > ' ' && key[0] < 127 {
		return key, true
	}

	base := key
	prefixes := 0
	for {
		switch {
		case strings.HasPrefix(base, "shift+"):
			base = base[len("shift+"):]
		case strings.HasPrefix(base, "ctrl+"):
			base = base[len("ctrl+"):]
		case strings.HasPrefix(base, "alt+"):
			base = base[len("alt+"):]
		default:
			if prefixes > 0 && caretKey(base) {
				return key, true
			}

			return "", false
		}
		prefixes++
		if prefixes > maxKeyPrefixes {
			return "", false
		}
	}
}

// key forwards one key press or release. Tab moves focus through
// host.Focuser because the native window consumes it; a screen without a
// Focuser still receives the key.
func (s *server) key(w http.ResponseWriter, r *http.Request) {
	if !requireOrigin(r) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxFormBytes)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	key, ok := normalizeKey(r.PostFormValue("key"))
	if !ok {
		http.Error(w, "bad key", http.StatusBadRequest)
		return
	}
	down, ok := keyDownFlag(r.PostFormValue("down"))
	if !ok {
		http.Error(w, "bad down", http.StatusBadRequest)
		return
	}
	shift := r.PostFormValue("shift") == "1"

	s.mu.Lock()
	defer s.mu.Unlock()

	if key == "tab" {
		if !down {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if focuser, ok := s.app.(host.Focuser); ok {
			if err := focusKey(r.Context(), focuser, shift); err != nil {
				log.Printf("key: %s", sanitizeLine(err.Error()))
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)

			return
		}
	}

	var err error
	if down {
		err = s.kbd().KeyDown(r.Context(), key)
		if err == nil && key == "enter" {
			err = s.kbd().Submit(r.Context())
		}
	} else {
		err = s.kbd().KeyUp(r.Context(), key)
	}
	if err != nil {
		log.Printf("key: %s", sanitizeLine(err.Error()))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// keyDownFlag accepts the 1/0 field the browser sends.
func keyDownFlag(raw string) (bool, bool) {
	switch raw {
	case "1":
		return true, true
	case "0":
		return false, true
	}

	return false, false
}

// focusKey moves focus for Tab or Shift+Tab.
func focusKey(ctx context.Context, focuser host.Focuser, shift bool) error {
	if shift {
		return focuser.FocusPrev(ctx)
	}

	return focuser.FocusNext(ctx)
}

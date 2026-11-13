package web

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

// maxDOMIDLength bounds element ids and action names the browser may name
// in /click and /change. maxChangeValue caps one stored control value at a
// quarter of the POST body cap.
const (
	maxDOMIDLength = 256
	maxChangeValue = maxFormBytes / 4
)

// domSource is a screen that exposes its executed template source and theme
// text. internal/page implements it; the web package only asserts it so the
// ownframe.Page surface does not grow.
type domSource interface {
	ExecutedSource() string
	ThemeSource() string
}

// domOption is one dropdown choice served to the live shell.
type domOption struct {
	Value    string `json:"value"`
	Label    string `json:"label"`
	Selected bool   `json:"selected"`
	Disabled bool   `json:"disabled"`
}

// domSelectSource is a screen that lists live select options. The executed
// source carries only the current label, so dropdowns need this to open.
type domSelectSource interface {
	SelectControls() map[string][]page.SelectOption
}

// formChecker stores a checked state without drawing.
type formChecker interface {
	SetFormChecked(id string, checked bool)
}

// webControls lists live controls for the browser document. A screen that
// does not implement it leaves the control list nil.
type webControls interface {
	WebControls() []page.WebControl
}

// webTextApplier stores one posted text or select value and redraws.
type webTextApplier interface {
	ApplyWebText(ctx context.Context, id, value string) error
}

// webCheckedApplier stores one posted checked state and redraws.
type webCheckedApplier interface {
	ApplyWebChecked(ctx context.Context, id string, checked bool) error
}

// domState is the JSON body of GET /dom. HTML is the browser document for
// the executed source, theme is the raw theme stylesheet, selects carries
// live dropdown options by control id, and fallback tells the shell to show
// the PNG when no source exists yet.
type domState struct {
	Generation uint64                 `json:"generation"`
	Width      int                    `json:"width"`
	Height     int                    `json:"height"`
	HTML       string                 `json:"html"`
	Theme      string                 `json:"theme"`
	Selects    map[string][]domOption `json:"selects,omitempty"`
	Fallback   bool                   `json:"fallback"`
}

// dom answers GET /dom with the live document. The shell renders these tags
// directly and patches only the id-anchored subtrees that change, so a
// redraw updates components instead of the whole page. Generation is the
// page generation without frameGen: retained-op ticks and scroll or hover
// frame bumps change the PNG frame but leave the executed source alone, so
// they must not trigger a /dom refetch.
func (s *server) dom(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.reload(r.Context())
	width, height := s.app.Size()
	state := domState{
		Generation: s.app.Generation(),
		Width:      width,
		Height:     height,
		Fallback:   true,
	}
	if src, ok := s.app.(domSource); ok {
		if html := src.ExecutedSource(); html != "" {
			theme := src.ThemeSource()
			var controls []page.WebControl
			if listed, ok := s.app.(webControls); ok {
				controls = listed.WebControls()
			}
			// An empty boot keeps the shell script out of every patch payload.
			state.HTML = browserDocument(
				html,
				theme,
				controls,
				"",
			)
			state.Theme = theme
			state.Fallback = false
		}
	}
	if !state.Fallback {
		if sel, ok := s.app.(domSelectSource); ok {
			for id, opts := range sel.SelectControls() {
				conv := make([]domOption, 0, len(opts))
				for _, o := range opts {
					conv = append(conv, domOption{
						Value:    o.Value,
						Label:    o.Label,
						Selected: o.Selected,
						Disabled: o.Disabled,
					})
				}
				if state.Selects == nil {
					state.Selects = map[string][]domOption{}
				}
				state.Selects[id] = conv
			}
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if err := json.NewEncoder(w).Encode(state); err != nil {
		log.Printf("write dom: %s", sanitizeLine(err.Error()))
	}
}

// changeWriter stores one control value. Select elements post here; text
// fields stream through /type so BeforeEdit and Change still run per edit.
type changeWriter interface {
	SetFormValue(id, value string)
	SetFormChecked(id string, checked bool)
}

// change stores a select or checkbox value from the live DOM. It focuses the
// control first so the redraw keeps the same focus the desktop window would.
// ApplyWebText and ApplyWebChecked redraw themselves. A screen without them
// stores the value here and redraws so the next /dom poll carries the source.
func (s *server) change(w http.ResponseWriter, r *http.Request) {
	if !requireOrigin(r) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxFormBytes)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	id := r.PostFormValue("id")
	if id == "" || len(id) > maxDOMIDLength || strings.ContainsAny(id, "<>\"'\r\n\x00") {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	_, hasValue := r.PostForm["value"]
	value := r.PostFormValue("value")
	if len(value) > maxChangeValue {
		http.Error(w, "bad value", http.StatusBadRequest)
		return
	}
	_, hasChecked := r.PostForm["checked"]
	checkedRaw := r.PostFormValue("checked")
	checked := checkedRaw == "1" || checkedRaw == "true"
	_, hasCaret := r.PostForm["caret"]
	caretRaw := r.PostFormValue("caret")
	if len(caretRaw) > 10 {
		http.Error(w, "bad caret", http.StatusBadRequest)
		return
	}
	caret := 0
	if hasCaret {
		var err error
		caret, err = strconv.Atoi(caretRaw)
		if err != nil || caret < 0 {
			http.Error(w, "bad caret", http.StatusBadRequest)
			return
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if focuser, ok := s.app.(interface {
		Focus(context.Context, string) error
	}); ok {
		if err := focuser.Focus(r.Context(), id); err != nil {
			http.Error(w, "bad id", http.StatusBadRequest)
			return
		}
	}
	// A caret without a value only aligns the server caret with the browser
	// before the backspaces and typing that follow it, so it does not draw.
	if hasCaret && !hasValue && !hasChecked {
		if placer, ok := s.app.(interface {
			PlaceCaret(string, int)
		}); ok {
			placer.PlaceCaret(id, caret)
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	// An id without a value only moves focus, so a focusin beat does not
	// wipe the field it just entered.
	if !hasValue && !hasChecked {
		if err := s.app.Redraw(r.Context()); err != nil {
			log.Printf("change: %s", sanitizeLine(err.Error()))
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		s.syncScroll()
		s.publishIfChangedLocked()
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if hasChecked {
		if applier, ok := s.app.(webCheckedApplier); ok {
			if err := applier.ApplyWebChecked(r.Context(), id, checked); err != nil {
				log.Printf("change: %s", sanitizeLine(err.Error()))
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			s.syncScroll()
			s.publishIfChangedLocked()
			w.WriteHeader(http.StatusNoContent)
			return
		}
	} else if applier, ok := s.app.(webTextApplier); ok {
		if err := applier.ApplyWebText(r.Context(), id, value); err != nil {
			log.Printf("change: %s", sanitizeLine(err.Error()))
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		s.syncScroll()
		s.publishIfChangedLocked()
		w.WriteHeader(http.StatusNoContent)
		return
	}
	writer, ok := s.app.(changeWriter)
	if !ok {
		http.Error(w, "unsupported", http.StatusNotImplemented)
		return
	}
	if hasChecked {
		if checker, ok := s.app.(formChecker); ok {
			checker.SetFormChecked(id, checked)
		} else {
			writer.SetFormValue(id, value)
		}
	} else {
		writer.SetFormValue(id, value)
	}
	if err := s.app.Redraw(r.Context()); err != nil {
		log.Printf("change: %s", sanitizeLine(err.Error()))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	s.syncScroll()
	s.publishIfChangedLocked()
	w.WriteHeader(http.StatusNoContent)
}

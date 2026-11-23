package web

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"log"
	"net/http"
	"strings"

	"github.com/chinmay-sawant/ownframe/internal/bitmap"
	"github.com/chinmay-sawant/ownframe/internal/host"
)

// reply is a response built while s.mu is held and written after it is
// released, so a slow client never stalls the tick loop or other routes.
type reply struct {
	status int
	ctype  string
	body   []byte
	errMsg string // when set, sent with http.Error instead of body
}

// send writes the reply. Call it with s.mu released.
func (rp reply) send(w http.ResponseWriter, what string) {
	if rp.errMsg != "" {
		http.Error(w, rp.errMsg, rp.status)
		return
	}
	w.Header().Set("Content-Type", rp.ctype)
	w.Header().Set("Cache-Control", "no-store")
	if rp.status != 0 && rp.status != http.StatusOK {
		w.WriteHeader(rp.status)
	}
	if _, err := w.Write(rp.body); err != nil {
		log.Printf("write %s: %s", what, sanitizeLine(err.Error()))
	}
}

// jsonReply encodes v into a buffered reply.
func jsonReply(ctype string, v any, what string) reply {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		log.Printf("encode %s: %s", what, sanitizeLine(err.Error()))
		return reply{status: http.StatusInternalServerError, errMsg: "internal server error"}
	}

	return reply{ctype: ctype, body: buf.Bytes()}
}

func (s *server) page(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	rp := s.pageReplyLocked(r)
	s.mu.Unlock()
	rp.send(w, "page")
}

func (s *server) pageReplyLocked(r *http.Request) reply {
	s.reload(r.Context())
	if src, ok := s.app.(domSource); ok {
		source := src.ExecutedSource()
		if source != "" && r.URL.Query().Get("dom") != "0" {
			doc := browserDocument(
				source,
				src.ThemeSource(),
				s.liveControls(),
				bootScript,
			)
			return reply{ctype: "text/html; charset=utf-8", body: []byte(doc)}
		}
	}
	viewW, viewH := s.app.Size()
	data := shellData{
		Areas:      areas(s.app),
		Reload:     s.watching(),
		Width:      viewW,
		Height:     viewH,
		Generation: s.app.Generation() + s.frameGen,
	}
	var buf bytes.Buffer
	if err := s.shell.Execute(&buf, data); err != nil {
		log.Printf("page: %s", sanitizeLine(err.Error()))
		return reply{status: http.StatusInternalServerError, errMsg: "internal server error"}
	}

	return reply{ctype: "text/html; charset=utf-8", body: buf.Bytes()}
}

func (s *server) liveControls() []host.WebControl {
	src, ok := s.app.(host.WebDOM)
	if !ok {
		return nil
	}
	return s.redactControls(src.WebControls())
}

// redactControls drops password values from controls headed for the browser
// when the server accepts remote clients. It copies before changing anything,
// because the slice belongs to the page.
func (s *server) redactControls(controls []host.WebControl) []host.WebControl {
	if !s.redactPasswords {
		return controls
	}
	out := controls
	for i := range controls {
		if !strings.EqualFold(controls[i].Type, "password") || controls[i].Value == "" {
			continue
		}
		if &out[0] == &controls[0] {
			out = append([]host.WebControl(nil), controls...)
		}
		out[i].Value = ""
	}

	return out
}

func (s *server) frame(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	rp := s.frameReplyLocked(r)
	s.mu.Unlock()
	rp.send(w, "png")
}

func (s *server) frameReplyLocked(r *http.Request) reply {
	s.reload(r.Context())
	var pngBytes []byte
	if s.app.Display() != nil {
		viewW, h := s.app.Size()
		gen := s.app.Generation() + s.frameGen
		if s.viewportPNG != nil && s.viewportGen == gen && s.viewportWidth == viewW && s.viewportHeight == h {
			pngBytes = s.viewportPNG
		} else {
			img, err := bitmap.Viewport(s.app.Display(), image.Rect(0, s.scrollY, viewW, s.scrollY+h), 1)
			if err != nil {
				return reply{status: http.StatusInternalServerError, errMsg: "render failed"}
			}
			var buf bytes.Buffer
			if err = png.Encode(&buf, img); err != nil {
				return reply{status: http.StatusInternalServerError, errMsg: "encode failed"}
			}
			pngBytes = buf.Bytes()
			s.viewportPNG = pngBytes
			s.viewportGen = gen
			s.viewportWidth = viewW
			s.viewportHeight = h
		}
	} else {
		pngBytes = s.app.PNG()
	}
	if len(pngBytes) == 0 {
		return reply{status: http.StatusNotFound, errMsg: "no frame"}
	}

	return reply{ctype: "image/png", body: pngBytes}
}

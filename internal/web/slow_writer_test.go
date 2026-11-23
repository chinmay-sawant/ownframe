package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// blockingWriter stalls its first Write until release is closed, like a
// client that stops reading its socket.
type blockingWriter struct {
	header  http.Header
	entered chan struct{}
	release chan struct{}
}

func (b *blockingWriter) Header() http.Header { return b.header }
func (b *blockingWriter) WriteHeader(int)     {}
func (b *blockingWriter) Write(p []byte) (int, error) {
	select {
	case b.entered <- struct{}{}:
	default:
	}
	<-b.release

	return len(p), nil
}

// TestHandlersReleaseLockBeforeSocketWrite proves the tick loop's mutex is
// free while a slow client holds a response write open.
func TestHandlersReleaseLockBeforeSocketWrite(t *testing.T) {
	handlers := map[string]func(*server) http.HandlerFunc{
		"/state":       func(s *server) http.HandlerFunc { return s.state },
		"/dom":         func(s *server) http.HandlerFunc { return s.dom },
		"/":            func(s *server) http.HandlerFunc { return s.page },
		"/debug/state": func(s *server) http.HandlerFunc { return s.debug },
		"/frame.png":   func(s *server) http.HandlerFunc { return s.frame },
	}
	for path, pick := range handlers {
		t.Run(path, func(t *testing.T) {
			t.Parallel()

			srv := &server{app: &pngScreen{domScreen{width: 100, height: 50, source: "<p>hi</p>"}}}
			w := &blockingWriter{header: http.Header{}, entered: make(chan struct{}, 1), release: make(chan struct{})}
			done := make(chan struct{})
			go func() {
				defer close(done)
				r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "http://example.com"+path, nil)
				pick(srv)(w, r)
			}()

			select {
			case <-w.entered:
			case <-time.After(2 * time.Second):
				t.Fatal("handler never wrote")
			}
			locked := make(chan struct{})
			go func() {
				srv.mu.Lock()
				srv.mu.Unlock() //nolint:staticcheck // reason: only acquisition is under test.
				close(locked)
			}()
			select {
			case <-locked:
			case <-time.After(2 * time.Second):
				close(w.release)
				t.Fatal("server mutex held across the socket write")
			}
			close(w.release)
			<-done
		})
	}
}

// pngScreen is a domScreen that serves a fixed PNG for /frame.png.
type pngScreen struct{ domScreen }

func (*pngScreen) PNG() []byte { return []byte("\x89PNG") }

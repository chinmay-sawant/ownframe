package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/host"
)

// racyBoxScreen mutates its box list in Tick and reads it in Boxes with no
// lock of its own, so -race fails when a handler reads Boxes outside s.mu.
type racyBoxScreen struct {
	host.Screen
	boxes []layout.Box
}

func (f *racyBoxScreen) Boxes() []layout.Box                           { return f.boxes }
func (f *racyBoxScreen) Click(context.Context, float64, float64) error { return nil }
func (f *racyBoxScreen) Generation() uint64                            { return 0 }
func (f *racyBoxScreen) Size() (int, int)                              { return 320, 200 }
func (f *racyBoxScreen) Tick(context.Context) error {
	f.boxes = append([]layout.Box(nil), layout.Box{ID: "a", W: 10, H: 10})

	return nil
}

func TestClickBoxesResolvedUnderLock(t *testing.T) {
	t.Parallel()

	app := &racyBoxScreen{boxes: []layout.Box{{ID: "a", W: 10, H: 10}}}
	srv := &server{app: app}
	ctx := t.Context()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for range 200 {
			_ = srv.tickOnce(ctx, app)
		}
	}()
	go func() {
		defer wg.Done()
		for range 200 {
			r := httptest.NewRequestWithContext(ctx, http.MethodPost, "http://example.com/click?id=a", nil)
			r.Header.Set("Origin", "http://example.com")
			srv.click(httptest.NewRecorder(), r)
		}
	}()
	wg.Wait()
}

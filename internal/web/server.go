// Package web shows a screen's PNG in a browser.
// It is one way to deliver mouse and keyboard events to a host.Screen.
// The native window is internal/window.
package web

import (
	"context"
	"html/template"
	"sync"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

type shellArea struct {
	Coords string
	Href   string
	Alt    string
	X      float64
	Y      float64
}

type shellData struct {
	Width  int
	Height int
	Areas  []shellArea
	Reload bool
}

type server struct {
	mu           sync.Mutex
	app          host.Screen
	shell        *template.Template
	lastNote     string
	perf         bool
	debugEnabled bool
}

// Serve listens on addr and blocks. The page at / shows the latest PNG.
// Developer endpoints stay off; ServeWithOptions with Options.Perf turns
// on /debug/pprof/*.
func Serve(app host.Screen, addr string) error {
	return ServeWithOptions(app, addr, Options{})
}

// ServeWithContext stops the listener when ctx ends.
func ServeWithContext(ctx context.Context, app host.Screen, addr string) error {
	return ServeWithOptionsContext(ctx, app, addr, Options{})
}

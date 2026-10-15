package web

import (
	"fmt"
	"html/template"
	"net"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// Options tunes ServeWithOptions. Perf registers /debug/pprof/* for CPU,
// heap, goroutine, and trace profiles. It is off by default, so a shipped
// server exposes no profiler until a developer opts in.
type Options struct {
	Perf bool
}

// ServeWithOptions listens on addr and blocks. The page at / shows the
// latest PNG. With Options.Perf it also serves the pprof endpoints.
func ServeWithOptions(app host.Screen, addr string, opts Options) error {
	shell, err := template.New("shell").Parse(shellHTML)
	if err != nil {
		return err
	}

	srv := &server{
		app:   app,
		shell: shell,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", srv.page)
	mux.HandleFunc("GET /frame.png", srv.frame)
	mux.HandleFunc("GET /debug/state", srv.debug)
	if opts.Perf {
		mux.HandleFunc("GET /debug/pprof/", pprof.Index)
		mux.HandleFunc("GET /debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("GET /debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("GET /debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("GET /debug/pprof/trace", pprof.Trace)
	}
	if _, ok := app.(pdfScreen); ok {
		mux.HandleFunc("GET /pdf", srv.pdf)
	}
	mux.HandleFunc("GET /click", srv.click)
	mux.HandleFunc("POST /type", srv.typeText)
	mux.HandleFunc("POST /backspace", srv.backspace)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	fmt.Println("http://" + addr + "/")

	httpSrv := &http.Server{
		Handler:           securityHeaders(mux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return httpSrv.Serve(ln)
}

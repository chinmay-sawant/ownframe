package web

import (
	"context"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// ServeWithOptions listens on addr and blocks.
func ServeWithOptions(app host.Screen, addr string, opts Options) error {
	return ServeWithOptionsContext(context.Background(), app, addr, opts)
}

// ServeWithOptionsContext stops on ctx end.
func ServeWithOptionsContext(ctx context.Context, app host.Screen, addr string, opts Options) error {
	if ctx == nil {
		return ErrNilContext
	}
	if app == nil {
		return ErrNilApp
	}
	if err := validateBindAddr(addr); err != nil {
		return err
	}
	t, err := template.New("shell").Parse(shellHTML)
	if err != nil {
		return err
	}
	srv := &server{app: app, shell: t, perf: opts.Perf, debugEnabled: opts.Perf || opts.Debug}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", srv.page)
	mux.HandleFunc("GET /frame.png", srv.frame)
	if srv.debugEnabled {
		mux.HandleFunc("GET /debug/state", srv.debug)
	}
	if opts.Perf {
		mux.HandleFunc("GET /debug/pprof/", pprof.Index)
		mux.HandleFunc("GET /debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("GET /debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("GET /debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("GET /debug/pprof/trace", pprof.Trace)
	}
	mux.HandleFunc("POST /click", srv.click)
	mux.HandleFunc("POST /type", srv.typeText)
	mux.HandleFunc("POST /backspace", srv.backspace)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	fmt.Println("http://" + addr + "/")
	hs := &http.Server{Handler: panicRecovery(securityHeaders(mux)), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	stop := context.AfterFunc(ctx, func() { hs.Shutdown(context.Background()) })
	defer stop()
	err = hs.Serve(ln)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return err
}

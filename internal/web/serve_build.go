package web

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"

	"github.com/chinmay-sawant/ownframe/internal/debugtag"
	"github.com/chinmay-sawant/ownframe/internal/host"
)

// ServeWithOptions listens on addr.
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
	if err := checkBindAddr(addr, opts); err != nil {
		return err
	}
	t, err := template.New("shell").Parse(shellHTML)
	if err != nil {
		return err
	}
	srv := &server{app: app, shell: t, perf: effectivePprof(opts), debugEnabled: effectiveDebug(opts)}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", srv.page)
	mux.HandleFunc("GET /frame.png", srv.frame)
	mux.HandleFunc("GET /state", srv.state)
	if srv.debugEnabled {
		if debugtag.Enabled {
			mux.HandleFunc("GET /debug/state", srv.debug)
		} else {
			log.Print("web: debug endpoints need -tags ownframe_debug")
		}
	}
	if effectivePprof(opts) {
		registerPprof(mux)
	}
	mux.Handle("POST /click", srv.limit(http.HandlerFunc(srv.click)))
	mux.Handle("POST /type", srv.limit(http.HandlerFunc(srv.typeText)))
	mux.Handle("POST /backspace", srv.limit(http.HandlerFunc(srv.backspace)))
	mux.Handle("POST /key", srv.limit(http.HandlerFunc(srv.key)))
	mux.Handle("POST /wheel", srv.limit(http.HandlerFunc(srv.wheel)))
	ln, err := (&net.ListenConfig{}).Listen(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	stopTick := srv.startTick(ctx)
	defer stopTick()
	fmt.Println("http://" + addr + "/")
	hs := &http.Server{Handler: panicRecovery(securityHeaders(requireLocalHost(bindHostName(addr), mux))), ReadHeaderTimeout: readHeaderTimeout, ReadTimeout: readTimeout, WriteTimeout: writeTimeout, IdleTimeout: idleTimeout, MaxHeaderBytes: maxHeaderBytes}
	stop := context.AfterFunc(ctx, func() { _ = hs.Shutdown(context.Background()) })
	defer stop()
	err = hs.Serve(ln)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return err
}

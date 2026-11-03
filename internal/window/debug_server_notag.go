//go:build !ownframe_debug

package window

import (
	"context"
	"log"
)

// startDebugServer is a no-op without the ownframe_debug build tag. The
// debug listener stays out of release binaries: no net/http server links,
// no pprof, no snapshot publish. Run continues with the window only.
func startDebugServer(_ context.Context, _ *shell, addr string) (string, func(), error) {
	if addr != "" {
		log.Print("window: debug endpoints need -tags ownframe_debug")
	}
	return "", func() {}, nil
}

package web

import (
	"errors"
	"net"
)

// errRemoteBind means a non-loopback bind lacks its explicit opt-in.
var errRemoteBind = errors.New("ownframe: non-loopback bind needs Options.InsecureAllowRemoteBind")

// compatPerf reads the deprecated Perf alias in one place.
func compatPerf(o Options) bool {
	return o.Perf
}

// compatDebug reads the deprecated Debug alias in one place.
func compatDebug(o Options) bool {
	return o.Debug
}

// effectivePprof reports whether pprof endpoints are opted in.
func effectivePprof(o Options) bool {
	return o.InsecureAllowPprof || compatPerf(o)
}

// effectiveDebug reports whether debug state is opted in.
func effectiveDebug(o Options) bool {
	return o.InsecureAllowDebugState || compatDebug(o)
}

// checkBindAddr keeps the wildcard rejection and adds the GEN-22 rule:
// a non-loopback host needs InsecureAllowRemoteBind. Unresolvable names
// fail safe the same way.
func checkBindAddr(addr string, o Options) error {
	if o.Listener != nil {
		return nil
	}
	if err := validateBindAddr(addr); err != nil {
		return err
	}
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}
	if host == "localhost" {
		return nil
	}
	if ip := net.ParseIP(host); ip == nil || !ip.IsLoopback() {
		if !o.InsecureAllowRemoteBind {
			return errRemoteBind
		}
	}

	return nil
}

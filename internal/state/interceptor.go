package state

// RouteInterceptor lets a handler intercept navigation without triggering a
// full document load. Returning true stops the default load.
type RouteInterceptor func(path string) bool

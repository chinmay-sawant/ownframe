//go:build js || android || ios

package reload

// hosts is false on wasm and mobile: a file is read once at New.
const hosts = false

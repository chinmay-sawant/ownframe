//go:build !js && !android && !ios

package reload

// hosts is true on the desktop and server hosts, which reread a file
// while the process runs.
const hosts = true

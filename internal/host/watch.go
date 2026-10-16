package host

// Watcher is a watched screen: it polls for file changes and reports
// whether a poll can reread a source file.
type Watcher interface {
	Reloader
	Watching() bool
}

package web

import (
	"context"
	"fmt"
	"os"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// reload rereads the screen's files before a request is answered.
func (s *server) reload(ctx context.Context) {
	reloader, ok := s.app.(host.Reloader)
	if !ok {
		return
	}

	if _, err := reloader.PollReload(ctx); err != nil {
		if line, show := s.reloadNote(err); show {
			fmt.Fprintln(os.Stderr, line)
		}

		return
	}

	s.lastNote = ""
}

// watching reports whether the shell page should refresh its frame.
func (s *server) watching() bool {
	w, ok := s.app.(host.Watcher)

	return ok && w.Watching()
}

// reloadNote returns the stderr line for err, once per distinct message.
func (s *server) reloadNote(err error) (string, bool) {
	line := "hot reload: " + err.Error()
	if line == s.lastNote {
		return "", false
	}

	s.lastNote = line

	return line, true
}

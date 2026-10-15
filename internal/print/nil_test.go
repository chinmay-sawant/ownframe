package print

import (
	"errors"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/fetch"
)

func TestRunRejectsNilContext(t *testing.T) {
	t.Parallel()
	if err := run(nil, "/tmp/r.pdf", linuxCommands); !errors.Is(err, fetch.ErrNilContext) {
		t.Fatalf("err = %v", err)
	}
}

//go:build (!linux || ios || js) && !android && !(windows && (amd64 || arm64))

package window

import (
	"context"
	"testing"
)

func TestA11yStubCompilesWithoutSuccess(t *testing.T) {
	s := newShell(context.Background(), &fakeScreen{width: 10, height: 10})
	s.initA11y()
	s.syncA11y()
}

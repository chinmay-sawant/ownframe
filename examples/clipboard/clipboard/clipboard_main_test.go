package clipboard_test

import (
	"os"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// TestMain keeps every test off the desktop clipboard.
func TestMain(m *testing.M) {
	ownframe.UseMemory(true)
	os.Exit(m.Run())
}

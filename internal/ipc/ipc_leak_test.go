package ipc_test

import (
	"os"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/leakcheck"
)

// TestMain fails the binary when ipc tests leak goroutines (GEN-26).
func TestMain(m *testing.M) {
	os.Exit(leakcheck.Verify(m))
}

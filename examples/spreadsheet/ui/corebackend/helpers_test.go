package corebackend

import (
	"context"
	"testing"
)

// newBackend opens a store with only the demo workbook. The shipped full
// fixture also carries a 100,000-row stress workbook, which no UI test reads.
func newBackend(t *testing.T, dir string) *Backend {
	t.Helper()

	b, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	if err := b.SeedDemo(context.Background()); err != nil {
		_ = b.Close()
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = b.Close() })

	return b
}

package storage

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestLockedDatabaseSurfacesError(t *testing.T) {
	st, dir := openDir(t)

	other := raw(t, dir)

	if _, err := other.Exec(`BEGIN IMMEDIATE`); err != nil {
		t.Fatal(err)
	}

	defer func() { _, _ = other.Exec(`ROLLBACK`) }()

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := st.SetPref(ctx, "k", "v")
	if err == nil {
		t.Fatal("write on a locked database succeeded")
	}

	if !errors.Is(err, context.DeadlineExceeded) && !strings.Contains(err.Error(), "locked") {
		t.Fatalf("lock error = %v", err)
	}

	if _, err := other.Exec(`ROLLBACK`); err != nil {
		t.Fatal(err)
	}

	if err := st.SetPref(context.Background(), "k", "v"); err != nil {
		t.Fatalf("write after unlock: %v", err)
	}
}

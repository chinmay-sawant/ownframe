package storage

import (
	"sync"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/system-monitor/domain"
)

// TestConcurrentReadWrite checks that one connection serializes a writer and
// a reader without a deadlock.
func TestConcurrentReadWrite(t *testing.T) {
	st := openTest(t)
	ctx := t.Context()
	id := newTestSession(t, st)

	errs := make(chan error, 2)

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()

		for i := range 50 {
			row := rec(domain.MetricCPU, "", time.Now().Add(time.Duration(i)), 1)
			if _, err := st.Append(ctx, id, []Row{row}); err != nil {
				errs <- err

				return
			}
		}
	}()

	go func() {
		defer wg.Done()

		for range 50 {
			if _, err := st.History(ctx, Query{SessionID: id, Limit: 10}); err != nil {
				errs <- err

				return
			}
		}
	}()

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Fatal(err)
	}
}

// TestTempStoresAreSeparate checks that each in-memory store has its own
// database.
func TestTempStoresAreSeparate(t *testing.T) {
	ctx := t.Context()

	a, err := OpenWithOptions(Options{Temp: true})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = a.Close() }()

	b, err := OpenWithOptions(Options{Temp: true})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = b.Close() }()

	if err := a.SetSetting(ctx, "k", "a"); err != nil {
		t.Fatal(err)
	}
	if err := b.SetSetting(ctx, "k", "b"); err != nil {
		t.Fatal(err)
	}

	if got, _, err := a.Setting(ctx, "k"); err != nil || got != "a" {
		t.Fatalf("a = %q err = %v", got, err)
	}
	if got, _, err := b.Setting(ctx, "k"); err != nil || got != "b" {
		t.Fatalf("b = %q err = %v", got, err)
	}
}

// TestCanceledAndClosed is in store_closed_test.go.

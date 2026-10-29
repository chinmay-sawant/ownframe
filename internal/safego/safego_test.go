package safego

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/crash"
)

const (
	reportTimeout = 5 * time.Second
	reportPoll    = 10 * time.Millisecond
)

func TestSurvivesPanic(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "reports")
	if err := crash.SetDir(dir); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = crash.SetDir(t.TempDir()) })

	done := make(chan struct{})
	Go("test", func() {
		defer close(done)
		panic("boom")
	})

	<-done

	// Go recovers the panic and writes the crash report on its own
	// goroutine, so done closes before the write lands. Poll for the
	// report file: returning early lets TempDir cleanup race the write
	// (directory not empty or ENOENT).
	deadline := time.Now().Add(reportTimeout)
	for {
		entries, readErr := os.ReadDir(dir)
		if readErr == nil && len(entries) > 0 {
			break
		}
		if time.Now().After(deadline) {
			if readErr != nil {
				t.Fatalf("crash report not written in %s: %v", dir, readErr)
			}
			t.Fatalf("crash report not written in %s", dir)
		}
		time.Sleep(reportPoll)
	}
}

func TestRunsClean(t *testing.T) {
	t.Parallel()

	done := make(chan string, 1)
	Go("test", func() { done <- "ok" })

	if got := <-done; got != "ok" {
		t.Fatalf("got %q", got)
	}
}

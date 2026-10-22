package safego

import (
	"path/filepath"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/crash"
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
}

func TestRunsClean(t *testing.T) {
	t.Parallel()

	done := make(chan string, 1)
	Go("test", func() { done <- "ok" })

	if got := <-done; got != "ok" {
		t.Fatalf("got %q", got)
	}
}

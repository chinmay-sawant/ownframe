package print

import (
	"context"
	"errors"
	"testing"
)

func TestRunPicksFirstInstalled(t *testing.T) {
	calls := fakeRunner(t, map[string]bool{"lp": true, "xdg-open": true})

	if err := run(context.Background(), "/tmp/r.pdf", linuxCommands); err != nil {
		t.Fatal(err)
	}

	got := *calls
	if len(got) != 1 || got[0].name != "/bin/lp" {
		t.Fatalf("calls = %+v", got)
	}

	if len(got[0].args) != 2 || got[0].args[0] != "--" || got[0].args[1] != "/tmp/r.pdf" {
		t.Fatalf("args = %v", got[0].args)
	}
}

func TestRunSkipsMissing(t *testing.T) {
	calls := fakeRunner(t, map[string]bool{"xdg-open": true})

	if err := run(context.Background(), "/tmp/r.pdf", linuxCommands); err != nil {
		t.Fatal(err)
	}

	if got := *calls; len(got) != 1 || got[0].name != "/bin/xdg-open" {
		t.Fatalf("calls = %+v", got)
	}
}

func TestRunNoTool(t *testing.T) {
	fakeRunner(t, nil)

	if err := run(context.Background(), "/tmp/r.pdf", linuxCommands); !errors.Is(err, ErrNoPrinter) {
		t.Fatalf("err = %v", err)
	}
}

func TestRunContinuesAfterFailure(t *testing.T) {
	calls := fakeRunner(t, map[string]bool{"lp": true, "xdg-open": true})

	runCommand = func(_ context.Context, path string, args ...string) error {
		if path == "/bin/lp" {
			return errors.New("lp failed")
		}

		*calls = append(*calls, call{name: path, args: args})

		return nil
	}

	if err := run(context.Background(), "/tmp/r.pdf", linuxCommands); err != nil {
		t.Fatal(err)
	}

	if got := *calls; len(got) != 1 || got[0].name != "/bin/xdg-open" {
		t.Fatalf("calls = %+v", got)
	}
}

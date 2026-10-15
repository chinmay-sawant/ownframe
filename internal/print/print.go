// Package print hands a PDF file to the operating system print path.
// Linux runs lp or xdg-open, macOS runs osascript or open, and Windows runs
// PowerShell Start-Process -Verb Print. A system with none of those returns
// ErrNoPrinter. OWNFRAME_PRINT_DEBUG=1 logs the command and the fallbacks.
package print

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/chinmay-sawant/ownframe/internal/fetch"
)

// ErrNoPrinter means the system has no print helper to run.
var ErrNoPrinter = errors.New("ownframe: no printer")

// Available reports whether this system has a print helper to try. It is
// false on wasm, Android, and iOS.
func Available() bool {
	switch runtime.GOOS {
	case "linux", "windows", "darwin":
		return true
	}

	return false
}

// command is one print program and how it turns a file path into arguments.
type command struct {
	name string
	args func(path string) []string
}

var (
	lookPath = exec.LookPath
	// runCommand runs one command. Tests replace it with a fake runner.
	runCommand = func(ctx context.Context, path string, args ...string) error {
		return exec.CommandContext(ctx, path, args...).Run()
	}
)

// run starts the first command on PATH and joins helper failures
// with ErrNoPrinter so callers keep the cause.
func run(ctx context.Context, path string, commands []command) error {
	if ctx == nil {
		return fetch.ErrNilContext
	}
	if err := validPath(path); err != nil {
		return err
	}

	var fails []error

	for _, cmd := range commands {
		exe, err := lookPath(cmd.name)
		if err != nil {
			debugf("%s not found", cmd.name)

			continue
		}

		args := cmd.args(path)
		debugf("running %s %v", exe, args)

		if err := runCommand(ctx, exe, args...); err != nil {
			debugf("%s failed: %v", cmd.name, err)
			fails = append(fails, fmt.Errorf("%s: %w", cmd.name, err))

			continue
		}

		return nil
	}

	if len(fails) > 0 {
		return errors.Join(append([]error{ErrNoPrinter}, fails...)...)
	}

	return ErrNoPrinter
}

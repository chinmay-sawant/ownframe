package print

import (
	"context"
	"errors"
	"fmt"

	"github.com/chinmay-sawant/ownframe/internal/fetch"
)

// run starts the first command on PATH and joins failures
// with ErrNoPrinter. Path is cleaned and absolutized first.
func run(ctx context.Context, path string, commands []command) error {
	if ctx == nil {
		return fetch.ErrNilContext
	}
	clean, err := canonicalPath(path)
	if err != nil {
		return err
	}
	var fails []error
	for _, cmd := range commands {
		exe, err := lookPath(cmd.name)
		if err != nil {
			debugf("%s not found", cmd.name)
			continue
		}
		args := cmd.args(clean)
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

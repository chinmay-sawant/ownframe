//go:build linux && !android

package clipboard

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os/exec"
	"strings"
	"time"
)

const wlWait = time.Second

// maxWLOut bounds what wl-paste may hand back, so a hostile or runaway
// clipboard owner cannot grow process memory without limit.
const maxWLOut = 16 << 20

var errWLTooLarge = errors.New("clipboard: wl-paste output too large")

// capBuffer collects output and fails the write once max is exceeded, which
// makes exec stop copying and Wait report the error. It deliberately does
// not embed bytes.Buffer: the promoted ReadFrom would let io.Copy bypass Write.
type capBuffer struct {
	buf bytes.Buffer
	max int
}

func (b *capBuffer) Write(p []byte) (int, error) {
	if b.buf.Len()+len(p) > b.max {
		return 0, errWLTooLarge
	}

	return b.buf.Write(p)
}

func runWL(ctx context.Context, name string, args []string, stdin string) bool {
	path, ok := lookWL(name)
	if !ok {
		return false
	}

	ctx, cancel := context.WithTimeout(ctx, wlWait)
	defer cancel()

	cmd := exec.CommandContext(ctx, path, args...) //nolint:gosec // reason: LookPath of a fixed tool name, constant flags, clipboard text only on stdin, no shell.
	cmd.Stdin = strings.NewReader(stdin)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run() == nil
}

func outWL(ctx context.Context, name string, args []string) (string, bool) {
	return outWLMax(ctx, name, args, maxWLOut)
}

func outWLMax(ctx context.Context, name string, args []string, limit int) (string, bool) {
	path, ok := lookWL(name)
	if !ok {
		return "", false
	}

	ctx, cancel := context.WithTimeout(ctx, wlWait)
	defer cancel()

	cmd := exec.CommandContext(ctx, path, args...) //nolint:gosec // reason: LookPath of a fixed tool name, constant flags, no shell.
	cmd.Stderr = io.Discard
	out := &capBuffer{max: limit}
	cmd.Stdout = out
	if err := cmd.Run(); err != nil {
		return "", false
	}

	return out.buf.String(), true
}

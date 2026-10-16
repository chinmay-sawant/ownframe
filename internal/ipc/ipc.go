// Package ipc passes strings between callers in this process.
// There is no socket, no second process, and no JavaScript.
// Registry is process-global by channel: two pages share one
// namespace, so each channel has one owner that prefixes the
// name and calls cancel. Listen caps per channel.
package ipc

import (
	"context"
	"errors"
	"sync"

	"github.com/chinmay-sawant/ownframe/internal/fetch"
)

// ErrNoHandler means Request found no Handle for that channel.
var ErrNoHandler = errors.New("ownframe: no ipc handler")

var errNilContext = fetch.ErrNilContext

// maxListeners caps Listen per channel.
const maxListeners = 256

type ear struct {
	id int64
	fn func(string)
}

type ask struct {
	id int64
	fn func(context.Context, string) (string, error)
}

var (
	mu   sync.Mutex
	seq  int64
	ears = map[string][]ear{}
	asks = map[string]ask{}
)

func noop() {}

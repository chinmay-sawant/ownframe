//go:build linux && !android

package clipboard

import (
	"errors"
	"net"
	"sync"
	"time"
)

const xWait = time.Second

var ErrX11 = errors.New("ownframe: clipboard x11")

type xReply struct {
	buf []byte
	err error
}

type xClient struct {
	c       net.Conn
	mu      sync.Mutex
	once    sync.Once
	dead    chan struct{}
	wait    map[uint16]chan xReply
	note    chan uint32
	seq     uint16
	base    uint32
	mask    uint32
	xid     uint32
	root    uint32
	win     uint32
	max     int
	clip    uint32
	utf8    uint32
	targets uint32
	incr    uint32
	prop    uint32
	text    string
	owned   bool
}

var (
	xGate sync.Mutex
	live  *xClient
)

func xOpen() (*xClient, error) {
	xGate.Lock()
	defer xGate.Unlock()

	if live != nil {
		select {
		case <-live.dead:
			live = nil
		default:
			return live, nil
		}
	}

	xc, err := dialX()
	if err != nil {
		return nil, err
	}

	live = xc

	return xc, nil
}

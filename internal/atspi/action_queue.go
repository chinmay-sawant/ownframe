package atspi

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

const (
	actionCapacity = 16
	actionBatch    = 4
	actionWaiting  = uint32(0)
	actionRunning  = uint32(1)
	actionCanceled = uint32(2)
)

var (
	ErrActionQueueClosed = errors.New("ownframe: accessibility action queue closed")
	ErrActionQueueFull   = errors.New("ownframe: accessibility action queue full")
)

type ActionHandler func(context.Context, host.AccessibilityRequest) error

// Actions admits requests from bus workers and executes them on the page loop.
// A canceled request cannot win the execution handoff. Cancellation after that
// handoff is cooperative through the action context and does not undo an edit.
type Actions struct {
	mu      sync.Mutex
	ctx     context.Context
	stop    func() bool
	closed  bool
	pending chan *pendingAction
	calls   map[*pendingAction]struct{}
	wake    func()
}

type pendingAction struct {
	ctx     context.Context
	cancel  context.CancelFunc
	request host.AccessibilityRequest
	state   atomic.Uint32
	result  chan error
}

func NewActions(ctx context.Context, wake func()) *Actions {
	a := &Actions{ctx: ctx, pending: make(chan *pendingAction, actionCapacity),
		calls: make(map[*pendingAction]struct{}), wake: wake}
	a.mu.Lock()
	a.stop = context.AfterFunc(ctx, a.Close)
	a.mu.Unlock()
	return a
}

// Request waits for execution, cancellation, or shutdown. Admission is bounded
// and nonblocking. Success means the page action completed without an error.
func (a *Actions) Request(ctx context.Context, request host.AccessibilityRequest) error {
	ctx, cancel := context.WithCancel(ctx)
	call := &pendingAction{ctx: ctx, cancel: cancel, request: request, result: make(chan error, 1)}
	defer func() {
		cancel()
		a.mu.Lock()
		delete(a.calls, call)
		a.mu.Unlock()
	}()
	a.mu.Lock()
	if a.closed || a.ctx.Err() != nil {
		a.mu.Unlock()
		return ErrActionQueueClosed
	}
	if err := ctx.Err(); err != nil {
		a.mu.Unlock()
		return err
	}
	select {
	case a.pending <- call:
		a.calls[call] = struct{}{}
	default:
		a.mu.Unlock()
		return ErrActionQueueFull
	}
	a.mu.Unlock()
	a.wake()
	select {
	case err := <-call.result:
		return err
	case <-ctx.Done():
		call.state.CompareAndSwap(actionWaiting, actionCanceled)
		return ctx.Err()
	}
}

// Drain runs a bounded batch on the caller's page loop. It never waits for bus
// traffic, and it schedules another frame when admitted requests remain.
func (a *Actions) Drain(perform ActionHandler) {
	for range actionBatch {
		select {
		case call := <-a.pending:
			a.mu.Lock()
			claim := !a.closed && a.ctx.Err() == nil && call.ctx.Err() == nil &&
				call.state.CompareAndSwap(actionWaiting, actionRunning)
			a.mu.Unlock()
			if claim {
				call.result <- perform(call.ctx, call.request)
			}
		default:
			return
		}
	}
	if len(a.pending) > 0 {
		a.wake()
	}
}

func (a *Actions) Close() {
	a.mu.Lock()
	a.closed = true
	for call := range a.calls {
		call.cancel()
	}
	a.stop()
	a.mu.Unlock()
}

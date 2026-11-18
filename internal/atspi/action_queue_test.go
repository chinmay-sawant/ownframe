package atspi

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

func actionQueueForTest(t *testing.T) (*Actions, <-chan struct{}) {
	t.Helper()
	wake := make(chan struct{}, actionCapacity)
	queue := NewActions(t.Context(), func() { wake <- struct{}{} })
	t.Cleanup(queue.Close)
	return queue, wake
}

func TestActionQueueWaitsForActualResult(t *testing.T) {
	queue, wake := actionQueueForTest(t)
	request := host.AccessibilityRequest{Document: 1, Node: "id:go", Kind: host.AccessibilityActivate}
	result := make(chan error, 1)
	go func() { result <- queue.Request(t.Context(), request) }()
	<-wake
	select {
	case err := <-result:
		_ = err
		t.Fatal("admission returned success before execution")
	default:
	}
	queue.Drain(func(_ context.Context, got host.AccessibilityRequest) error {
		if got != request {
			t.Fatal("action target changed during delivery")
		}
		return host.ErrAccessibilityNode
	})
	if err := <-result; !errors.Is(err, host.ErrAccessibilityNode) {
		t.Fatalf("execution result = %v", err)
	}
}

func TestActionQueueCancellationBeforeExecution(t *testing.T) {
	queue, wake := actionQueueForTest(t)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	result := make(chan error, 1)
	go func() { result <- queue.Request(ctx, host.AccessibilityRequest{}) }()
	<-wake
	cancel()
	if err := <-result; !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled request = %v", err)
	}
	queue.Drain(func(context.Context, host.AccessibilityRequest) error {
		t.Fatal("canceled request mutated the page")
		return nil
	})
}

func TestActionQueueShutdownCancelsExecutingAndQueuedRequests(t *testing.T) {
	queue, wake := actionQueueForTest(t)
	result := make(chan error, 1)
	go func() { result <- queue.Request(t.Context(), host.AccessibilityRequest{}) }()
	<-wake
	queue.Drain(func(ctx context.Context, _ host.AccessibilityRequest) error {
		queue.Close()
		<-ctx.Done()
		return ctx.Err()
	})
	if err := <-result; !errors.Is(err, context.Canceled) {
		t.Fatalf("executing request survived shutdown: %v", err)
	}
	queue.Close()
	if err := queue.Request(t.Context(), host.AccessibilityRequest{}); !errors.Is(err, ErrActionQueueClosed) {
		t.Fatalf("closed queue accepted action: %v", err)
	}
}

func TestActionQueueShutdownBeforeExecution(t *testing.T) {
	queue, wake := actionQueueForTest(t)
	result := make(chan error, 1)
	go func() { result <- queue.Request(t.Context(), host.AccessibilityRequest{}) }()
	<-wake
	queue.Close()
	if err := <-result; !errors.Is(err, context.Canceled) {
		t.Fatalf("queued request survived shutdown: %v", err)
	}
	queue.Drain(func(context.Context, host.AccessibilityRequest) error {
		t.Fatal("queued request ran after shutdown")
		return nil
	})
}

func TestActionQueueSaturationAndBatchLimit(t *testing.T) {
	queue, wake := actionQueueForTest(t)
	for range actionCapacity {
		queue.pending <- &pendingAction{ctx: t.Context(), result: make(chan error, 1)}
	}
	if err := queue.Request(t.Context(), host.AccessibilityRequest{}); !errors.Is(err, ErrActionQueueFull) {
		t.Fatalf("full queue = %v", err)
	}
	executed := 0
	queue.Drain(func(context.Context, host.AccessibilityRequest) error { executed++; return nil })
	if executed != actionBatch || len(queue.pending) != actionCapacity-actionBatch {
		t.Fatal("page loop did not respect the action batch limit")
	}
	select {
	case <-wake:
	default:
		t.Fatal("pending actions did not schedule another frame")
	}
}

func TestActionQueueExpiredContextDoesNotAdmit(t *testing.T) {
	queue, _ := actionQueueForTest(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := queue.Request(ctx, host.AccessibilityRequest{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("expired context = %v", err)
	}
	if len(queue.pending) != 0 {
		t.Fatal("expired request occupied the queue")
	}
}

func TestActionQueueDeadlineBeforeExecution(t *testing.T) {
	queue, wake := actionQueueForTest(t)
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Millisecond)
	defer cancel()
	result := make(chan error, 1)
	go func() { result <- queue.Request(ctx, host.AccessibilityRequest{}) }()
	<-wake
	if err := <-result; !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("deadline request = %v", err)
	}
	queue.Drain(func(context.Context, host.AccessibilityRequest) error {
		t.Fatal("expired action reached the page")
		return nil
	})
}

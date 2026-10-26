package main

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

const patchWait = 5 * time.Second

func patchHash(seed byte) string {
	return strings.Repeat(string(seed), commitHashLen)
}

func waitPatchResult(t *testing.T, loader *commitPatchLoader) commitPatchResult {
	t.Helper()

	deadline := time.Now().Add(patchWait)

	for {
		if result, ok := loader.Poll(); ok {
			return result
		}

		if time.Now().After(deadline) {
			t.Fatal("timed out waiting for a patch result")
		}

		time.Sleep(time.Millisecond)
	}
}

func TestCommitPatchLoaderCacheHitAndOrder(t *testing.T) {
	loader := NewCommitPatchLoader(t.TempDir(), 4)
	t.Cleanup(loader.Close)

	hash := patchHash('a')
	other := patchHash('b')
	calls := 0

	loader.load = func(_ context.Context, _, path string) (*changedFile, error) {
		calls++

		return &changedFile{Path: path, Add: 1}, nil
	}

	ctx := context.Background()
	loader.Request(ctx, hash, "a.go")
	loader.Request(ctx, hash, "b.go")
	loader.Request(ctx, hash, "a.go")
	loader.Request(ctx, other, "a.go")

	wantKeys := []string{
		commitDetailKey(hash, "a.go"),
		commitDetailKey(hash, "b.go"),
		commitDetailKey(hash, "a.go"),
		commitDetailKey(other, "a.go"),
	}

	results := make([]commitPatchResult, 0, len(wantKeys))

	for range wantKeys {
		results = append(results, waitPatchResult(t, loader))
	}

	for i, result := range results {
		if result.Key != wantKeys[i] {
			t.Fatalf("result %d key = %q, want %q", i, result.Key, wantKeys[i])
		}

		if result.Err != nil || result.File == nil {
			t.Fatalf("result %d = %+v", i, result)
		}
	}

	if calls != 3 {
		t.Fatalf("load calls = %d, want 3", calls)
	}

	if results[2].File != results[0].File {
		t.Fatal("cache hit returned a fresh file instead of the cached one")
	}

	if results[2].File.Path != "a.go" || results[1].File.Path != "b.go" {
		t.Fatalf("files = %q %q", results[2].File.Path, results[1].File.Path)
	}
}

func TestCommitPatchLoaderCapacity(t *testing.T) {
	loader := NewCommitPatchLoader(t.TempDir(), 1)
	t.Cleanup(loader.Close)

	calls := 0

	loader.load = func(_ context.Context, _, path string) (*changedFile, error) {
		calls++

		return &changedFile{Path: path}, nil
	}

	ctx := context.Background()
	first := patchHash('a')
	second := patchHash('b')

	loader.Request(ctx, first, "a.go")
	if result := waitPatchResult(t, loader); result.File == nil {
		t.Fatalf("first load = %+v", result)
	}

	loader.Request(ctx, second, "b.go")
	if result := waitPatchResult(t, loader); result.File == nil {
		t.Fatalf("second load = %+v", result)
	}

	loader.Request(ctx, first, "a.go")
	if result := waitPatchResult(t, loader); result.File == nil {
		t.Fatalf("evicted load = %+v", result)
	}

	if calls != 3 {
		t.Fatalf("load calls = %d, want 3 after eviction", calls)
	}
}

func TestCommitPatchLoaderCanceledRequest(t *testing.T) {
	loader := NewCommitPatchLoader(t.TempDir(), 2)
	t.Cleanup(loader.Close)

	loader.load = func(ctx context.Context, _, path string) (*changedFile, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(patchWait):
			return &changedFile{Path: path}, nil
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	loader.Request(ctx, patchHash('a'), "a.go")

	result := waitPatchResult(t, loader)
	if !errors.Is(result.Err, context.Canceled) {
		t.Fatalf("canceled request error = %v", result.Err)
	}

	if result.File != nil {
		t.Fatalf("canceled request returned a file: %+v", result.File)
	}
}

func TestCommitPatchLoaderClose(t *testing.T) {
	loader := NewCommitPatchLoader(t.TempDir(), 2)
	t.Cleanup(loader.Close)

	started := make(chan struct{})

	loader.load = func(ctx context.Context, _, _ string) (*changedFile, error) {
		close(started)
		<-ctx.Done()

		return nil, ctx.Err()
	}

	loader.Request(context.Background(), patchHash('a'), "a.go")

	select {
	case <-started:
	case <-time.After(patchWait):
		t.Fatal("worker never started the patch load")
	}

	closed := make(chan struct{})

	go func() {
		loader.Close()
		close(closed)
	}()

	select {
	case <-closed:
	case <-time.After(patchWait):
		t.Fatal("Close waited on the in-flight load")
	}

	if result, ok := loader.Poll(); ok {
		t.Fatalf("poll after Close = %+v", result)
	}

	loader.Close()

	loader.Request(context.Background(), patchHash('b'), "b.go")

	if result, ok := loader.Poll(); ok {
		t.Fatalf("request after Close produced %+v", result)
	}
}

func TestCommitPatchLoaderDefaultLoad(t *testing.T) {
	dir, hashes := historyRepo(t)

	loader := NewCommitPatchLoader(dir, 2)
	t.Cleanup(loader.Close)

	loader.Request(context.Background(), hashes[0], "app.go")

	result := waitPatchResult(t, loader)
	if result.Err != nil || result.File == nil {
		t.Fatalf("default load = %+v", result)
	}

	if result.File.Path != "app.go" || len(result.File.Hunks) == 0 {
		t.Fatalf("default load file = %+v", result.File)
	}

	if result.Key != commitDetailKey(hashes[0], "app.go") {
		t.Fatalf("default load key = %q", result.Key)
	}
}

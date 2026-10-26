package main

import (
	"testing"
	"time"
)

func TestBranchLogCachePutGet(t *testing.T) {
	cache := newBranchLogCache(4)

	if _, ok := cache.Get("refs/heads/main"); ok {
		t.Fatal("empty cache returned a log")
	}

	cache.Put("refs/heads/main", commitLog{Branch: "main", Ref: "refs/heads/main"})
	cache.Put("refs/heads/feature", commitLog{Branch: "feature"})

	log, ok := cache.Get("refs/heads/main")
	if !ok || log.Branch != "main" || log.Ref != "refs/heads/main" {
		t.Fatalf("main log = %+v ok %v", log, ok)
	}

	if _, ok := cache.Get("refs/heads/missing"); ok {
		t.Fatal("missing ref returned a log")
	}
}

func TestBranchLogCacheEvictsOldest(t *testing.T) {
	cache := newBranchLogCache(2)
	base := time.Now().Add(-time.Hour)

	cache.Entries["old"] = cachedCommitLog{Log: commitLog{Branch: "old"}, FetchedAt: base}
	cache.Entries["new"] = cachedCommitLog{Log: commitLog{Branch: "new"}, FetchedAt: base.Add(time.Minute)}

	cache.Put("newest", commitLog{Branch: "newest"})

	if len(cache.Entries) != 2 {
		t.Fatalf("entries = %d, want 2", len(cache.Entries))
	}

	if _, ok := cache.Get("old"); ok {
		t.Fatal("oldest entry survived eviction")
	}

	if _, ok := cache.Get("new"); !ok {
		t.Fatal("newer entry was evicted")
	}

	if _, ok := cache.Get("newest"); !ok {
		t.Fatal("new entry was dropped on insert")
	}
}

func TestBranchLogCacheDefaultLimit(t *testing.T) {
	cache := newBranchLogCache(0)

	for i := 0; i < branchLogCacheLimit+4; i++ {
		cache.Put(string(rune('a'+i)), commitLog{Branch: string(rune('a' + i))})
	}

	if len(cache.Entries) != branchLogCacheLimit {
		t.Fatalf("entries = %d, want %d", len(cache.Entries), branchLogCacheLimit)
	}
}

func TestBranchLogCachePrune(t *testing.T) {
	cache := newBranchLogCache(4)
	cache.Put("fresh", commitLog{Branch: "fresh"})
	cache.Entries["stale"] = cachedCommitLog{
		Log:       commitLog{Branch: "stale"},
		FetchedAt: time.Now().Add(-branchLogCacheTTL - time.Minute),
	}

	cache.Prune(branchLogCacheTTL)

	if _, ok := cache.Get("stale"); ok {
		t.Fatal("stale entry survived Prune")
	}

	if _, ok := cache.Get("fresh"); !ok {
		t.Fatal("fresh entry was pruned")
	}

	cache.Entries["old-default"] = cachedCommitLog{
		Log:       commitLog{Branch: "old-default"},
		FetchedAt: time.Now().Add(-branchLogCacheTTL - time.Minute),
	}

	cache.Prune(0)

	if _, ok := cache.Get("old-default"); ok {
		t.Fatal("Prune(0) did not use the default TTL")
	}
}

func TestBranchLogCacheSnapshotRestore(t *testing.T) {
	cache := newBranchLogCache(4)
	cache.Put("refs/heads/main", commitLog{Branch: "main", Commits: []gitCommit{{Short: "abc1234"}}})
	cache.Put("refs/remotes/origin/main", commitLog{Branch: "origin/main"})

	snapshot := cache.Snapshot()
	if len(snapshot) != 2 || snapshot["refs/heads/main"].Branch != "main" {
		t.Fatalf("snapshot = %+v", snapshot)
	}

	delete(snapshot, "refs/heads/main")

	if _, ok := cache.Get("refs/heads/main"); !ok {
		t.Fatal("Snapshot shared its map with the cache")
	}

	restored := newBranchLogCache(4)
	restored.Restore(cache.Snapshot())

	log, ok := restored.Get("refs/heads/main")
	if !ok || log.Branch != "main" || len(log.Commits) != 1 || log.Commits[0].Short != "abc1234" {
		t.Fatalf("restored main log = %+v ok %v", log, ok)
	}

	if _, ok := restored.Get("refs/remotes/origin/main"); !ok {
		t.Fatal("restored cache lost a ref")
	}

	restored.Prune(branchLogCacheTTL)

	if _, ok := restored.Get("refs/heads/main"); !ok {
		t.Fatal("restored entries were not stamped fresh")
	}
}

func TestBranchLogCacheRestoreHonorsLimit(t *testing.T) {
	cache := newBranchLogCache(2)

	cache.Restore(map[string]commitLog{
		"a": {Branch: "a"},
		"b": {Branch: "b"},
		"c": {Branch: "c"},
		"d": {Branch: "d"},
	})

	if len(cache.Entries) != 2 {
		t.Fatalf("entries = %d, want 2", len(cache.Entries))
	}
}

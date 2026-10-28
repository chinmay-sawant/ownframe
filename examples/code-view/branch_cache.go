package main

import "time"

const (
	// branchLogCacheLimit is the default number of refs one cache holds.
	branchLogCacheLimit = 16
	// branchLogCacheTTL is how long a loaded branch log stays fresh.
	branchLogCacheTTL = 10 * time.Minute
)

// cachedCommitLog is one branch log with the time it was read.
type cachedCommitLog struct {
	Log       commitLog
	FetchedAt time.Time
}

// branchLogCache keeps recently read branch logs keyed by full ref. It is plain
// page-loop state: no mutex and no global instance.
type branchLogCache struct {
	Entries map[string]cachedCommitLog
	limit   int
}

// newBranchLogCache returns a cache bounded to limit refs. A limit of zero or
// less uses branchLogCacheLimit.
func newBranchLogCache(limit int) *branchLogCache {
	if limit <= 0 {
		limit = branchLogCacheLimit
	}

	return &branchLogCache{
		Entries: make(map[string]cachedCommitLog, limit),
		limit:   limit,
	}
}

// Get returns the cached log for ref, or false when ref is not cached.
func (c *branchLogCache) Get(ref string) (commitLog, bool) {
	if c == nil {
		return commitLog{}, false
	}

	entry, ok := c.Entries[ref]
	if !ok {
		return commitLog{}, false
	}

	return entry.Log, true
}

// Put stores one log as fresh now and evicts the oldest entries over the bound.
func (c *branchLogCache) Put(ref string, log commitLog) {
	c.ensure()
	c.Entries[ref] = cachedCommitLog{Log: log, FetchedAt: time.Now()}

	for len(c.Entries) > c.limit {
		c.evictOldest()
	}
}

// Prune drops every entry older than ttl. A ttl of zero or less uses
// branchLogCacheTTL.
func (c *branchLogCache) Prune(ttl time.Duration) {
	if c == nil {
		return
	}

	if ttl <= 0 {
		ttl = branchLogCacheTTL
	}

	cutoff := time.Now().Add(-ttl)

	for ref, entry := range c.Entries {
		if entry.FetchedAt.Before(cutoff) {
			delete(c.Entries, ref)
		}
	}
}

// Snapshot copies the cached logs for persistence. The commit slices are shared
// and read-only; restore them elsewhere with Restore.
func (c *branchLogCache) Snapshot() map[string]commitLog {
	if c == nil {
		return map[string]commitLog{}
	}

	out := make(map[string]commitLog, len(c.Entries))
	for ref, entry := range c.Entries {
		out[ref] = entry.Log
	}

	return out
}

// Restore adds logs read from a snapshot, stamped fresh now, and keeps the
// bound. Existing refs are replaced.
func (c *branchLogCache) Restore(logs map[string]commitLog) {
	c.ensure()

	now := time.Now()
	for ref, log := range logs {
		c.Entries[ref] = cachedCommitLog{Log: log, FetchedAt: now}
	}

	for len(c.Entries) > c.limit {
		c.evictOldest()
	}
}

// ensure makes a zero-value cache usable and restores the default bound.
func (c *branchLogCache) ensure() {
	if c.limit <= 0 {
		c.limit = branchLogCacheLimit
	}

	if c.Entries == nil {
		c.Entries = make(map[string]cachedCommitLog, c.limit)
	}
}

// evictOldest drops the entry with the earliest FetchedAt. The ref tie-break
// keeps the choice deterministic when two entries share a timestamp.
func (c *branchLogCache) evictOldest() {
	oldestRef := ""
	oldestAt := time.Time{}
	found := false

	for ref, entry := range c.Entries {
		if !found ||
			entry.FetchedAt.Before(oldestAt) ||
			(entry.FetchedAt.Equal(oldestAt) && ref < oldestRef) {
			oldestRef, oldestAt, found = ref, entry.FetchedAt, true
		}
	}

	if found {
		delete(c.Entries, oldestRef)
	}
}

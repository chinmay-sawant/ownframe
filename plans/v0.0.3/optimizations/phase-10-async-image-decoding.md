# Phase 10, async image decoding

> **Parent:** `plans/v0.0.3/optimizations/README.md` - async image decoding is optimization family 10.
> **Status:** not used. Current path decodes synchronously on first use.
> **Estimated effort:** roughly three to five days after the decode cost measurement in Phase 10.4. Touches `internal/replay/image_cache.go` first, then `internal/masking`, then callers.

---

## Overview

Ownframe paints a `layout.DisplayList` through the tile grid in
`internal/window`, and `internal/replay` turns each `DisplayOpImage` into an
`*ebiten.Image` the first time it is drawn. Today that translation happens
on the draw goroutine: `decodedImage` calls `masking.DecodeImage`, which calls
`image.Decode` on the raw payload, then uploads the result with
`ebiten.NewImageFromImage`. On a frame with many new images, or one large
image, the decode blocks `Draw` and stalls the whole tile pass. The cache
already prevents repeat decodes, but the first use of every distinct payload
still pays full decode cost inline.

This phase moves the decode off the draw path into a bounded worker pool.
The pool decodes bytes to a standard `image.Image` and hands the upload to
the draw goroutine, which stays the only owner of GPU resources. The cache
keeps its key, its eviction policy, and its "second hit check" shape; only
the producer of a decoded value changes.

Scope is ownframe. The engine side (`github.com/chinmay-sawant/blinkless`) is
not edited by these files. `layout.DisplayOp.ImageBytes` and `IsJPEG` are read
as they are. If payload metadata or a "decode hint" is needed later, that
landing happens in blinkless first and is pinned in `go.mod` after review.

A `[x]` row is code checked in on the date named. A `[ ]` row is open work.
`[~]` is not used here; deferred items move to a later phase ledger with a
reason and a gate.

## Where it lives

Current synchronous decode path:

- `internal/replay/image_cache.go` - `decodedImage`, the cache struct, and the
  eviction loop. Lines 31 to 70 are the whole path today.
- `internal/masking/image_decode.go` - `DecodeImage`, `ImageAlpha`,
  `SkipImageSize`. Lines 19 to 35 are the bytes to pixels rule.
- `internal/replay/image.go` - `drawImage`. Lines 16 to 21 call
  `decodedImage` per op per tile paint.
- `internal/bitmap/image.go` - `paintImage`. Lines 18 to 25 call
  `masking.DecodeImage` directly for the bitmap fallback path.
- `internal/replay/image_key.go` - `imageBytesKey`. Lines 13 to 28 build the
  FNV-1a payload key that both the cache and `sigImage` use.
- `internal/replay/chunk_sig_extra.go` - `sigImage`. Lines 29 to 37 fold the
  payload key into the display-list signature, so a cached decode is valid
  only while the payload identity is stable.

Related tests:

- `internal/replay/image_cache_test.go` - cache hit, miss, skip, eviction.
- `internal/bitmap/image_policy_test.go` - bad payload handling for bitmap.
- `internal/replay/image_bad_test.go` - bad payload handling for replay.

Cache keys and sizing also surface in `documentation/repaint.md` around the
decoded-image cache note (line 107 and 1892 of the folded docs).

## Phase 10.1, current synchronous decode inventory

Goal: write down exactly what the draw path does today, with line citations, so
the async version replaces it rather than duplicates it.

- [ ] `decodedImage` acquires a read lock, checks `images.entries[key]`, and
  returns the cached `*ebiten.Image` on hit (`image_cache.go:36-41`).
- [ ] On miss it calls `masking.DecodeImage(op)` on the draw goroutine, which
  runs `image.Decode` on `bytes.NewReader(data)` (`image_cache.go:42`,
  `image_decode.go:27`).
- [ ] A skip (nil op, empty payload, empty bounds) returns nil immediately and
  is not retried as a decode error (`image_cache.go:43-45`,
  `image_decode.go:20-33`).
- [ ] A decode error returns nil and is also not cached, so the next frame
  re-decodes (`image_cache.go:43-45`, `image_decode.go:28-29`).
- [ ] On success it acquires the write lock, re-checks the entry, uploads with
  `ebiten.NewImageFromImage(src)`, inserts, and evicts the oldest past
  `imageCacheLimit` (64) (`image_cache.go:47-69`).
- [ ] `drawImage` calls `decodedImage` every paint, including every tile paint,
  and skips the op when the result is nil (`image.go:17-20`).
- [ ] `paintImage` in the bitmap path calls `masking.DecodeImage` directly
  with no cache, then composites with `xdraw.BiLinear.Transform`
  (`bitmap/image.go:19-39`).
- [ ] The payload key is `imageBytesKey`, FNV-1a over the bytes plus a JPEG
  tag byte (`image_key.go:13-28`).
- [ ] `sigImage` folds the same key into the display-list signature, so a
  decode that changes the identity forces a repaint (`chunk_sig_extra.go:31-37`).
- [ ] There is no in-flight tracking today: a miss always decodes, even if a
  decode for that key is already running on another goroutine.
- [ ] There is no cancellation: a frame drop cannot abort a decode that has
  started.
- [ ] There is no notion of "pending" in the cache: callers get either a
  decoded image or nil, never "not yet, try again later."

## Phase 10.2, decode worker pool

Goal: move bytes to image decoding into a bounded pool of goroutines that never
touch Ebiten. Uploads remain on the draw goroutine.

- [ ] Add `internal/replay/decode_pool.go` (or a same-package file) holding a
  pool type with a fixed worker count derived from `runtime.NumCPU`, capped at
  a small constant, and a FIFO job queue.
- [ ] Define a `decodeJob` carrying the payload key, the raw bytes, the JPEG
  flag, and a channel or callback result slot.
- [ ] Workers decode using `image.Decode` on `bytes.NewReader` exactly as
  `masking.DecodeImage` does today, preserving the JPEG/PNG format handlers
  imported there.
- [ ] Workers return `image.Image` on success, an error on failure, and a bool
  for skip, matching `masking.DecodeImage`'s contract.
- [ ] Workers never call `ebiten.NewImageFromImage` or touch any Ebiten type.
- [ ] The pool is started lazily on first use and shut down through a lifecycle
  hook when the app quits.
- [ ] Pool depth and queue length are exposed as counters for the dump
  harness, not per frame allocations.
- [ ] A duplicate job for a key already in flight is coalesced into the same
  result slot, so a burst of tile paints for one image decodes once.
- [ ] Pool submit is non-blocking on the hot path when a worker is free;
  when full, the job is queued and the submitter returns immediately with a
  pending state rather than waiting.
- [ ] Worker exit is clean: no leaked goroutine, no lost result channel.
- [ ] Add `TestDecodePoolCoalescesDuplicateJobs` proving two submits with the
  same key share one result.
- [ ] Add `TestDecodePoolBoundedWorkers` proving worker count never exceeds the
  cap under load.

## Phase 10.3, cache integration

Goal: teach the cache to track pending decodes and publish results, without
changing the key, the eviction policy, or the second hit check.

- [ ] Add a `pending map[uint64]*decodeResult` field to `imageCache` to hold
  in-flight slots.
- [ ] Define `decodeResult` wrapping a `sync.WaitGroup` or a done channel, plus
  `src image.Image`, `err error`, and `skip bool`.
- [ ] On cache miss, check `pending` under lock; if present, return nil and
  record that the draw should retry next frame (or wait briefly).
- [ ] If absent, create a `decodeResult`, insert into `pending`, release the
  lock, and submit a job to the pool.
- [ ] When a result arrives, move it from `pending` to `entries` under the write
  lock, applying the same eviction rule (oldest past `imageCacheLimit`).
- [ ] Keep the second hit check: if another goroutine uploaded in the
  meantime, dispose the new upload and return the existing entry.
- [ ] Keep `entries` as the only place `*ebiten.Image` lives; results carry
  `image.Image`, never an Ebiten image.
- [ ] Decoded errors are cached for a short window so a bad payload is not
  re-decoded every frame; reuse the skip path for consistency.
- [ ] Cache hit latency stays a read lock only, unchanged from today.
- [ ] GPU upload (`ebiten.NewImageFromImage`) happens on the draw goroutine,
  inside `decodedImage`, when a completed result is consumed.
- [ ] Update `image_cache_test.go` expectations only where the async contract
  legitimately changes; do not weaken the eviction proof.
- [ ] Add `TestDecodedImagePendingReturnsNil` proving a first-use miss returns
  nil while the decode is in flight.
- [ ] Add `TestDecodedImagePublishesAfterDecode` proving the image appears once
  the pool result lands.
- [ ] Add `TestDecodedImageDedupesInFlight` proving a second caller during
  decode does not start a second job.

## Phase 10.4, decode cost measurement

Goal: capture numbers before any async code ships, per `documentation/` rules
that no optimization ships without a number.

- [ ] Use `go test -run` or a small benchmark in `internal/masking` that decodes
  a representative payload set: 1, 4, 16, and 64 images, small and large.
- [ ] Record `DecodeImage` ns per payload size in the ledger.
- [ ] Run `go run ./examples/departure-board -dump -pace 16ms` on a page with
  images and record worst frame and mean, before changing behavior.
- [ ] Identify the smallest payload count or size at which async wins; this
  becomes the threshold for enabling the pool.
- [ ] Do not enable async for payloads below the threshold; keep them on the
  synchronous fast path to avoid worker churn.
- [ ] Record the same command and metrics after the async pass and compare by
  ratio, not fixed counts.
- [ ] Update `documentation/performance.md` or its baseline section with the
  decode numbers and the threshold.

## Phase 10.5, cancellation and lifecycle

Goal: make async decode safe across page drops, redraws, and shutdown without
leaking goroutines or stale results.

- [ ] Thread a `context.Context` from the page or app into the decode request;
  cancel it on page drop or app quit.
- [ ] A cancelled job's result is discarded rather than published into the
  cache, but the worker still drains so the slot frees.
- [ ] Cache entries inserted for a dropped page are still evicted normally; do
  not introduce page-scoped caches unless measured.
- [ ] On app shutdown, `Wait` for in-flight decodes to finish or cancel them,
  then close the job queue.
- [ ] Shutdown completes within a bounded time even if a decode hangs; use a
  timeout and log a warning rather than blocking forever.
- [ ] In flight decodes for a key cancelled mid-flight clear the `pending`
  entry so a later request for the same key starts fresh.
- [ ] Tile paints that find a pending decode return nil and skip the op this
  frame, same as today's miss, so rendering stays stable.
- [ ] Add `TestDecodePoolCancellation` proving a cancelled job publishes nothing
  and frees its worker.
- [ ] Add `TestShutdownWaitsForInFlight` proving clean shutdown with pending
  decodes.
- [ ] Add `TestPendingClearedAfterCancel` proving a cancelled key can be
  re-requested.

## Phase 10.6, draw path integration

Goal: change callers as little as possible; `decodedImage` keeps its signature
so `drawImage` and tile paint code stay untouched.

- [ ] Keep `decodedImage(op *layout.DisplayOp) *ebiten.Image` returning the same
  public-ish helper signature used by `drawImage`.
- [ ] Add a helper `requestDecode(op)` that ensures a job is in flight for the
  op's key without waiting.
- [ ] On a miss where the pool is busy, ensure the job is queued rather than
  dropped, so the image arrives eventually.
- [ ] Keep the bitmap path (`paintImage`) deciding independently whether to use
  the cache or decode directly; do not silently change its error contract.
- [ ] If the bitmap path adopts the cache, route it through `decodedImage`
  too, and record the alpha compositing rule so the two paths stay unified
  (see `plans/v0.0.2/code-review` EXT-11 note in the sibling report).
- [ ] Gate the whole async path behind a build tag or a config flag until the
  closure rows pass, so a regression can be disabled without a revert.
- [ ] Keep `TakeDirty` and `Invalidate` working; async decode changes no
  invalidation behavior.
- [ ] Leave `sigImage` folding `imageBytesKey`, so a payload change still
  drives repaints.
- [ ] Preserve the skip and error handling so bad payloads do not stall frames
  (they resolve fast, and the cache keeps returning nil).

## Phase 10.7, concurrency safety

Goal: keep the cache's locking discipline and avoid data races under the race
detector.

- [ ] All `pending` reads and writes happen under `mu` (read lock for lookup,
  write lock for insert and publish).
- [ ] Result publication uses the write lock and the second hit check.
- [ ] The pool queue is guarded by its own mutex or channel, separate from the
  cache lock, to avoid lock ordering traps.
- [ ] No lock is held while a worker is submitted or while waiting on a result;
  submission happens outside any cache lock.
- [ ] Uploads happen after the write lock is released, or under a documented
  short hold; never while a worker is decoding the same key.
- [ ] Channel receives on the draw goroutine must not block the frame; use a
  non-blocking drain.
- [ ] Add a race-detector test that paints the same image from multiple tiles
  concurrently and reports zero races (run with `-race`).
- [ ] Run `go test -race ./internal/replay/...` before closing.

## Phase 10.8, test gates

- [ ] `internal/replay/image_cache_test.go` - add pending behavior rows and keep
  the eviction proof passing.
- [ ] Add `internal/replay/decode_pool_test.go` with coalescing, bounds, cancel,
  and shutdown cases.
- [ ] Add `internal/replay/async_image_test.go` covering publish, eviction after
  async, bad payload handling, and cache dedup under load.
- [ ] Keep `internal/replay/image_bad_test.go` and `bitmap/image_policy_test.go`
  passing unchanged.
- [ ] Run `make golden` after touching decode, cache, or paint code.
- [ ] Run `make test` (capped concurrency, `TEST_P=2 TEST_PARALLEL=2`) at
  session end.
- [ ] Run `go test -race ./internal/replay/...` once with the pool enabled.
- [ ] No test may call `image.Decode` directly against the pool's internal
  buffers; go through `masking.DecodeImage` to keep the contract single.

## Phase 10.9, closure and documentation

- [ ] Run `go run ./examples/departure-board -dump -pace 16ms` before and after
  with the same page set and record worst frame and mean in this ledger.
- [ ] Confirm the improvement is ratio-based against the baseline row, not a
  fixed operation count.
- [ ] Update `documentation/repaint.md` describing the cache miss, pending
  state, and worker pool contract.
- [ ] Update `documentation/performance.md` with decode numbers and the async
  threshold.
- [ ] Flip this phase's status in `README.md` from "not used" to "partial" or
  "yes" only after the gate passes.
- [ ] Leave every unproven row `[ ]`; do not close from plan prose.
- [ ] Run `make lint` once, clean, before the PR.
- [ ] Run `make claim-scan` once before the PR.

## Exit criteria

- Images decode off the draw goroutine when the threshold is met, while small
  payloads stay on the synchronous fast path.
- The cache keeps returning either a decoded image or nil on a miss; callers
  are unchanged aside from a nil retry next frame.
- No duplicate in-flight decode for one key.
- Cancelled or dropped work does not leak goroutines, stale results, or
  pending entries.
- Shutdown is bounded even with downloads in flight.
- `image_cache_test.go`, the new pool tests, and `image_bad_test.go` pass.
- `go test -race ./internal/replay/...` passes with the pool enabled.
- `make golden` and `make test` pass.
- Before and after numbers are recorded with the same dump command.

## Risks

- Worker count too high pins CPU and hurts the tile pass; too low starves
  large pages. Count and cap must come from measurement, not a guess.
- Queue depth growth on a image heavy page can lag indefinitely; either grow
  the threshold or prioritize keys already needed this frame.
- GPU upload still happens on the draw goroutine, so a finished decode can
  stall the frame at upload time. Keep decode cost separate from upload cost
  in the measurements.
- Coalescing bugs can publish one result to two different keys; the key must
  be carried on the job, not looked up by pointer.
- Cancelled jobs that still publish a late result can overwrite a newer
  entry; guard with a cancellation flag checked under the write lock.
- Race detector runs are slower and may time out on CI; keep the race test
  small and targeted.
- Engine-side payload metadata is not owned here; any need for a decode hint
  requires a blinkless PR and a pin bump, handled separately.
- The bitmap fallback path may keep decoding synchronously even after replay
  goes async, which splits the contract; track this as an explicit row, not a
  silent divergence.

## Open questions

1. Should the pool decode to `image.Image` only, or should it also run
   color-space normalization before upload?
2. Should small payloads bypass the pool entirely (fast path), or always
   enqueue to keep one code path?
3. Should results be pushed to the draw goroutine through a channel, or polled
   from the cache on the next frame?
4. Should page drops cancel in-flight decodes eagerly, or let them finish and
   evict later?
5. Should decode errors be cached for a bounded time, or retried every frame?
6. Is a single global pool fine, or should browsers-style per-page pools be
   considered later?

EOF
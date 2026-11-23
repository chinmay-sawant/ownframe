# v0.0.4 - Go Review Hardening Plan

> **Parent:** [plans/v0.0.4/README.md](../README.md) (and [maintenance ledger](../../README.md))
> **Branch:** `chore/v0.0.4-go-review-hardening`
> **Baseline:** `master` at `cf529aba` (pinned to `blinkless` `b2fe5ba5105e`, 2026-11-23)
> **Status:** Code complete at `17caf6c9` (pinned to `blinkless` `master` `8bbafad`); `[~]` rows need a Windows host/CI runner or Android bind job; remaining gates need full `make test`/`make test-race`, benchmarks, and CI

---

## Overview

This directory holds the phase-wise execution checklist produced from the cross-checked 10-part Go code review of `ownframe` (`cf529aba`) and `blinkless` (`b2fe5ba5105e`). Every row below corresponds to a source-verified defect, concurrency hazard, security gap, hot-path regression, or CI/module hygiene issue.

Earlier structural cleanup in [chore-improvements](../chore-improvements/README.md) addressed repository footprint, doc alignment, and package consolidation. This plan targets **runtime safety, security, concurrency correctness, per-frame performance, and platform verification**.

## Phase Summary

| Order | Phase | Scope | Tasks | Closure Gates | File |
| :---: | :--- | :--- | :---: | :---: | :--- |
| 1 | **P0: Safety, Security & Concurrency** | Win32 UIA COM UAF/GC & subclass races, WSL PowerShell injection, `-web` theme XSS, password Copy/Cut, `web` & `page` data races, Windows clipboard OS thread lock, WSLg X11 error handler | **12** | **3** | [phase-01](phase-01-p0-safety-security-and-concurrency.md) |
| 2 | **P1: Correctness & Security Hardening** | AT-SPI coalescing, UIA `ctxDone` detach, `emoji` parser fixes, stale display caches, `NaN` guards, `fetch` SSRF/pool/sentinels, `CheckDOM` expansion, `web` lock-free socket writes & SSE cap, `pprof` timeout, `textrun` font features | **13** | **3** | [phase-02](phase-02-p1-correctness-and-hardening.md) |
| 3 | **P1: Hot-Path Performance & GPU Lifecycle** | Cache `isWSL()`, O(1) `textRasterCache` LRU, pre-compiled motion easings, `placeUIA()` throttle, bounded backdrop/blend/rendererPool buffers, full `runWindow` GPU disposal | **9** | **3** | [phase-03](phase-03-p1-hot-path-performance-and-gpu-lifecycle.md) |
| 4 | **P2: CI Coverage, Modules & API Hygiene** | Windows `amd64` UIA/window CI, nil-`Page`/context consistency, pre-v1.0 API/coupling tracking, `host` interface assertions, synonym/lint cleanup, `x/perf` & `go.mod` hygiene, hot-reload/IPC/crash fixes, gated `engineMu` removal | **11** | **3** | [phase-04](phase-04-p2-ci-modules-and-api-hygiene.md) |
| | **Total** | | **45** | **12** | |

---

## Execution & Evidence Rules

- Use consistent row states across all phase files:
  - `[ ]` open / not yet implemented or not proven;
  - `[x]` implemented and proven with commit hash, exact command, exit status, and observed result beside the row;
  - `[~]` intentionally deferred or blocked on an external host/upstream tag, with owner boundary and reconsideration gate.
- Write a failing regression test (or `-race` reproducer) before fixing each bug wherever headless testing is possible.
- Never mark a Windows, WSLg, or macOS native row `[x]` from Linux-only compilation; record the actual host or CI runner evidence, or leave open/blocked explicitly.
- Keep `AGENTS.md` invariants intact: `ownframe.Page` stays wrapped with an unexported `*page.Page`, no unapproved root `type X =` aliases, no new `examples/` imports of `ownframe/internal/`, and clipboard tests keep `clipboard.UseMemory(true)`.
- Every code phase requires `gofmt`, `make lint`, `make test`, and `make test-race`. Changes touching paint, replay, motion, or windowing also require `make graphics-check`.

## Cross-Repo Dependencies & Guardrails

- **Paired Blinkless plan:** `blinkless` branch `chore/v0.1.6-go-review-hardening` (`plans/v0.1.6/go-review-hardening/`).
- **Do not remove `internal/render/engineMu` prematurely:** `engineMu` (`internal/render/engine.go:5-12`) currently shields `ownframe` from two confirmed `blinkless` races (`Node.EnsureAttrs()` and `prepare.SheetCache` in `css.Relayout`). Keep `engineMu` in place until `blinkless` Phase 01 lands and `ownframe` bumps its `blinkless` pin (`HYG-09` in Phase 04).
- **Order of execution:** Finish Phase 01 before Phase 02–04 so memory-safety, injection, XSS, and data-race bugs do not stay open on `master`.

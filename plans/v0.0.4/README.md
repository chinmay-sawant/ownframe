# v0.0.4 - Interaction, accessibility, and rendering plans

> **Parent:** [maintenance ledger](../README.md)
> **Status:** All four tracks are complete. Their checklist rows and retained validation evidence are linked below.
> **Estimated effort:** estimate each implementation slice after its investigation gate. Native validation and the engine release have external dependencies.

---

## Overview

Improve everyday application behavior through measured rendering work, Linux
accessibility actions, reusable component interactions, and CSS conformance.
This directory records the four tracks agreed after reviewing `085ba7ea`.
Creating these plans does not implement their checklist rows.

## Executive Summary

| Order | Track | Acceptance | Canonical checklist |
|---|---|---|---|
| 1 | Documentation and performance | **Accurate support claims and measured workload improvements** | [01](01-documentation-and-performance.md) |
| 2 | Linux accessibility actions | **A live client can focus and activate controls** | [02](02-linux-accessibility-actions.md) |
| 3 | Component interactions | **Applications install standard behavior without repeating its mechanics** | [03](03-component-interactions.md) |
| 4 | Engine CSS conformance | **Height queries and sticky behavior pass geometry and integration checks** | [04](04-css-conformance-and-engine-adoption.md) |

The linked files own their task rows. This index carries no duplicate task
checklist. Start the documentation correction and baseline capture first.
Prioritize accessibility correctness before performance optimization. The
performance implementation follows its profiles and can land independently.

## Phase 1: Establish evidence and ownership

### 1.1 Execution rules

Use `[ ]` for unstarted or unproven work and `[x]` only after implementation and
the stated proof pass. Do not use partial or deferred states in these tracks.
Keep unfinished rows open, name any missing input, and finish the current phase
before starting the next. Report back after each phase. The parent maintenance
ledger retains its existing status syntax.

Record the source commit, exact command, exit status, environment, and retained
artifact beside each completed validation row. Evidence paths must remain
available to reviewers. Ignored `temp/` captures can support investigation, but
release evidence needs a committed record or a durable CI artifact link.

An implementation phase requires `make lint` and `make test`. Paint or window
changes also require `make graphics-check`. Platform observations must identify
the actual host. Xvfb results do not close live AT-SPI or device validation.

## Dependencies

- [01](01-documentation-and-performance.md) owns the immediate flex-absolute
  documentation correction and the performance baseline. Track 04 owns later
  documentation changes for newly implemented CSS behavior.
- [02](02-linux-accessibility-actions.md) reuses the existing page action API.
  It does not claim Windows, macOS, or Android accessibility adapters.
- [03](03-component-interactions.md) retains host-owned component values and
  uses existing public input contracts where they suffice.
- [04](04-css-conformance-and-engine-adoption.md) requires separately scoped
  Blinkless work and a published engine revision before OwnFrame adoption.
- Root `ownframe.Page` stays wrapped. Preserve the public API and example
  internal-import boundaries described in `AGENTS.md`.

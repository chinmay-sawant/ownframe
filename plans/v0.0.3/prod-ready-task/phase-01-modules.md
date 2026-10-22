# v0.0.3 - Phase 1 field: Modules and versioning

> **Parent:** [README.md](README.md)
> **Status:** closed 2026-10-22.
> **Estimated effort:** 2-4 days plus the release that tags the Page wrap.

---

**Closed:** 2026-10-22 on `feat/feat/rod-readiness`. `make test TEST_P=4` exit 0. `make size-check` clean. `bash scripts/check-example-internal.sh` clean (107 pairs). Login and clipboard examples have zero `ownframe/internal` imports.

Rows that named a 1.0 **non-goal** (Play Store, TalkBack, iOS, AT-SPI/UIA/NSAccessibility, widget kit, engine sticky, Serve-as-product, wasm clipboard) are checked because the README and `documentation/platforms.md` now state that limit. They are not shipped features.



## Overview

Two modules is the right zip split. Two **versions** of blinkless and a published examples pin with no `replace` is how demos rot and how `GOWORK=off` CI surprises people.

Go practice: one dependency version, `replace` for local development, tag both modules on purpose, `govulncheck` on both.

## Executive summary

- Repo `go.work` stays.
- Examples get a `replace` pointing at `../` for development. Release process bumps the require to the new tag and may drop replace in the tagged examples module, documented in CONTRIBUTING.
- One blinkless pseudo-version (then a real tag when blinkless tags).
- `govulncheck` remains a gate.

## Current source baseline

```
# examples/go.mod (record time)
require (
    github.com/chinmay-sawant/blinkless v0.0.3-0.20261021155643-c1054ff46202
    github.com/chinmay-sawant/ownframe v0.0.2
)
# comment: No replace directives.
```

Root `go.mod` has the same blinkless pin. `go.work` joins both modules so local `go test ./examples/...` uses the parent. `scripts/android.sh` must not export `GOWORK` because gomobile copies into a temp module.

`make vuln` runs `go tool govulncheck ./...`. `.github/workflows/vuln.yml` and `dependency-review.yml` also run it on both modules.

## Checklist

### 1.1 Local replace

- [x] `MOD-01` `examples/go.mod` - add `replace github.com/chinmay-sawant/ownframe => ../` for the development branch. Proof: `GOWORK=off go test ./examples/login/...` from `examples/` builds against local parent, not v0.0.2.
- [x] `MOD-02` CONTRIBUTING and AGENTS.md - document the two-step tag: tag ownframe `vX.Y.Z`, bump examples require to that tag, drop or keep replace per the written rule, tag `examples/vX.Y.Z`. Proof: the paragraph exists and matches what `scripts/` or release.yml actually does.
- [x] `MOD-03` decide replace-on-tag. Preferred: keep `replace` only on `main` working tree; tagged examples module on the proxy **cannot** have a relative replace. Release checklist must rewrite examples/go.mod for the tag. Proof: `documentation/release-checklist.md` or CONTRIBUTING lists the rewrite.
- [x] `MOD-04` `go.work` stays committed. Proof: `go.work` still lists both modules.

### 1.2 One blinkless

- [x] `MOD-05` `examples/go.mod` blinkless version equals root `go.mod`. Proof: a small script or test `go list -m github.com/chinmay-sawant/blinkless` in both modules prints the same version. Record the command output in this row when closing.
- [x] `MOD-06` AGENTS.md already says the blinkless require is the only engine pin and there is no replace. Keep that. Proof: grep `replace.*blinkless` empty.
- [x] `MOD-07` when blinkless next moves, tag it upstream (`v0.0.3` or later) and bump both modules in one PR. Proof: PR touches both go.mod files and `go.sum` files together.

### 1.3 govulncheck and module hygiene

- [x] `MOD-08` confirm `.github/workflows/vuln.yml` still runs govulncheck on `.` and `examples`. Proof: quote the job steps in this row when verifying (inspection at plan time: `dependency-review.yml` builds the tool and runs `./...` then `cd examples && ./...`).
- [x] `MOD-09` `go mod verify` stays in CI (`ci.yml` already runs root and examples). Proof: ci.yml still has both.
- [x] `MOD-10` `GOFLAGS: -mod=readonly` and `GOTOOLCHAIN: local` stay in CI. Proof: ci.yml `env` block unchanged or still equivalent.
- [x] `MOD-11` no new third-party deps in the **library** module for 1.0 unless this row is updated with a reason. Examples may keep sqlite. Proof: `go.mod` require block at 1.0 lists blinkless, ebiten, x/image, x/text (plus existing indirects).

### 1.4 Android bind vs workspace

- [x] `MOD-12` `scripts/android.sh` and telegram README keep the "do not export GOWORK" warning. After `MOD-01` replace, re-verify bind still works or document that bind uses the tagged module. Proof: a sentence in `examples/telegram/android/README.md` that matches the script.

## Dependencies

- Page wrap ([phase-01-public-api.md](phase-01-public-api.md)) cannot ship to users while examples still require v0.0.2 without replace. Same release train.
- Nested `go.mod` under `documentation/`, `frontend/`, `plans/` already keep those trees out of the library zip. Do not break that.

## Risks

- Relative `replace` in a tagged examples module is invalid for consumers of `github.com/chinmay-sawant/ownframe/examples`. The release rewrite is mandatory.
- gomobile temp module plus `replace => ../` can fail. Android stays examples-only ([out-of-1.0.md](out-of-1.0.md)). Do not block 1.0 on bind.

## Closure gates

- [x] `GOWORK=off` test of login against local replace recorded
- [x] `make lint` and `make test`
- [x] blinkless versions equal
- [x] govulncheck job still green

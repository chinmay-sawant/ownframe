## Summary

Prepare the v0.0.2 release: pin the tagged `github.com/chinmay-sawant/blinkless v0.0.2` in both modules, replace `plans/v0.0.2/release-notes.md` with the v0.0.2 edition, fix stale engine references across ten documentation guides, rebuild the published `docs/` site, and land the fact-check corrections to the note. No library source changes.

## Motivation / context

- Plans: `plans/v0.0.2/release-notes.md` is the release body; the cycle is tracked in `plans/v0.0.2/README.md` and `PHASES.md`.
- The engine was renamed from `gowkhtmltopdf` to `blinkless` and tagged `v0.0.2`; the modules still pinned the pseudo-version `v0.0.0-20261013163225-c085b3423260`, and several guides still named old paths or behaviour.
- No issue is linked; this repository keeps plans under `plans/` instead of an issue tracker.

## Changes

### Module pins

- `go.mod`, `examples/go.mod`, and the sums move the engine require to `github.com/chinmay-sawant/blinkless v0.0.2`; `go.work.sum` gains the workspace entries.

### Release note

- Replace `plans/v0.0.2/release-notes.md` with the v0.0.2 edition: highlights, migrating-from-v0.0.1 notes, area sections (window, repaint, devtools, input, Android, security), and a "What's changed" summary.
- Correct the note's fact-check findings in a follow-up commit: commit and diff counts, the port range, `examples/readme.md` coverage, APK checksum coverage, and the SQLite demo count.

### Documentation and site

- Refresh engine references and shipped behaviour in ten guides: `documentation/android-runtime.md`, `devtools.md`, `emoji.md`, `features-examples.md`, `features.md`, `performance-baseline.md`, `performance.md`, `repaint.md`, `screen.md`, and `window.md`.
- `AGENTS.md`: engine pin note updated to `v0.0.2`.
- Rebuild `docs/` with the new guide content; `index.html` and the hashed asset names change.

## Impact

| Area | Impact |
|------|--------|
| **Performance** | No performance claim; pins, notes, docs, and the site only. |
| **Memory** | None measured. |
| **Behavior / correctness** | Documentation only; the engine require moves to its tagged `v0.0.2` release. |
| **API / CLI** | None. |
| **Dependencies** | `github.com/chinmay-sawant/blinkless v0.0.2` replaces the pseudo-version in both modules. |
| **Binary size / build time** | Not measured; no library source change. |

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| None | No migration needed. |

## Test plan

- [x] `make build` (`go vet -p 1 ./... ./examples/...`) on the branch tip.
- [x] `make test` (`go test -p 1 ./... ./examples/...`) on the branch tip.

### Commands

```sh
make build
make test
```

## Screenshots / sample output

None. The release body is `plans/v0.0.2/release-notes.md`.

## Related issues

None. No issue is linked to this work.

## PR metadata checklist (author)

- [x] Self-assigned with `--assignee "@me"` when opened.
- [ ] Labels: none applied to this PR.
- [x] Related issues checked; none are linked.
- [x] This PR record is committed under `plans/v0.0.2/pr/pr-release-v0.0.2.md`.

## Follow-ups (out of scope)

- Examples module follow-up: require `ownframe v0.0.2`, drop the local replace, `go mod tidy`, and tag `examples/v0.0.2`.
- Refresh `PHASES.md` after the tag (close the pending section, move the pins).
- Add `digital-wellbeing` to `examples/android-host/APK-SHA256SUMS` and its `APPS.md` entry; six of seven APKs are covered today.
- `examples/lg-remote` defaults to port 8131, the same as `examples/perf-benchmarks/normal`.

## Reviewer checklist

- [ ] Behavior matches the summary and test plan.
- [ ] No unrelated changes in the diff.
- [ ] Public API changes are documented.
- [ ] PR has an assignee and labels.
- [ ] Related issues use the correct closing keywords.
- [ ] No secrets or generated artifacts were committed.
- [ ] Diff statistics were generated from the merge base.

## Diff stat by extension

Generated with `bash scripts/pr-diff-stat.sh master` after the final commit.

| Extension | Files | Insertions | Deletions |
| --- | ---: | ---: | ---: |
| `.html` | 1 | 1 | 1 |
| `.js` | 2 | 24 | 26 |
| `.md` | 13 | 218 | 85 |
| `.mod` | 2 | 2 | 2 |
| `.sum` | 3 | 75 | 4 |
| **Total** | **21** | **320** | **118** |

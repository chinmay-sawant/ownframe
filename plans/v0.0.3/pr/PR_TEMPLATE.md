# PR template (ownframe)

Copy everything below into the GitHub PR body. Save the filled copy under `plans/v0.0.3/pr/pr-<slug>.md` before running `gh pr create`.

## Summary

1-3 sentences: what changed and why.

## Base and head

- Base: `master`
- Head: `fix/miscellaneous-fixes-and-lint-fix`

## Changes

### Core library
-

### Window, replay, render
-

### Web, fetch, IPC, OS integration
-

### Examples
-

### CI, supply chain, docs
-

## Security audit 3 (GEN-1 to GEN-50)

All boxes in `plans/v0.0.3/security/SECURITY_AUDIT_3.md` are [x]. One line per group.

## Test plan

- [ ] `go vet -p 1 ./... ./examples/...`
- [ ] `go test -p 1 -run Sec ./... ./examples/...`
- [ ] Targeted `go test -count=1` on touched packages
- [ ] `bash scripts/check-file-size.sh`
- [ ] `make lint` (when clean)

## Diff stat

Run `git diff --stat origin/master...HEAD` and paste the tail plus the extension table.

## Related

- Plans: `plans/v0.0.3/security/SECURITY_AUDIT_3.md`
- Branch: `fix/miscellaneous-fixes-and-lint-fix`
- Commit: `06653c4`

## Checklist

- [ ] Self-assigned
- [ ] Labels applied
- [ ] No secrets committed
- [ ] Filled body saved under `plans/v0.0.3/pr/`

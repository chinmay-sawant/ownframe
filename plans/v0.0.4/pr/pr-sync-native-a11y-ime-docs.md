# docs: reconcile Windows UIA, Android TalkBack, and Windows IME status

Filled copy of [`skills/PR/PR_TEMPLATE.md`](../../../skills/PR/PR_TEMPLATE.md).

- Base: `master`
- Head: `docs/sync-native-a11y-ime-status`
- Release cycle: v0.0.4

---

## Summary

Reconciles remaining stale paragraphs in `README.md`, `documentation/platforms.md`, and `documentation/accessibility-native.md` with the Windows UI Automation provider, Windows Japanese IME validation, and Android TalkBack bridge shipped in PRs #95 and #96.

---

## Motivation / context

- Plans: `plans/v0.0.4/pr/pr-sync-native-a11y-ime-docs.md`
- Issues: none. Follows `chinmay-sawant/ownframe#95` and `chinmay-sawant/ownframe#96`.

---

## Changes

### Documentation (`README.md`, `documentation/platforms.md`, `documentation/accessibility-native.md`)

- **`README.md`:** Updated the overview and documentation summary paragraphs to record the preview Windows UI Automation provider (`amd64` and `arm64`, validated with NVDA and Narrator), the Android TalkBack bridge (`androidaccess` and `AccessibilityBridge`), and Linux AT-SPI `EditableText`, leaving macOS VoiceOver (`NSAccessibility`) as the remaining unshipped screen-reader adapter.
- **`documentation/platforms.md`:** Updated the intro paragraph, current vs planned table, Android native section, and capability matrix Desktop `IME` and `A11y` rows to reflect the shipped Windows UIA provider, Android TalkBack bridge, and Windows Microsoft Japanese IME validation.
- **`documentation/accessibility-native.md`:** Updated the `v1.0 accessibility policy` section so it matches the top-of-file status for Linux AT-SPI, Windows UIA, and Android TalkBack.

---

## Impact

| Area | Impact |
|------|--------|
| **Performance** | None (documentation only). |
| **Memory** | None. |
| **Behavior / correctness** | Eliminates contradictory accessibility and IME status statements across `README.md`, `platforms.md`, and `accessibility-native.md`. |
| **API / CLI** | None. |
| **Dependencies** | None. |
| **Binary size / build time** | None. |

---

## Breaking changes / migration

| Item | Migration |
|------|-----------|
| None | - |

---

## Test plan

- [x] `make docs-check`
- [x] `make plans-check`
- [x] `make examples-check`
- [x] `make tools-check`
- [x] `make repo-size-check`

### Commands

```sh
make docs-check plans-check examples-check tools-check repo-size-check
```

---

## Screenshots / sample output

```text
docs-check: 39 snippets compiled; 12 external illustrations listed explicitly
plans-check: 24 rows with distinct states
examples-check: maintained teaching set: bind, desktop-cat, frontend, forms, list, login, telegram; 58 experimental/support roots
repo-size-check: 10 allowlisted >5 MiB files; 0 new violations
```

---

## Related issues

- Relates to `chinmay-sawant/ownframe#95` and `chinmay-sawant/ownframe#96`.

---

## PR metadata checklist (author)

- [x] Self-assigned (`--assignee @me`)
- [x] Labels applied (`documentation`)
- [x] Related issues filled (`#95`, `#96`)
- [x] Filled body committed under `plans/v0.0.4/pr/pr-sync-native-a11y-ime-docs.md`

---

## Follow-ups (out of scope)

- None.

---

## Reviewer checklist

- [x] Behavior matches summary and test plan
- [x] No unrelated changes in diff
- [x] Public API / CLI changes documented
- [x] New rules have fixture coverage when applicable
- [x] PR has assignee and labels
- [x] Related issues use correct Closes/Relates keywords
- [x] No secrets or generated artifacts committed
- [x] Diff-stat-by-extension table pasted at the bottom

---

## Diff stat by extension

| Extension | Files | Insertions | Deletions |
| --- | ---: | ---: | ---: |
| `.md` | 3 | 31 | 24 |
| **Total** | **3** | **31** | **24** |

# Phase 03: diff

Goal: show a real unified diff, with file statuses, line numbers, and the
changed words marked. The same parser serves the bundled fixture and a live
repository.

## Where it lives

- `diff.go` parses `git diff`, pairs deleted and added lines, and builds
  unified and split rows.
- `git.go` reads a live working tree for `-root`.
- `fixtures/sample.diff` and `sample_diff.go` provide the bundled change set.
- `components/diff.css` draws the rows and the emphasis.
- `diff_test.go` covers the parser, emphasis, split pairing, untracked files,
  and hunk labels.

## Design

1. The parser reads `diff --git`, `new file mode`, `deleted file mode`,
   `rename from` and `rename to`, `---`/`+++`, `Binary files`, `@@`, and the
   three line prefixes.
2. Status comes from the headers: A for new, D for deleted, R for rename, U for
   untracked, M otherwise. Counts come from the hunk lines.
3. Deleted and added runs pair by index. The common prefix and suffix are
   trimmed, then the middle is expanded to word boundaries. The two ranges get
   `em-d` and `em-a`.
4. Unified rows show old and new numbers, a sign column, and one cell.
   Split rows show a pair of numbers and cells, with padding on the short side.
5. Hunk rows show the real range, for example `@@ -10,26 +10,92 @@`, plus the
   enclosing symbol when git supplied one.
6. Live git uses `git -C <root> diff --no-color --no-ext-diff HEAD` and
   `git ls-files --others --exclude-standard -z` with a three-second budget.
   A repository without a first commit falls back to the working tree and
   index diffs.
7. The bundled fixture is the default. It is fixed data and says so in the
   sidebar. A clean live tree says "working tree is clean" instead of replacing
   the result with the fixture.

## Checklist

- [x] Parse files, hunks, line numbers, and statuses.
- [x] Handle new, deleted, renamed, binary, and untracked files.
- [x] Pair deleted and added lines and mark the changed words.
- [x] Build unified rows with old and new gutters.
- [x] Build split rows with paired context and padding.
- [x] Show per-file addition and deletion counts.
- [x] Add a source-control list with status letters and totals.
- [x] Add unified and split toggles.
- [x] Add previous and next hunk actions.
- [x] Read a live tree for `-root` with a timeout.
- [x] Keep the fixture for the embedded sample.
- [x] Refuse a path that escapes the scanned root.
- [ ] Add a context-expand action between hunks.
- [ ] Add a rename arrow and old path in the file row.

## Gate

`TestSampleDiffShape` confirms seven files and `+562 -109`.
`TestWordEmphasis` confirms the deleted and added ranges.
`TestSplitPairing` confirms the paired rows. `TestUntrackedChange` confirms an
all-added hunk.

## Risks

- A diff path with a space or a quote can defeat the `diff --git` path split.
  The `---` and `+++` lines are the fallback; add a fixture before relying on
  the quoted form.
- A very large added file produces a large row list. The window keeps the DOM
  small, but the parser still holds every line. Cap per-file rows in phase 07.
- The three-second git budget covers a normal repository, not a cold network
  path. A failure becomes a toast, not an empty screen.

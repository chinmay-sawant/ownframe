# Code review example

> **Status:** phases 01 to 04 and phase 09 are in `examples/code-view`.
> Phases 05 to 08 are substantially landed; their remaining items are marked
> open in each checklist.
> **Parent:** [v0.0.3 plans](../../../README.md)
> **Example:** [examples/code-view](../../../../examples/code-view/README.md)

The example is a dark multi-pane viewer for a directory. The explorer lists
files, tabs open them, and Source Control shows a diff. It uses one HTML
template, the public `ownframe` API, and Go data built from the scanned files.

## The rule

The example may contain HTML, CSS, Go data, and handlers. It may not contain a
script tag, a canvas, or an import of `ownframe/internal/`. Highlighting runs in
Go before `SetData`; the page prints plain spans.

The presenter is one native window. "Window" in this example means an editor
pane or a tab. ownframe has no second OS window.

## What ships

- The shell: title bar, activity bar, explorer, tabs, breadcrumbs, editor,
  details drawer, status bar, and toast.
- The workspace: `testdata/sample`, an embedded project with Go, Python,
  Markdown, CSS, JSON, HTML, YAML, and shell files, plus
  `fixtures/sample.diff`.
- The syntax: Go, Python, JavaScript and TypeScript, Markdown, CSS, JSON, HTML,
  YAML, shell, and plain text. JavaScript is covered by tests; the sample
  carries no script file.
- The diff: unified and split rows, file statuses, word emphasis, hunk
  navigation, wrapped cells, async commit patches, and live `git diff HEAD`
  with `-root`.
- The live workspace: a debounced fsnotify rescan, git metadata watch, scan
  budget, and session restore.
- Two editor groups with their own tab lists, edit/save, in-file find, and
  workspace content search.
- The windowing: fixed chrome, `SetScrollWindow`, 20 px rows, a windowed
  explorer with wheel/rail/keyboard scrolling, and a sidebar history split.
- The details drawer: language, size, status, and outline.
- The Markdown preview: tables, HTML text, Mermaid text diagrams, and reflow.
- Developer tools on by default.

## Phase index

| # | Phase | Status | Depends on | File |
|---|---|---|---|---|
| 01 | Shell and workspace | Landed | none | `phase-01-shell-and-workspace.md` |
| 02 | Syntax | Landed | 01 | `phase-02-syntax.md` |
| 03 | Diff | Landed | 01 | `phase-03-diff.md` |
| 04 | Scrolling and chrome | Landed | 01 | `phase-04-scrolling-and-chrome.md` |
| 05 | Search and command palette | Mostly landed | 01, 02 | `phase-05-search-and-palette.md` |
| 06 | Live workspace | Mostly landed | 01, 03 | `phase-06-live-workspace.md` |
| 07 | Editor depth | Partly landed | 02, 04 | `phase-07-editor-depth.md` |
| 08 | Accessibility and mobile | Partly landed | 01, 04 | `phase-08-accessibility-and-mobile.md` |
| 09 | Docs and gates | Landed | all | `phase-09-docs-and-gates.md` |

## Master checklist

### Phase 01: shell and workspace

- [x] Fixed shell at 1500x900 with an 840x700 minimum. Evidence:
      `TestShellMetrics`, `TestResponsiveLayout`.
- [x] Embedded demo workspace with 17 files and a fixture diff. Evidence:
      `TestSampleWorkspace`, `TestSampleDiffShape`.
- [x] Explorer tree with fold state, language badges, change letters, and a
      flat filtered list. Evidence: `TestFilter`.
- [x] Tabs with an eight-file cap, close, and a welcome panel. Evidence:
      `TestOpenAndCloseTabs`.
- [x] Details drawer with metadata and an outline. Evidence: `TestDetails`.
- [x] Developer tools start on and the status chip toggles them. Evidence:
      `TestDevTools`.
- [x] No script file in the bundled workspace. Evidence:
      `TestSampleHasNoJavaScript`.

### Phase 02: syntax

- [x] Shared whole-file lexers for Go, Python, and JavaScript. Evidence:
      `TestGoFunctionReceivers`, `TestPythonTripleQuotedString`,
      `TestJavaScriptTokens`.
- [x] Markdown blocks, inline emphasis, and fenced code. Evidence:
      `TestMarkdownStructure`.
- [x] CSS, JSON, HTML, YAML, and shell tokenizers. Evidence:
      `TestMarkupTokens`.
- [x] Every token is a plain string and the template prints it as a span. No
      `template.HTML`. Evidence: `syntax.go`, `components/editor.html`.
- [x] Token text preservation for every language. Evidence:
      `TestHighlightLineAlwaysHasText`, `syntax_test.go`.

### Phase 03: diff

- [x] Parse `git diff` into files, hunks, and numbered lines. Evidence:
      `TestHunkNumbersAndText`, `TestSampleDiffShape`.
- [x] File statuses for modified, added, deleted, renamed, and untracked.
      Evidence: `TestSampleDiffShape`, `TestUntrackedChange`.
- [x] Word-level emphasis for paired deleted and added lines. Evidence:
      `TestWordEmphasis`.
- [x] Unified and split row builders with paired context. Evidence:
      `TestSplitPairing`.
- [x] Live git with a three-second budget and a bundled fallback. Evidence:
      `git.go`, `TestLiveRoot`.
- [x] Hunk navigation, refresh, and a clean-tree state. Evidence:
      `TestModesAndDiff`, `TestCleanChangeSet`.
- [x] Sidebar history pane with expandable commits, hover messages, and a
      lazy one-file patch. Evidence: `TestHistoryPaneAndCommitExpand`,
      `commits.go`.
- [x] Branch picker that loads history without changing the checkout.
      Evidence: `TestBranchPickerLoadsHistoryWithoutCheckout`, `branches.go`.
- [x] Fixed sample history built from the sample diff. Evidence:
      `TestSampleHistoryRail`, `sample_history.go`.
- [x] Draggable split between the working-tree files and History in the left
      drawer. Evidence: `TestHistoryLivesInSidebar`,
      `TestMouseDragHistoryDivider`.

### Phase 04: scrolling and chrome

- [x] Page scroll with fixed title bar, activity bar, explorer, tab strip,
      breadcrumbs, details, and status bar. Evidence: `TestShellMetrics`.
- [x] Windowed code and diff rows with a 20 px row. The slice starts at the
      first visible row so overscan boxes cannot sit over the chrome. Evidence:
      `TestScrollWindow`.
- [x] Optional details drawer that narrows the editor and rebuilds the slice.
      Evidence: `TestDetails`.
- [x] Single-key navigation for views, scrolling, tabs, hunks, and details.
      Evidence: `TestKeyboard`.
- [x] Soft-wrap long lines as fixed 20 px visual rows. Evidence:
      `TestWrapLongCodeLine`, `wrap.go`.

### Phase 05: search and command palette

- [x] `Ctrl+P` quick open with five default file suggestions and path
      filtering. Evidence: `TestQuickOpenPalette`, `palette.go`.
- [x] Open, arrow, Enter, Escape, and backdrop close paths. Evidence:
      `TestQuickOpenPalette`.
- [ ] Fuzzy ranking with a total match count.
- [x] `Ctrl+F` in-file find with a match count and next/previous. Evidence:
      `TestFindInFile`, `find.go`.
- [x] `Ctrl+Shift+F` workspace content search with a bounded list and
      jump-to-line. Evidence: `TestFeatureWorkspaceSearch`,
      `workspace_search.go`.
- [ ] Symbol search over the details outline.
- [ ] Filter history and a clear shortcut.
- [x] Modal focus trap for the palette, content search, and branch picker.
      Evidence: `TestFeatureModalFocusTrap`.
- [ ] Keyboard-only path for every action.
- Gate: a file opens from the palette with no pointer, and a find result
  scrolls to the matching row.

### Phase 06: live workspace

- [x] Watch the workspace for added, changed, and removed files. Evidence:
      `TestWatcherRescanAddsAndRemovesFiles`, `rescan.go`.
- [x] Follow `.git/HEAD`, refs, and the index instead of polling git status.
      Evidence: `TestWatcherSeesExternalBranchChange`.
- [x] Keep folder expansion and the active file across a rescan. Evidence:
      `TestWatcherRescanAddsAndRemovesFiles`, `installRescan`.
- [x] Window the explorer tree instead of the 26-row cap. Evidence:
      `TestTreeExpandedFolderScrolls`.
- [x] Recent files list that survives a restart. Evidence:
      `TestFeatureSessionRestoresLayout`, `session.go`.
- [x] Large-repository budget: stop scanning at one clear limit and say so.
      Evidence: `TestFeatureScanBudget`, `TestLoadDiskWorkspaceLimitTruncates`.
- Gate: a new file on disk appears in the tree without a restart, and the
  expanded folders stay open.

### Phase 07: editor depth

- [x] A second editor group with its own tab list. Evidence:
      `TestFeatureSplitEditor`, `editor_group.go`.
- [ ] Per-group scroll offsets. The two panes share the page offset.
- [ ] Horizontal scroll.
- [x] A word-wrap toggle for long lines. Evidence: `TestWrapLongCodeLine`.
- [x] Wrapped unified and split diff cells. Evidence:
      `TestFeatureDiffWrapIntegration`, `diff_wrap.go`.
- [x] Edit, selection, and save in the source view. Evidence:
      `TestFeatureEditAndSave`, `edit_buffer.go`.
- [ ] Expandable diff context between hunks.
- [ ] Go-to-line and a line-range copy.
- [ ] Breadcrumb dropdowns for directories and symbols.
- [ ] A minimap or scroll overview for files over 500 lines.
- Gate: two files are readable side by side, and a wrapped line keeps the
  row math exact.

### Phase 08: accessibility and mobile

- [x] Accessible names for the icon-only controls and the tree rows. Evidence:
      `TestAccessibilityLabels`.
- [x] Focus order that follows the visual order, with a visible ring and a
      modal trap. Evidence: `TestFeatureModalFocusTrap`, `base.css`.
- [ ] High-contrast tokens that keep the diff backgrounds distinct. The sign
      column already carries `+`/`-`.
- [x] Reduced motion for the drawer and the welcome card. Evidence:
      `base.css`.
- [ ] Live screen-reader bridging. `internal/atspi` maps the tree, but the
      production bus does not open D-Bus.
- [ ] Touch scrolling and a phone layout for the explorer and editor.
- [ ] Android build notes, insets, and the back key.
- Gate: the page is usable with no pointer, and a phone build reaches the
  file list and the diff.

### Phase 09: docs and gates

- [x] Add the example to `examples/maintenance.json` and `examples/readme.md`.
      Evidence: `examples/maintenance.json`, `examples/readme.md`.
- [x] Record the screenshot matrix for 1280x720, 1500x900, and 1920x1080.
      Evidence: `TestScreenshotMatrix`; fifteen PNGs under
      `/tmp/opencode/code-view-shots/`. The review set adds `edit`,
      `split-editor`, and `search`.
- [x] Document the diff fixture regeneration steps. Evidence:
      `examples/code-view/README.md`.
- [x] Confirm the docs catalog lists the example. Evidence:
      `examples/readme.md` feeds `frontend/src/content.js`; the Home
      showcase cards are curated and unchanged.
- [x] Run `make test`, `make build`, `make lint`, and the doc and example
      checks. Evidence: `make test`, `make build`, `make lint`.
- [x] Audit the public API use against `AGENTS.md` (no internal imports).
      Evidence: `make internal-import-check`, `make lint`.
- Gate: every gate is green and the example appears in the catalog.

## Risks

- The windowing math depends on one row height in CSS and in Go. A change to
  either side must change both, or the last row overlaps the status bar.
- JavaScript support has no bundled file. The tokenizer tests are the only
  visible proof until a repository is opened with `-root`.
- The explorer windows its visible rows; a very large repository still stops at
  the scan budget. The status bar and explorer badge say when that happened.
- The details drawer and the developer dock both take the right side. The dock
  covers the drawer while it is open, so the drawer is not a place for the only
  copy of information.
- The two editor groups share one page scroll. Phase 07 still owes independent
  offsets and a horizontal track.

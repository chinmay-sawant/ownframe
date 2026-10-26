# Code review

A dark code viewer built from one HTML template and Go data. The explorer lists a
workspace, a click opens a file in a tab, and Source Control shows a unified or
split diff. Every line is tokenized before it reaches the page, so the viewer
needs no browser, no script tag, and no canvas.

## Run

From the repository root:

```sh
go run ./examples/code-view
```

The viewer opens the current directory and reads the working-tree diff at
startup. The default window is 1500x900 and the minimum is 840x700. The
developer dock is 340 px on wide windows and shrinks below 1180 px, so the page
keeps usable width; F12 still closes it. The scan stops at 10,000 files by
default; `-files 25000` raises the budget and the status bar says when the limit
was hit. The bundled `testdata/sample` is the deterministic fixture the test
suite uses.

Point the viewer at another directory:

```sh
go run ./examples/code-view -root /path/to/repo
go run ./examples/code-view -root /path/to/monorepo -files 25000
```

Source Control runs `git diff HEAD` and lists untracked files. A background
fsnotify watcher rescans the tree, updates the diff, reopens changed documents,
and follows `.git/HEAD` and refs, so a new file appears without a restart and
an external commit or branch switch updates history. Expanded folders and open
tabs stay put; tabs for deleted files close. A clean tree says so; a directory
without a repository says that too. The viewer writes only when you edit a file
and press Save.

Other flags: `-dev=false` hides the developer dock, `-perf=false` stops pipeline
sampling, `-addr` changes the `-web` address, and `-web` serves the interactive
page (tick, keys, wheel, clicks, live frame updates).

## Screen

- The title bar holds the Code Review brand and the quick-open field.
- The activity bar switches Explorer and Source Control, opens quick open,
  opens Details, collapses the tree, and toggles the developer dock.
- The explorer is a foldered tree with a file filter, language badges, and
  change letters. Folders show a right or down chevron. The header has a
  collapse-all icon. When the tree is taller than the sidebar, use the mouse
  wheel over the tree, drag the thin scrollbar, use the footer arrows, or move
  focus with the keyboard. Hover a clipped tree row or tab to read its full
  path in the status bar.
- Tabs keep up to eight files and shrink their names so all eight stay
  visible. The Working tree tab has its own X; closing it returns to the open
  file. Closing the last file tab shows the welcome panel. The split button
  beside the tabs opens a second editor group with its own tab list; click a
  tab or the L button to move focus, and close the split from either strip.
- The editor has line numbers, token colours, Copy path, Copy file, a Wrap
  toggle, `Ctrl+F` find, `Ctrl+E` edit, a split editor, and a Markdown preview
  toggle. `Ctrl+S` saves an edited file. The active tab shows M, A, D, or U for
  a changed file and an unsaved dot in the status bar.
- Editing is a plain textarea over the editor body: selection, cut, copy,
  paste, undo, and redo use the same window chords as a form field. The sample
  workspace is read-only and Save reports that; a disk workspace writes the
  file and reloads its tokens.
- Details reads the active file: language, line count, exact size, encoding,
  indentation, change status, and an outline. In Source Control it follows the
  selected diff file. Click an outline row to jump to the line.
- The status bar shows the branch, changed count, the active language or the
  diff counts, the devtools state, the scan budget, rescan summaries, and
  unsaved files.

The developer dock opens on start, so F12 closes it instead of opening it. The
window lays the page out beside the dock, so the page controls stay clickable.
The layout fits a 1280x720 window and a 1920x1080 one.

Long lines soft-wrap by default in files and diffs. The Wrap chip in the
breadcrumb or `Z` turns that off and back on. Wrapped rows keep the 20 px grid,
and a continuation row leaves the line number column blank. A long unified diff
cell wraps inside its column; a split cell wraps independently and the taller
side sets the row count.

## Search

The top command bar opens quick open with `Ctrl+P`. It starts with five file
suggestions: the open tabs first, then the recent files, then the first files in
the workspace. Typing filters the paths in the open folder. Enter opens the
selected row, the arrows move it, and Escape or a click outside closes it.

The sidebar filter is separate. It keeps a path filter on the explorer while
the tree is open.

`Ctrl+F` opens find in the active file. It shows the match count, supports
case-sensitive matching, next and previous with wrap-around, and highlights
the visible matches. Escape closes the bar and restores the previous focus.

`Ctrl+Shift+F`, or the magnifier in the activity bar, opens workspace content
search. It reads every scanned text file, skips binary and oversized files,
caps one file at eight hits, and returns up to 200 matches. The list windows
eight rows around the selection; arrows move, Enter opens the file at the
source line, and Escape closes. Case matching is a toggle.

## Markdown preview

Press `Ctrl+Shift+V` or click Preview in the breadcrumb on a Markdown file. The
editor renders headings, paragraphs, lists, quotes, fenced code, tables, rules,
bold, italic, inline code, links, raw HTML text, and Mermaid flowcharts and
sequence diagrams as fixed-row text diagrams. `Source` returns to the token
view.

The preview reflows to the current editor width when the window or details
drawer changes. Table cells wrap inside aligned columns instead of printing
their pipes, and a long cell grows the row block. HTML comments are removed and
common tags (`b`, `i`, `code`, `a`, `br`, `img alt`, and block tags) render
their text instead of their source. Mermaid fences with `flowchart`, `graph`,
or `sequenceDiagram` render boxes and arrows; other diagram types show a note
and their source lines.

The preview uses the same 20 px rows as the code view, so the scroll and
windowing math is unchanged. It is a fixed-row approximation, not a browser:
HTML is reduced to text and Mermaid has no subgraph layout.

## Languages

The highlighter covers Go, Python, JavaScript and TypeScript, Markdown, CSS,
JSON, HTML, YAML, and shell. Everything else opens as plain text.

JavaScript support lives in the tokenizer and in `syntax_test.go`. The bundled
workspace carries no `.js` file on purpose, so open a repository with `-root` to
read one. The viewer does not run JavaScript and has no script tag.

## Source control

The viewer runs these against the workspace at startup and on every refresh:

```sh
git -C <root> diff --no-color --no-ext-diff HEAD
git -C <root> ls-files --others --exclude-standard -z
```

Untracked files become all-added hunks. Binary files get a notice instead of
rows. A repository without a first commit falls back to the working tree and
index diffs. Git gets a three-second budget. A directory without a repository
keeps the explorer and reports that git is unavailable.

The file list pages 13 changed files at a time. The arrows in the Source
Control summary switch pages. The diff has unified and split modes. Changed
words carry a stronger tint than the line background. Hunk arrows walk the
`@@` rows.

The Source Control drawer holds the working-tree files at the top and a
History pane below them. It reads up to 20 recent non-merge commits with one
`git log` call. Click a commit to expand its author line and changed files
inside the pane; clicking a file opens that patch in the editor. A patch is
read by a background worker and the pane says it is loading, so a click never
blocks the page. Hover a commit to read its message. Drag the divider between
the files and History to resize them; the pane auto-fits the commit count until
you drag it. `Working tree` returns the editor to the uncommitted diff.

In a live repository, click the branch name in the History header to list local
and remote branches and filter by name. The picker shows as many matches as the
pane height fits (at least four, up to 20) and says how many are hidden; type to
filter. Picking a branch loads its commits without changing the checkout, and
the result is cached for the session and across restarts. The status bar keeps
showing the checked-out branch. `Follow HEAD` returns the pane to the
checked-out branch.

The bundled sample uses a fixed three-commit history built from
`fixtures/sample.diff`. It is labeled `sample` and never runs git. A directory
without a repository hides the pane.

`fixtures/sample.diff` is the checked-in fixture for the sample workspace. It
touches seven files: modified, added, and deleted.

### Regenerating the sample diff

The fixture has to match the shipped `testdata/sample` tree. To rebuild it:

1. Copy `testdata/sample` to a temporary directory outside this repository.
2. Replace the changed files with their earlier content, remove the files that
   were added, and restore the files that were deleted.
3. Run `git init`, `git add -A`, and `git commit` in the temporary directory.
4. Copy the shipped `testdata/sample` over the temporary tree.
5. Add the new side and write the diff:

```sh
git -C /tmp/code-view-old add -A
git -C /tmp/code-view-old diff --cached --no-color --no-ext-diff HEAD > examples/code-view/fixtures/sample.diff
```

Keep the paths relative to the workspace root (`backend/probe.go`,
`web/theme.css`). The parser expects git headers, so a hand-written diff must
keep the `diff --git`, `---`, `+++`, and `@@` lines. The sample has no nested
`go.mod`: one would make `go:embed` prune the `backend` directory.

## Scrolling

The page scrolls. The title bar, activity bar, explorer, tab strip, breadcrumbs,
details drawer, and status bar stay fixed. `SetScrollWindow` slices the code or
diff rows around the viewport, and `Page.Scroll().SetWindowing(true)` turns the
scroll callback into a relayout. Every row is 20 px, which is what makes the
slice and spacer math exact.

The details drawer is 280 px. Opening it narrows the editor and its code rows,
so the windowing slice is rebuilt.

The viewer saves recents, both groups' tabs and active files, the split, the
history ratio, details, the dock state, wrap, preview, and diff mode under
`ownframe/code-view/session-<hash>.json` in the user config directory. The file
is written atomically with owner-only permissions and restored by `main` at
startup. Tests call `New` directly, so they never touch that file. `EnableSession`
is the only entry point that reads or writes it.

The history pane is a fixed paged region at the bottom of the Source Control
drawer. The divider changes how much of the drawer the file list and History
get. The diff uses the full editor body and the page scroll. The history pane
pages its commits, so it does not need a second scroll viewport.

## Keys

| Key | Action |
|---|---|
| `Ctrl+P` | Quick open |
| `Ctrl+F` | Find in the active file |
| `Ctrl+Shift+F` | Search workspace contents |
| `Ctrl+E` | Edit the active file |
| `Ctrl+S` | Save while editing |
| `Ctrl+\` | Split or close the second editor |
| `M` or `Ctrl+Shift+V` | Markdown preview |
| `Z` | Toggle soft wrap in files and diffs |
| `Enter` `Space` | Activate the focused control |
| `1` | Explorer |
| `2` | Source control |
| `3` | Details |
| `F` | Focus the sidebar file filter |
| `Up` `Down` | Scroll a row group |
| `Page Up` `Page Down` | Scroll a page |
| `Home` `End` | First or last row |
| `R` | Reload the open file, or refresh the diff |
| `W` | Close the active tab, or the Working tree tab in Source Control |
| `S` `U` | Split or unified diff |
| `N` `P` | Next or previous hunk |
| `Esc` | Close Details, the toast, search, or the edit buffer |

F12 and Ctrl+Shift+I belong to the window and toggle the developer dock. Tab
moves through the id-bearing controls; while the palette, content search, or
branch picker is open, background controls get `tabindex="-1"`, so Tab stays
inside the overlay. Ctrl+Shift+V now reaches the app; the window keeps plain
Ctrl+V for paste.

## Accessibility

Icon-only buttons and rows carry accessible names, toggles expose `aria-pressed`,
the palette, search panel, and branch picker are dialogs, and the toast and find
count are live regions. Focus rings use `:focus-visible`, and reduced motion
disables the toast and drawer transforms. Live screen-reader bridging is still
a library limitation: `internal/atspi` maps the tree but `NewLiveBus` does not
open D-Bus, so Orca and similar tools cannot attach yet.

## Files

| File | Purpose |
|---|---|
| `app.go` | State, view assembly, windowing, tabs, modes |
| `pane.go` | Two-group pane geometry and the second editor's rows |
| `split_ops.go` | Split toggle and focus routing |
| `edit_ops.go` | Edit buffer, save, and dirty tracking |
| `search_ops.go` | Workspace content search panel |
| `session.go` | Session file load, save, sanitize |
| `app_session.go` | Restore and persist layout |
| `actions.go` | Click, filter, search, and keyboard routing |
| `palette.go` | Quick-open rows, selection, and focus |
| `markdown_preview.go` | Rendered Markdown rows and source anchors |
| `markdown_html.go` | HTML text extraction and entities |
| `markdown_mermaid.go` | Flowchart and sequence diagrams |
| `syntax.go` | Shared lexer and the Go, Python, and JavaScript specs |
| `syntax_markup.go` | CSS, JSON, HTML, YAML, and shell |
| `syntax_markdown.go` | Markdown blocks and inline spans |
| `diff.go` | Unified diff parser, word emphasis, row builders |
| `diff_wrap.go` | Unified and split diff wrapping |
| `workspace.go` | Sample and disk scanning, file reads, limits |
| `workspace_write.go` | Save one file under the root |
| `workspace_search.go` | Content-search engine |
| `edit_buffer.go` | Text editing model |
| `editor_group.go` | One group's tab list |
| `rescan.go` | Workspace delta and git metadata detection |
| `outline.go` | The details panel symbols |
| `git.go` | Live working-tree changes |
| `watch.go` | Recursive fsnotify watcher, rescan, git refresh |
| `branches.go` | Local and remote branch list for the history picker |
| `branch_filter.go` | Branch filter and rank |
| `branch_cache.go` | Session branch-log cache |
| `commits.go` | Commit metadata and one-file patches |
| `commit_loader.go` | Background patch worker |
| `sample_history.go` | Fixed sample commits built from the fixture diff |
| `find.go` | In-file find, match rows, and highlights |
| `wrap.go` | Soft-wrap plan and source-line anchors |
| `components/` | HTML fragments and the stylesheets |

## Limits

- ownframe has no second native window. The two editor groups live in one
  window and share one page scroll, so a split keeps both panes on the same
  row offset.
- Files over 512 KB and files with a NUL byte show a notice. A file over 20,000
  lines keeps the first 20,000.
- The tree windows its visible rows. The wheel scrolls it under the pointer;
  the scrollbar, footer arrows, and keyboard also work.
- The scan skips `.git`, `node_modules`, `vendor`, build output, and binary
  extensions. It stops at 10,000 files unless `-files` raises the budget; the
  explorer badge, status bar, and quick open all say when the limit was hit.
- There is no horizontal scroll. Files and diff cells soft-wrap; `Z` turns it
  off and the diff clips with an ellipsis.
- The History pane keeps 20 commits and 20 files per commit. Patches load on a
  background worker with a three-second git budget. The branch picker shows as
  many matches as the pane fits and reports how many more need the filter.
- This example targets desktop. There is no phone layout, Android host, or
  mobile accessibility path.
- The fsnotify watcher is desktop-only. `-web` keeps the manual Refresh and `R`
  path. The watcher follows file adds, removes, and changes plus `.git/HEAD`
  and refs, so an external commit or branch switch updates without Refresh.
- `-web` serves the same tick, key, wheel, click, and live-frame loop as the
  browser shell; it does not start the fsnotify watcher.
- Quick open searches file names; `Ctrl+Shift+F` searches file contents.
- The Markdown preview uses fixed rows. Tables wrap and raw HTML is reduced to
  text; Mermaid supports basic flowcharts and sequence diagrams, not subgraphs
  or browser HTML.
- The sample fixture is synthetic test data; the default workspace is the
  current directory.
- Editing is a plain textarea, not an IDE: no syntax highlighting inside the
  textarea, no multi-cursor, and no unsaved-change prompt when a tab closes.
  The sample workspace is read-only. A very long file edits in a fixed-height
  field, so use the source view for reading long files.

## Tests

```sh
go test -p 1 ./examples/code-view
```

The suite covers language detection, token text preservation, Go, Python,
JavaScript, and Markdown tokens, the diff parser, hunk numbers, word emphasis,
split pairing, unified and split diff wrapping, untracked files, workspace
scanning and scan limits, content search, session round trips, shell metrics,
tabs, modes, two editor groups, editing and saving, filtering, pointer hits on
the activity/tree/tab/diff/status controls, quick open defaults and filtering,
Markdown preview rows, tables, HTML, Mermaid, reflow, and the chord, outline
scroll jumps, details metadata, the diff pager, async commit patches, live and
sample commit history, branch filtering without checkout, commit expansion, the
sidebar history divider, diff-cell clipping, soft-wrap rows, in-file find,
explorer windowing, wheel and keyboard navigation, background file watching and
rescan, external branch changes, scrolling after a long file opens, devtools,
the clipboard, accessibility names and the modal focus trap, and keyboard
routing and activation.

Capture the review frames:

```sh
OWNFRAME_CODE_VIEW_SCREENSHOT=1 go test -run '^TestScreenshot$' ./examples/code-view
```

PNGs land in `/tmp/opencode/code-view-shots/`: `explorer`, `preview`, `palette`,
`changes`, `split`, `details`, `edit`, `split-editor`, `search`, `scrolled`, and
`welcome`. Nothing is added to the repository.

The size matrix runs five views at 1280x720, 1500x900, and 1920x1080:

```sh
OWNFRAME_CODE_VIEW_SCREENSHOT=1 go test -run '^TestScreenshotMatrix$' ./examples/code-view
```

It writes `matrix-<width>x<height>-<view>.png` for the explorer, details,
scrolled, changes, and split views.

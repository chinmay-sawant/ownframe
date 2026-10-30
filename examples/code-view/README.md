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
keeps usable width; F12 opens it. The scan stops at 10,000 files by
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

Other flags: `-dev` opens the developer dock (hidden by default), `-perf=false` stops pipeline
sampling, `-chrome=false` keeps the native title bar instead of the drawn minimize, maximize,
and close controls, `-addr` changes the `-web` address, and `-web` serves the interactive
page (tick, keys, wheel, clicks, live frame updates).

On Windows, opening a workspace under a WSL network path (`\\wsl$\...`) hands the viewer off
into WSL: the Windows build re-runs its Linux companion there via `wsl.exe` with the path
translated, so scanning, git, and watching run on Linux. `-wsl=false` forces the native
Windows build, and `-wsl-bin` overrides the companion location (by default it sits next to
the `.exe`). The Open folder browser lists subdirectories; Up at a drive root steps out to
This PC with every available drive, the list scrolls with the mouse wheel, and a long path
clips with an ellipsis while the full path stays in the tooltip.

## Screen

- The title bar holds the Code Review brand and the quick-open field.
- The activity bar switches Explorer and Source Control, opens quick open,
  opens Details and Settings, and toggles the developer dock. Settings offers six
  color themes, saved with the workspace session.
- The explorer is a foldered tree with a file filter, language badges, and
  change letters. Folders show a right or down chevron. The header has a
  collapse-all button. When the tree is taller than the sidebar, use the mouse
  wheel over the tree, drag the thin scrollbar, use the footer arrows, or move
  focus with the keyboard. Hover a clipped tree row or tab to read its full
  path in the status bar. Open folders pin as sticky headers above the scroll
  window; `Space` toggles the focused folder or opens the focused file.
  Right-click a row for Copy Path, Copy File, and Paste.
- Tabs keep up to eight files and shrink their names so all eight stay
  visible. Right-click a tab or its close button for Close, Close Others, and
  Close All; Close Others needs at least two tabs. The Working tree tab has
  its own X; closing it returns to the open file. Closing the last file tab
  shows the welcome panel. The split button beside the tabs opens a second
  editor group with its own tab list; click a tab or the L button to move
  focus, and close the split from either strip.
- The editor has line numbers, token colours, Copy path, Copy file, a Wrap
  toggle, `Ctrl+F` find, click-to-edit, a split editor, and a Markdown preview
  toggle. Click anywhere in the source (or press `Ctrl+E`) to start editing;
  there is no separate Edit button. `Ctrl+S` saves. The active tab shows M, A,
  D, or U for a changed file and an unsaved dot in the status bar.
- A changed file shows a gutter mark on each added, modified, or deleted line
  (`+`, `~`, or `-`) and a faint line tint. Click the mark to open a line card
  with the old and new text. The card has up and down arrows that stop at the
  first and last change in the same file.
- A translucent changed-file bar sits at the bottom center with previous and
  next arrows, the current position, and the file's add/delete counts. It has
  no backdrop, so the editor behind it stays visible. Click the file name to
  open it, or use the arrows to move through the working tree.
- Editing keeps the highlighted code view. Click a line to start editing on
  it; after that, a click places the caret at the glyph, and a drag selects.
  The token colours, line numbers, gutter marks, caret, and selection stay in
  place. `Ctrl+A` selects all, and `Ctrl+C` copies. Arrow keys, Home/End, Page
  Up/Down, Enter, Backspace, Delete, word delete, cut, paste, undo, and redo
  work on the buffer. The sample workspace is read-only and Save reports
  that; a disk workspace writes the file and refreshes its tokens.
- Files and folders listed in `.gitignore` stay in the explorer as solid,
  muted gray rows with an `IGNORED` badge. A skipped ignored directory such as
  `node_modules` still gets one row so it is visible without scanning its
  tree.
- Details reads the active file: language, line count, exact size, encoding,
  indentation, change status, and an outline. In Source Control it follows the
  selected diff file. Click an outline row to jump to the line.
- The status bar shows the branch, change summary, the active language or the
  diff counts, the devtools state, the scan budget, rescan summaries, and
  unsaved files.

The developer dock stays hidden on start, so F12 opens it. The
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
bold, italic, inline code, links, structured HTML, and Mermaid flowcharts and
sequence diagrams as fixed-row text. `Source` returns to the token view.

The preview reflows to the current editor width when the window or details
drawer changes. Table cells wrap inside aligned columns instead of printing
their pipes, wrapped rows get a gap row, and every visual row is exactly 20 px
with its own line height. HTML block tags become structured rows: headings,
paragraphs, lists, quotes, details/summary, images with alt text, and tables
with rows or cells. Inline tags keep their emphasis. `script` and `style`
contents are removed. Mermaid fences with `flowchart`, `graph`, or
`sequenceDiagram` render boxes and arrows; other diagram types show a note and
their source lines.

The preview uses the same 20 px rows as the code view, so the scroll and
windowing math is unchanged. It is a fixed-row approximation, not a browser:
HTML becomes preview rows rather than live DOM, and Mermaid has no subgraph
layout.

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

The changed-file list is a vertical scroll with a draggable rail and the mouse
wheel; it has no pager buttons. The diff has unified and split modes. Changed
words carry a stronger tint than the line background. The up and down arrows
step through contiguous change blocks and highlight the selected block.

A header such as `@@ -10,6 +10,8 @@` describes one Git hunk: six lines starting
at line 10 in the old file and eight lines starting at line 10 in the new file.
Counts include unchanged context lines as well as deletions on the old side
and additions on the new side. A count of one omits `,1`; a count of zero marks
an insertion or deletion boundary. On an empty side, the start number is the
line before that boundary; zero means the beginning of the file. Unified and
split modes show the same ranges, regardless of soft wrapping or which rows
are visible.

One hunk can contain several change blocks separated by unchanged lines.
The revert arrow restores only its block. In unified mode it appears on the
first added row, or the first deleted row for a deletion-only block. In split
mode it sits beside the center divider. Reverting refreshes the open file and
keeps the diff view open; Ctrl+Z restores the reverted block.

The Source Control drawer holds the working-tree files at the top and a
History pane below them. The History rail reads 20 recent commits at a time,
draws a VS Code-style lane graph, and scrolls vertically with a draggable rail;
it has no pager buttons. The graph keeps a first parent in its lane, opens a
lane for an extra parent, and joins a lane another commit already owns, up to
six lanes. When the scroll passes half of the loaded rows it asks for a larger
page on a background worker, growing 20, 40, 60, 90, up to 500. Click a commit
to load its first patch in the editor. A patch is read by a background worker
and the pane says it is loading, so a click never blocks the page. Hover a
commit to read its message in a floating card to the right of the history
rail; the card clamps to the window and shows the subject, body, author, and
date. Drag the divider between the files and History to resize them; the pane
auto-fits the commit count until you drag it. `Working tree` returns the editor
to the uncommitted diff.

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

The history pane is a fixed region at the bottom of the Source Control
drawer. The divider changes how much of the drawer the file list and History
get. The diff uses the full editor body and the page scroll. The history pane
has its own vertical scroll and loads a larger commit page when the user passes
half of the rows already loaded.

## Keys

| Key | Action |
|---|---|
| `Ctrl+P` | Quick open |
| `Ctrl+F` | Find in the active file |
| `Ctrl+Shift+F` | Search workspace contents |
| `Ctrl+E` | Edit the active file (or click inside it) |
| `Ctrl+S` | Save while editing |
| `Ctrl+A` | Select all in the edit buffer |
| `Ctrl+C` | Copy the edit selection |
| `Ctrl+V` | Paste a staged copy or clipboard file into the explorer folder |
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

Paste in the explorer with `Ctrl+V` or the Paste menu row. A staged copy
lands in the focused folder, or beside the open file, with ` (copy)`
suffixes on collision; a `file://` URI list on the clipboard copies the
first existing file, and plain text becomes a new `paste.txt` (non-UTF-8
bytes become `paste.bin`). PNG, JPEG, and PDF bytes copy the same way. A
`C:\…` clipboard path is translated to `/mnt/<drive>/…` on WSL. The sample
workspace is read-only, and `-web` exposes no clipboard route, so paste
there runs from the menu.

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
| `context_menu.go` | Tab and explorer context-menu rows, close, copy, and paste actions |
| `context_menu_bridge.go` | Public menu row conversion |
| `palette.go` | Quick-open rows, selection, and focus |
| `markdown_preview.go` | Rendered Markdown rows and source anchors |
| `markdown_html.go` | Structured HTML preview blocks and entities |
| `markdown_mermaid.go` | Flowchart and sequence diagrams |
| `syntax.go` | Shared lexer and the Go, Python, and JavaScript specs |
| `syntax_markup.go` | CSS, JSON, HTML, YAML, and shell |
| `syntax_markdown.go` | Markdown blocks and inline spans |
| `diff.go` | Unified diff parser, word emphasis, row builders |
| `diff_wrap.go` | Unified and split diff wrapping |
| `workspace.go` | Sample and disk scanning, file reads, limits |
| `explorer_clipboard.go` | Explorer copy path/file, staged paste, and the shared pending-copy state |
| `explorer_paste_wire.go` | Ctrl+V chord and menu paste into the explorer folder |
| `workspace_write.go` | Save one file under the root |
| `workspace_search.go` | Content-search engine |
| `edit_buffer.go` | Text editing model |
| `selection_model.go` | Selection ranges, word and line marks |
| `selection_ops.go` | Click and drag selection in the edit body |
| `text_edit.go` | Caret keys, undo, clipboard, and edit snapshots |
| `editor_group.go` | One group's tab list |
| `rescan.go` | Workspace delta and git metadata detection |
| `outline.go` | The details panel symbols |
| `git.go` | Live working-tree changes |
| `changes_scroll.go` | Changed-file list scroll model and thumb |
| `watch.go` | Recursive fsnotify watcher, rescan, git refresh |
| `branches.go` | Local and remote branch list for the history picker |
| `branch_filter.go` | Branch filter and rank |
| `branch_cache.go` | Session branch-log cache |
| `commits.go` | Commit metadata and one-file patches |
| `commit_loader.go` | Background patch worker |
| `commit_graph.go` | Commit lane graph layout |
| `history_scroll.go` | History rail scroll and page-growth math |
| `history_load.go` | Background history page worker |
| `sample_history.go` | Fixed sample commits built from the fixture diff |
| `find.go` | In-file find, match rows, and highlights |
| `wrap.go` | Soft-wrap plan and source-line anchors |
| `components/` | HTML fragments and the stylesheets |
| `file_nav_ops.go` | Bottom-center changed-file navigator |

## Limits

- ownframe has no second native window. The two editor groups live in one
  window and share one page scroll, so a split keeps both panes on the same
  row offset.
- Files over 512 KB and files with a NUL byte show a notice. A file over 20,000
  lines keeps the first 20,000. A file with a line over 32 KB shows a notice
  instead of opening.
- A diff over 512 KB, 20,000 lines, or one line over 32 KB renders one notice
  row instead of the changed lines, so a deleted or regenerated bundle cannot
  stall the pane. A commit patch that hits its 512 KB read cap shows the same
  notice. The wrapped rows are cached per file and column budget, and scrolling
  reuses them.
- The tree windows its visible rows. The wheel scrolls it under the pointer;
  the scrollbar, footer arrows, and keyboard also work.
- The scan skips `.git`, `node_modules`, `vendor`, build output, and binary
  extensions. It stops at 10,000 files unless `-files` raises the budget; the
  explorer badge, status bar, and quick open all say when the limit was hit.
- There is no horizontal scroll. Files and diff cells soft-wrap; `Z` turns it
  off and the diff clips with an ellipsis.
- The History pane loads 20 commits, then grows the page as the rail crosses
  half of the loaded rows, up to 500. Patches load on a background worker with a
  three-second git budget. The branch picker shows as many matches as the pane
  fits and reports how many more need the filter.
- This example targets desktop. There is no phone layout, Android host, or
  mobile accessibility path.
- The fsnotify watcher is desktop-only. `-web` keeps the manual Refresh and `R`
  path. The watcher follows file adds, removes, and changes plus `.git/HEAD`
  and refs, so an external commit or branch switch updates without Refresh.
- `-web` serves the same tick, key, wheel, click, and live-frame loop as the
  browser shell; it does not start the fsnotify watcher.
- Quick open searches file names; `Ctrl+Shift+F` searches file contents.
- The Markdown preview uses fixed rows. Tables wrap, raw HTML becomes preview
  rows, and Mermaid supports basic flowcharts and sequence diagrams, not
  subgraphs or a live browser DOM.
- Gutter marks and their line cards cover working-tree changes only. A commit
  patch keeps its unified/split diff view.
- The sample fixture is synthetic test data; the default workspace is the
  current directory.
- Editing keeps the highlighted view, not a plain textarea: syntax tokens,
  line numbers, selection, and a blinking caret stay visible. It is still a
  simple buffer: no multi-cursor, no language server, and no unsaved-change
  prompt when a tab closes. The sample workspace is read-only.
- Git-ignored paths stay in the explorer as solid, muted gray rows with an
  `IGNORED` badge. The viewer asks git for the ignored list, so a non-git
  directory has no ignored styling; a skipped ignored directory gets one empty
  placeholder row.

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
Markdown preview rows, tables, HTML, Mermaid, reflow, and the chord, click-to-edit
with preserved syntax spans and caret typing, drag selection, select all and
copy, inline change cards with prev/next arrows, gutter markers, the
bottom-center navigator, tab context menus, the floating commit tooltip,
git-ignored tree rows, outline scroll jumps, details metadata, the changed-file
scroll rail, commit graph lanes, async history pages, async commit patches,
live and sample commit history, branch filtering without checkout, commit
expansion, the sidebar history divider, diff-cell clipping, soft-wrap rows,
in-file find, explorer windowing, wheel and keyboard navigation, background
file watching and rescan, external branch changes, scrolling after a long file
opens, devtools, the clipboard, explorer copy and paste with Ctrl+V, accessibility names and the modal focus trap,
and keyboard routing and activation.

Capture the review frames:

```sh
OWNFRAME_CODE_VIEW_SCREENSHOT=1 go test -run '^TestScreenshot$' ./examples/code-view
```

PNGs land in `/tmp/opencode/code-view-shots/`: `explorer`, `preview`, `palette`,
`changes`, `split`, `details`, `edit`, `inline-card`, `split-editor`,
`commit-tip`, `search`, `scrolled`, `scrollrails`, `ignored`, and `welcome`.
Nothing is added to the repository.

The size matrix runs five views at 1280x720, 1500x900, and 1920x1080:

```sh
OWNFRAME_CODE_VIEW_SCREENSHOT=1 go test -run '^TestScreenshotMatrix$' ./examples/code-view
```

It writes `matrix-<width>x<height>-<view>.png` for the explorer, details,
scrolled, changes, and split views.

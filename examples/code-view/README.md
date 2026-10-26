# Beacon: code view

A dark code viewer built from one HTML template and Go data. The explorer lists a
workspace, a click opens a file in a tab, and Source Control shows a unified or
split diff. Every line is tokenized before it reaches the page, so the viewer
needs no browser, no script tag, and no canvas.

## Run

From the repository root:

```sh
go run ./examples/code-view
```

The bundled workspace is a small uptime monitor. Source Control shows a
checked-in sample diff so the view has something to read.

Point the viewer at a real directory:

```sh
go run ./examples/code-view -root /path/to/repo
```

With `-root`, Source Control runs `git diff HEAD` and lists untracked files. A
clean tree says so. The viewer never writes to the repository.

Other flags: `-dev=false` hides the developer dock, `-perf=false` stops pipeline
sampling, `-addr` changes the `-web` address, and `-web` serves the first frame.

## Screen

- The title bar holds the workspace, the active path, the branch, the details
  toggle, and the devtools toggle.
- The activity bar switches Explorer and Source Control, focuses the filter, and
  opens Details.
- The explorer is a foldered tree with a file filter, language badges, and
  change letters.
- Tabs keep up to eight files. Closing the last tab shows the welcome panel.
- The editor has line numbers, token colours, Copy path, and Copy file.
- Details shows language, line count, size, status, and an outline. Click an
  outline row to jump to it.
- The status bar shows the branch, changed count, the active language or the
  diff counts, and the devtools state.

The developer dock opens on start, so F12 closes it instead of opening it. The
layout fits a 1280x720 window and a 1920x1080 one, but long lines clip sooner on
the small size.

## Languages

The highlighter covers Go, Python, JavaScript and TypeScript, Markdown, CSS,
JSON, HTML, YAML, and shell. Everything else opens as plain text.

JavaScript support lives in the tokenizer and in `syntax_test.go`. The bundled
workspace carries no `.js` file on purpose, so open a repository with `-root` to
read one. The viewer does not run JavaScript and has no script tag.

## Source control

The bundled diff is a fixture in `fixtures/sample.diff`. It touches seven files:
modified, added, and deleted. With `-root`, the viewer runs:

```sh
git -C <root> diff --no-color --no-ext-diff HEAD
git -C <root> ls-files --others --exclude-standard -z
```

Untracked files become all-added hunks. Binary files get a notice instead of
rows. A repository without a first commit falls back to the working tree and
index diffs. Git gets a three-second budget.

The diff has unified and split modes. Changed words carry a stronger tint than
the line background. Hunk arrows walk the `@@` rows.

### Regenerating the sample diff

The fixture has to match the shipped `testdata/sample` tree. To rebuild it:

1. Copy `testdata/sample` to a temporary directory outside this repository.
2. Replace the changed files with their earlier content, remove the files that
   were added, and restore the files that were deleted.
3. Run `git init`, `git add -A`, and `git commit` in the temporary directory.
4. Copy the shipped `testdata/sample` over the temporary tree.
5. Add the new side and write the diff:

```sh
git -C /tmp/beacon-old add -A
git -C /tmp/beacon-old diff --cached --no-color --no-ext-diff HEAD > examples/code-view/fixtures/sample.diff
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

## Keys

| Key | Action |
|---|---|
| `1` | Explorer |
| `2` | Source control |
| `3` | Details |
| `F` | Focus the file filter |
| `Up` `Down` | Scroll a row group |
| `Page Up` `Page Down` | Scroll a page |
| `Home` `End` | First or last row |
| `R` | Reload the open file, or refresh the diff |
| `W` | Close the active tab |
| `S` `U` | Split or unified diff |
| `N` `P` | Next or previous hunk |
| `Esc` | Close Details or the toast |

F12 and Ctrl+Shift+I belong to the window and toggle the developer dock. Tab
focuses the filter input when the explorer is open.

## Files

| File | Purpose |
|---|---|
| `app.go` | State, view assembly, windowing, tabs, modes |
| `actions.go` | Click, filter, and keyboard routing |
| `syntax.go` | Shared lexer and the Go, Python, and JavaScript specs |
| `syntax_markup.go` | CSS, JSON, HTML, YAML, and shell |
| `syntax_markdown.go` | Markdown blocks and inline spans |
| `diff.go` | Unified diff parser, word emphasis, row builders |
| `workspace.go` | Sample and disk scanning, file reads, limits |
| `outline.go` | The details panel symbols |
| `git.go` | Live `-root` changes |
| `components/` | HTML fragments and the stylesheets |

## Limits

- ownframe has no second native window. The panes in this example live in one
  window.
- Files over 512 KB and files with a NUL byte show a notice. A file over 20,000
  lines keeps the first 20,000.
- The tree shows 26 rows. Use the filter to reach the rest of a large
  repository.
- The scan skips `.git`, `node_modules`, `vendor`, build output, and binary
  extensions.
- Long lines clip. There is no horizontal scroll or word wrap yet.
- The developer dock covers the details drawer when both are open.
- `-web` serves one frame because `Serve` does not tick. The desktop window is
  the interactive path.
- The sample diff is fixed data. Only `-root` reads the working tree.
- The viewer is read-only. Editing, saving, and a second editor group are in
  the plan.

## Tests

```sh
go test -p 1 ./examples/code-view
```

The suite covers language detection, token text preservation, Go, Python,
JavaScript, and Markdown tokens, the diff parser, hunk numbers, word emphasis,
split pairing, untracked files, workspace scanning, shell metrics, tabs, modes,
filtering, scrolling, details, devtools, the clipboard, and keyboard routing.

Capture the review frames:

```sh
OWNFRAME_CODE_VIEW_SCREENSHOT=1 go test -run '^TestScreenshot$' ./examples/code-view
```

PNGs land in `/tmp/opencode/code-view-shots/`: `explorer`, `changes`, `split`,
`details`, `scrolled`, and `welcome`. Nothing is added to the repository.

The size matrix runs five views at 1280x720, 1500x900, and 1920x1080:

```sh
OWNFRAME_CODE_VIEW_SCREENSHOT=1 go test -run '^TestScreenshotMatrix$' ./examples/code-view
```

It writes `matrix-<width>x<height>-<view>.png` for the explorer, details,
scrolled, changes, and split views.

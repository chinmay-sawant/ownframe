# Phase 01: shell and workspace

Goal: a fixed shell around a scanned workspace, with a tree, tabs, a welcome
panel, details, and a status bar. This phase landed with the first example.

## Where it lives

- `examples/code-view/app.go` holds the state, the view assembly, tabs, and
  the two modes.
- `components/titlebar.html`, `activity.html`, `sidebar.html`, `editor.html`,
  `details.html`, and `statusbar.html` draw the panes.
- `components/base.css` and `chrome.css` carry the shell metrics.
- `workspace.go` scans the embedded sample or a `-root` directory.

## Design

1. The shell is one fixed layout. The title bar is 42 px, the status bar 24 px,
   the activity bar 46 px, and the explorer 268 px. The editor starts at
   x=314 and y=42.
2. Every fixed pane gets an explicit height. The engine does not stretch a
   fixed box from `top` and `bottom` alone; a zero height collapses the code
   body.
3. The explorer shows 26 rows. Folders start closed except `backend` and
   `web`. A non-empty filter flattens the list to matching paths.
4. Change letters come from the loaded diff: M, A, D, U, or R.
5. Tabs keep at most eight paths. Opening a ninth drops the oldest tab, not
   the active one. Closing the last tab shows the welcome panel.
6. The details drawer is 280 px and starts closed. Opening it narrows the
   editor through `.with-details`.
7. The welcome panel lists four quick-open files and the key map.
8. Developer tools start on through `Config.DevTools` and `-dev=true`.

## Checklist

- [x] Parse the template and draw a fixed 1500x900 shell.
- [x] Enforce a 1180x700 minimum.
- [x] Scan the embedded sample with `fs.WalkDir`.
- [x] Scan a real directory with `filepath.WalkDir` and skip `.git`,
      `node_modules`, `vendor`, build output, and binary extensions.
- [x] Sort folders before files and names case-insensitively.
- [x] Fold and unfold a directory.
- [x] Filter by path and show the hidden-row count.
- [x] Open a file in a tab, switch tabs, close tabs, and show the welcome.
- [x] Show language, line count, size, status, and an outline in Details.
- [x] Add a status bar with the branch, changed count, and devtools chip.
- [x] Start the developer dock on.
- [x] Add the sample project under `testdata/sample`.

## Gate

`make test` passes, and `TestShellMetrics` confirms every pane box at
1500x900. `TestResponsiveLayout` confirms the status bar and editor at
1280x720 and 1920x1080.

## Risks

- A fixed pane that loses its explicit height collapses the code body. The
  metric test is the guard.
- A folder with its own `go.mod` is pruned from `go:embed`. The sample moved
  under `testdata` and dropped the nested module file so the Go files embed.
- The 26-row cap is a limit, not a design. Phase 06 replaces it with a windowed
  tree.

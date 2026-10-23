# Hot reload

Recorded 2026-10-03 against `5a428f4`.

Every example embeds its HTML and hands `New` one string
(`examples/login/login/login.go:12`, `examples/flappy-bird/flappy/html.go:8`).
Change the file and the window shows the old bytes until the process stops and
`go run` builds again. This plan gives a page a file to read, watches that
file by default, and redraws the window when it changes. The switch is a
config field first and a `-reload` flag in the examples, default on. A page
built from an HTML string, which is every test and every embedded build,
keeps the behavior it has today.

## What happens today

1. `New` takes `Config.HTML`, a string (`internal/page/page_new.go:9`), and
   parses it once (`internal/page/page_new.go:38`). No path rereads a file
   after that, and `Config` has no path field (`internal/page/page.go:39`).
2. `Load` reparses the HTML and pushes it on the history stack
   (`internal/page/page_nav.go:19`). `Back` and `Forward` redraw from that
   stack (`internal/page/page_nav.go:41`). There is no reload that replaces
   the current entry without touching history.
3. The window loop has one per-frame hook, `tickFrame`
   (`internal/window/game.go:95`), and it belongs to the page's `SetTick`. A
   watch cannot use it: a page may already have a tick, and `SetTick` is a
   single slot the caller owns.
4. `Serve` has no loop at all (`documentation/web.md:59`), so it has no frame
   to poll on.
5. The loop a developer runs is edit, Ctrl-C, `go run` again. With
   `//go:embed`, even that step does not see the edit until the build runs.
6. `Page.SetTheme` (`internal/page/theme.go:13`) restyles a running page, but
   the theme source arrives in the same string as everything else. A CSS file
   has the same restart problem.

## Target behaviour

| Case | Today | After |
|---|---|---|
| Built from `Config.File` | no such field | The file is read at `New` and watched. |
| Edit the template | rebuild and restart | The open window redraws within one poll interval. |
| Edit `Config.ThemeFile` | no such field | The theme is reread and reapplied on the same poll. |
| Bad edit | n/a | The last good picture stays. The error is kept and printed once to stderr; the window does not die. |
| History | n/a | The reload replaces the current history entry. `HTML()` returns the new source and `Back` still walks older entries. |
| Form and pointer state | n/a | Focus, hover, active, and scroll survive when their ids and geometry still exist; a missing id is cleared. |
| Opt out | n/a | `Config.DisableHotReload: true`, or `-reload=false` in the example flag. |
| `Serve` | no loop | The next request reloads first, and the shell page refreshes the frame on an interval while a watch is active. |
| wasm and mobile | n/a | No watch. Keep `Config.HTML` there; `Config.File` is a desktop and server source. |

## Shape

Add `Config.File` and `Config.ThemeFile`, both paths, both optional.
`Config.HTML` stays for callers that assemble the string themselves. A config
that sets both `HTML` and `File` is an error rather than a precedence rule, so
there is one answer to which source is live. When `File` is set, the page
reads it before parsing. Watching is on by default and off with
`Config.DisableHotReload: true` or `Page.SetHotReload(false)`.

The watch is a poll, not an OS notification. The ground rules forbid new
modules, and two `os.Stat` calls a second on one file cost nothing. The page
keeps the path, the last modification time, the size, and a hash of the bytes
it last accepted. A poll that sees a different mtime or size reads the file,
hashes it, and reparses only when the hash differs. A parse error keeps the
last good template and stores the bytes as pending, so the next poll retries a
half-written file until it parses. A syntax error is fixed by editing the file
again, with no restart.

## Phase 1: read the source from disk

- [x] Add `File`, `ThemeFile`, and `DisableHotReload` to `Config`
      (`internal/page/page.go:39`).
- [x] Add `ErrBadSource` for a config with both `HTML` and `File` set, and for
      a `File` that cannot be read at `New`. `ErrEmptyHTML` keeps its meaning
      for a blank string and a blank file.
- [x] Read `File` in `New` before the template parse
      (`internal/page/page_new.go:38`), and `ThemeFile` before `parseTheme`
      (`internal/page/page_new.go:43`).
- [x] Point `past` at the file contents (`internal/page/page_new.go:61`), so
      `HTML()` returns what the window shows, not the path.
- [x] Store a `watch` value on the page: the two paths, their last
      `FileInfo`, and the bytes each last accepted. New file
      `internal/page/page_watch.go`, under 2000 characters.
- [x] Keep every existing call site working. `Config` grows two string fields
      and a bool, and `New` keeps its signature.

Exit: `New` on a temp file renders the same picture as `New` on the string it
contains, and `HTML()` returns the file bytes.

## Phase 2: the poll and the swap

- [x] Add `page.PollReload(ctx) (bool, error)`. It stats each watched path,
      rereads a changed one, and returns whether the page changed. New file
      `internal/page/page_reload.go`.
- [x] On a changed file, parse the new template in memory. On success, swap
      `p.tpl`, replace `p.past[p.pastAt]` with the new bytes, clear the
      pending state, and redraw. History length and index do not move.
- [x] On a parse error, keep `p.tpl` and `p.past[p.pastAt]`, store the bytes
      and the error in the pending state, and return the error. The next poll
      retries the same bytes until they parse or change again.
- [x] A `ThemeFile` change calls the same path as `SetTheme`
      (`internal/page/theme.go:13`). A broken stylesheet keeps the old theme
      and stores the error in the same pending slot.
- [x] Invalidate whatever `Redraw` caches, once it caches anything
      (dynamic-resize Phase 1). A reload is a `SetData`-class change: the
      executed bytes and the parse cache both go.
- [x] Clamp the scroll offset after a reload (dynamic-resize Phase 6). A
      shorter page must not leave the view below its own bottom.
- [x] Resolve focus, hover, and active ids against the new boxes after the
      redraw, and clear the ones that no longer exist. The resolution code is
      the same one dynamic-resize Phase 4 adds for a relayout.

Exit: a test writes a temp file, calls `PollReload`, and finds one new
generation, the new text in `HTML()`, and the same history length.

## Phase 3: the window watch

- [x] Add an optional `host.Reloader` interface with
      `PollReload(ctx context.Context) (bool, error)` next to `Ticker`
      (`internal/host/screen.go:14`).
- [x] In `shell.Update` (`internal/window/game.go:61`), after `tickFrame` and
      before `syncImage`, call the reloader when the screen has one and the
      poll interval, 250 ms, has passed.
- [x] A reload that returns true lets the existing `syncImage` pick up the
      new generation (`internal/window/sync.go:12`). No new sync path.
- [x] Print a reload error to stderr once per distinct error, in the shape
      `hot reload: <path>: <error>`, and keep drawing the last good frame. Do
      not log a successful reload.
- [x] Test with the `fakeScreen` pattern
      (`internal/window/fake_screen_test.go:9`): a fake reloader that reports
      a change once, a shell whose poll interval has passed, and one `Update`
      that redraws exactly once.

Exit: `TestHotReloadPollsOnTheInterval` and
`TestHotReloadWaitsForTheInterval` pass in `internal/window`.

## Phase 4: state survives the reload

- [x] Keep the focused control when its id is still in `Boxes()`; clear it
      otherwise. The handler receives no Change callback for a hot reload;
      this is not an edit.
- [x] Keep hover and active the same way. Re-resolve both from the last
      cursor position the shell saw, because a frozen `:hover` from a
      mid-drag state is wrong after the page is replaced.
- [x] Keep the scroll offset when the content still overflows; clamp it
      otherwise.
- [x] Keep `SetImage` entries. A reloaded template that still references the
      same `src` finds the same bytes.
- [x] A page with a running `SetTick` keeps its tick. The reload swaps the
      template only.
- [x] Test: type into a field, reload the file with an extra paragraph, and
      assert the field still holds the text and the focus.

Exit: typing survives a reload, and a field removed from the new template
loses focus without an error.

## Phase 5: Serve

- [x] `Serve` has no frame loop, so `internal/web` calls the reloader before
      it answers `GET /` and `GET /frame.png`, under the same mutex the
      handlers already take (`internal/web/server.go:28`). The second request
      in a session sees the edit.
- [x] The shell page refreshes the image while a watch is active:
      `internal/web/shell.go` checks a `Reload bool` on its data and, when it
      is set, a short script reloads `/frame.png` on an interval. The browser
      shows the new picture without a manual refresh. The script belongs to
      the shell page, not to a template, so the no-JavaScript rule for pages
      is untouched.
- [x] Do not poll on every request when nothing is watched. `Reload` is false
      for a string page, and the shell emits no script.
- [x] `documentation/web.md` gets the behavior and the interval.

Exit: run `-web` on a file-backed example, edit the file, and the open
browser tab shows the new page within the interval.

## Phase 6: the example and the flag

- [x] New `examples/reload`: one `index.html` on disk with a counter, a text
      field, and a long paragraph. `main.go` takes `-reload` (default true)
      and passes `DisableHotReload: !*reload`, next to the existing `-web`
      and `-addr` flags. `-web` on port 8123.
- [x] The example's `New` reads the file through `Config.File`, so tests can
      point it at a temp copy. Keep the test version string-based.
- [x] Put the `-reload` pattern at the top of `examples/readme.md` as the dev
      loop: `go run ./examples/reload`, edit, watch.
- [x] The game examples keep their embed for the shipped build. Their
      `main.go` can pass a `Config.File` under the same `-reload` flag when a
      dev path exists. That is per example and optional, not a required edit.
- [x] New `documentation/hot-reload.md`: the config fields, the poll
      interval, the parse-error behavior, what survives a reload, the
      `host.Reloader` interface, and the wasm and mobile limits.
- [x] `documentation/README.md` table row, and a paragraph in
      `documentation/window.md` next to the resize behavior.
- [x] `documentation/features.md` gets a Hot reload section.
- [x] `documentation/platforms.md`: wasm and mobile read the file at startup
      only.
- [x] Add the new files to the map in `../../AGENTS.md`, and record what
      shipped in `../PHASES.md`.

Exit: `go run ./examples/reload`, edit `index.html` in place, and the counter
keeps its value while the paragraph changes.

## Phase 7: prove it

- [x] A poll with no file change does not redraw. Count redraws on the page.
- [x] A poll with one changed byte redraws exactly once.
- [x] A file deleted mid-run keeps the last good picture and reports the stat
      error once per change. A file recreated later reloads.
- [x] A file that begins invalid and becomes valid reloads on the fix, with
      no restart. This is the pending retry.
- [x] Two edits inside one poll interval produce one reload of the second
      bytes.
- [x] `GOOS=js GOARCH=wasm go build ./...` still builds.
- [x] `make test` and `go vet ./...` pass.

Exit: the tests above exist and pass, and the two limits, no Go-source watch
and no wasm watch, are written in the guide.

## Risks and limits

- This watches the files a page reads: the template and the theme. It does
  not recompile Go. A changed handler or view struct still needs a restart.
  `air` or a small wrapper around `go run` is the tool for that, outside this
  library.
- `Config.HTML` and `Config.File` both set is an error. Silent precedence is
  how a developer edits the wrong file for an hour.
- Polling sees a file by path, so a write into the file in place is picked
  up, and so is an editor that writes a new inode over the path. A network
  share that updates content without changing mtime or size is not. The
  pending retry covers the torn-write case.
- Reading on every `Serve` request adds a stat per request. At two watched
  paths that is nothing next to a PNG encode.
- A reload during a drag applies at the pending size; the generation and
  size guard in `sync.go` rebuilds the frame when the resize commit lands,
  so the two do not fight. The reload does not wait for the commit.
- `DisableHotReload` is a negative field because the default is on and a Go
  zero value has to mean the default. Read it twice before renaming it.
- A page that watches files behaves differently from a test page, which
  reads a string. Keep every test on `Config.HTML` unless the test is about
  the watch, so the suite never depends on the working directory.

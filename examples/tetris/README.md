# Tetris

A playable falling-block game on one ownframe page. The board is a grid
of retained display operations, so a frame moves fills and text runs
without parsing the HTML again. Scores, settings, replays, and one resume
slot live in a local SQLite database.

Run it from `examples/`:

```
go run ./tetris
```

Flags:

- `-data DIR` overrides the database directory.
- `-seed N` fixes the piece sequence. `0` uses the clock.
- `-perf` records pipeline timing and turns on the window's frame sampling for the Frame tab and `-debug-addr`.
- `-debug-addr ADDR` serves window profiling on a loopback address (off by default; needs `-tags ownframe_debug`).

## Controls

| Key | Action |
| --- | --- |
| Left, A | Move left |
| Right, D | Move right |
| Down, S | Soft drop |
| Up, W, X | Rotate clockwise |
| Z | Rotate counter-clockwise |
| Space | Hard drop |
| P, Escape | Pause or resume |
| R | Restart |
| Enter | Start, or restart after a top-out |
| H | Open the score history |
| T | Switch light and dark |

Movement, rotation, and soft drop repeat in Go after a short delay,
because the window forwards one key event per press and drops the OS
auto-repeat pulses. Focus loss releases every held key and pauses a
running game. A frame gap of half a second or more pauses too, so a
system sleep does not fast-forward the board.

The history screen pages live rankings 20 rows at a time, or switches to
the seeded demo entries with the DEMO button. Opening it pauses the game
and closing it resumes, with the board untouched.

## Storage

The database is `os.UserConfigDir()/ownframe/tetris/tetris.db`, so on
Linux `~/.config/ownframe/tetris/tetris.db`. `-data` replaces the
directory. The store opens with WAL on local files, runs migrations in
one transaction, and seeds 12 demo scores once behind a seed-version
marker. Reopening never duplicates fixtures or overwrites a live score.

If the directory cannot be opened, the game falls back to an in-memory
database and the status line reads `STORAGE TEMPORARY - SCORES NOT
KEPT`. Nothing is deleted and the next run retries the real path.

Scores are written on game over in one transaction with the run's replay
and a stable game ID, so a retry cannot insert the same score twice. A
failed write leaves the game playable, shows `SCORE NOT SAVED -
RETRYING`, and tries again every two seconds. Demo entries appear only
in the DEMO list and every demo row carries a DEMO tag.

A running or paused game stores one resume snapshot, on pause and on
window close. The snapshot clears on restart and on game over. A
snapshot found at startup is restored paused with `GAME RESUMED` in the
status line.

## Rendering

The game page keeps a retained display list. Board cells, the next-piece
preview, the score, level, lines, status, and theme caption change in
place each frame. A resize, a theme switch, or a phase overlay (ready,
paused, game over) rebuilds the page. If the engine rejects the display
list and paints a bitmap instead, the tick rebuilds at most every 50 ms.

The window has a minimum size of 680x740 CSS pixels and does not scale
the board, so the whole board stays visible at any size.

## Performance

Measured on 2026-10-07, 13th Gen Intel Core i7-13700HX, headless Linux,
Go 1.26.4, with:

```
go test -run XXX -bench 'BenchmarkTick|BenchmarkRedraw' -benchtime=2000x ./tetris/scene/
```

| Benchmark | ns/op | What it does |
| --- | --- | --- |
| `BenchmarkTickPaint` | 2579 | One active frame: fixed step, board and text paint, frame diff |
| `BenchmarkRedraw` | 6102239 | Full template and layout rebuild |
| `BenchmarkTickBitmap` | 1373298 | Fallback path average with the 50 ms throttle |

A paint-only frame costs about 2.6 microseconds of scene work. The
rebuild is the expensive path at about 6 ms, and it only runs on a
resize, a theme change, or a phase change. These numbers do not include
the window's replay draw; that needs a native session.

## Platform notes

This example is desktop only, through `ownframe.Run`. There is no `-web`
mode. Window scaling, monitor moves, minimize and restore, and wheel
input on Windows still need a native test session.

Closing the window stops the frame callback, saves a resume point, waits
up to half a second for it to reach the store, then joins the store
worker within two seconds.

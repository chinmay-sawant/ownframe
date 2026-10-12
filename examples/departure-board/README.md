# Departure board

A split-flap departure board in one ownframe page. Eight rows each show a time,
a destination, a platform, and a status as per-character tiles. Every 12 seconds
the board advances to the next slice of a seeded timetable, and the characters
that changed flap through random glyphs before settling.

## Running

From the repo root:

```
go run ./examples/departure-board
```

That opens the board in a desktop window. Pass `-web` to serve it instead on
127.0.0.1:8137. Pass `-livery` with `amber`, `flap`, `lcd`, or `paper` to pick
the starting livery; `amber` is the default.

## Controls

The header has one button per livery, a night button, and a pause button. The
four liveries are `amber`, `flap`, `lcd`, and `paper`. The switch changes only
the stylesheet, so the layout does not move and the board repaints.

The night button toggles the night variant of the current livery, which also
applies on its own between 20:00 and 06:00 local time. The pause button stops
the 12 second timetable rotation.

## How the flip works

`App.Tick` runs once per frame through `Page.SetTick`. The flipper binds the
text operation of every cell once per `Page.Generation()`, and after that a
flap frame writes only `op.Text`, so a flap never parses or lays the page out
again. The flipper finds cell text with `frame.Text`, which matches a box and
is blind to colour. The refresh bar is a fill, and `frame.Fill` matches the
fill colour, so that lookup needs the accent of the current livery.

## Performance

The board opts into the frame-dirty path, so a tick repaints one rect instead
of the whole 631 to 640 operation list. `PERFORMANCE.md` has the before and
after dump numbers, how to read them, and the open follow-ups. To measure it
yourself:

```
go run ./examples/departure-board -dump -pace 16ms -out /tmp/board-dump
```

The `-pace 16ms` flag reproduces the window's 60 ticks per second. Without it
the dump runs flat out, and the wall-clock driven refresh bar barely moves.

## Web

`Serve` never ticks, and `Page.PNG` paints from the template source, so the
`-web` page and its `/frame.png` show the last `Redraw` and not the flips.

## Tests

From the repo root:

```
go test -p 1 ./examples/departure-board/...
```

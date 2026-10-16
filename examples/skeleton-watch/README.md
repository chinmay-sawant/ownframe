# Mechanical Skeleton Watch & Tourbillon

A high-beat mechanical skeleton watch featuring a 1-RPM tourbillon escapement,
a dual-pusher chronograph, and a 72-hour power reserve indicator built on ownframe.

## Running

From the repository root:

```bash
go run ./examples/skeleton-watch
```

This opens the watch in a native desktop window. Pass `-web` to serve the preview
in a browser on port 8143 instead:

```bash
go run ./examples/skeleton-watch -web
```

## Controls

The case features interactive pushers, winding crown, and keyboard shortcuts:

- **Top Pusher (`pusher-start` / `Space`)**: Starts and stops the chronograph.
- **Bottom Pusher (`pusher-reset` / `R`)**: Resets the chronograph to zero when stopped, or performs flyback restart when running.
- **Winding Crown (`crown` / `W`)**: Winds the mainspring barrel by 6 hours per stroke, up to the 72-hour maximum reserve.
- **Lume Button (`btn-lume` / `L`)**: Toggles Super-LumiNova phosphor night mode, illuminating hands, indices, and the tourbillon cage.

## Horology Mechanics

1. **High-Beat Escapement (28,800 vph / 4 Hz)**:
   The balance wheel oscillates at 4 Hz, producing 8 distinct beats per second.
   The seconds hand steps across 8 sub-second increments each second for authentic
   mechanical sweep movement.

2. **60-Second Tourbillon**:
   The tourbillon carriage rotates once every 60 seconds (1 RPM) at the 6 o'clock
   aperture, with an oscillating balance wheel and hairspring inside.

3. **Chronograph**:
   A column-wheel timing train driving a central sweep seconds hand and a
   30-minute counter at the 3 o'clock subdial, with split lap capture.

4. **Power Reserve**:
   The 9 o'clock subdial indicates remaining mainspring energy from 0 to 72 hours.
   Energy drains continuously as the watch runs. If the reserve reaches zero, the
   escapement stops and hands freeze until manually wound with the crown.

5. **Skeleton Gear Train**:
   Visible mainspring barrel, center wheel (1 RPH), third wheel, and fourth wheel
   (1 RPM) rotating at their respective gear ratios.

## Display List Animation

`App.Tick` runs once per frame through `Page.SetTick`. Instead of reparsing HTML or
triggering full layout cascades every frame, the movement updates rotation matrices
directly on retained `DisplayOp` operations via `op.SetXform()`. The window replays
the display list at 60 FPS with minimal CPU overhead.

## Tests

Run the headless unit test suite from the repository root:

```bash
go test -v -p 1 ./examples/skeleton-watch/...
```

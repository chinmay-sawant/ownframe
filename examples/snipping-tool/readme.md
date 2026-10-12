# Snipping tool

Desktop screen capture example for Windows and Linux. It takes a full-screen shot, crops it from numeric fields, and saves or copies the result.

## Run

```sh
go run ./examples/snipping-tool          # desktop window
go run ./examples/snipping-tool -web     # still preview on 127.0.0.1:8136
```

Port 8136 is the first free port after the table in `examples/readme.md`. Port 8134 looks free in that table but `examples/perf-stress` already serves on it, and 8135 belongs to `perf-complex`.

## Toolbar

The toolbar has three buttons.

- Capture takes a new full-screen shot and replaces the preview.
- Save writes the current cropped image as PNG.
- Copy puts the current cropped image on the OS clipboard.

Each button reports its result in the status line, including the capture source and the pixel size.

## Selection

Four numeric fields set the crop rectangle: X, Y, Width, and Height. Editing a field re-crops the last capture without taking a new shot. Values outside the image are clamped to the image bounds, and an empty or zero-size rectangle keeps the full capture.

## Save

Save writes a PNG named `snip-YYYYMMDD-HHMMSS.png` under `os.UserConfigDir()/ownframe/snipping-tool/`, so on Linux `~/.config/ownframe/snipping-tool/`. Saving the same crop twice writes two files because the timestamp differs. A failed write leaves the preview untouched and the status line names the error.

## Copy

Copy sends the cropped PNG bytes to the text clipboard through
`internal/clipboard`, which only handles text. The bytes round-trip
in memory and the tests read them back, but pasting as an image into
another OS app is not guaranteed. In web mode there is no OS
clipboard, so Copy reports the capture size in the status line instead
of writing.

## Capture fallback

Capture tries one OS tool at a time and uses the first one that runs.

- Linux tries `import -window root png:-` (ImageMagick, stdout), then `gnome-screenshot -f`, `scrot`, and `grim`, each to a temp PNG file.
- Windows runs a PowerShell capture built on `System.Windows.Forms` and `System.Drawing`.
- macOS runs `screencapture -x -t png` to a temp file as best effort.

When no tool is found or every tool fails, headless CI included, the example falls back to a generated demo image, a deterministic gradient plus grid (320x200 default), and the status line says the image is a demo. The crop, save, and copy paths work the same on a demo image.

## Platform limits

Desktop Windows and Linux through `ownframe.Run`. macOS capture is best effort through `screencapture` and is not verified here. There is no mobile bind, so `BindMobile` and the wasm build do not apply. Web mode is a still preview through `Serve`: it shows the last capture, it does not take a live screenshot, and it does not tick.

## Tests

Headless, no window and no OS capture tool needed. The tests inject a fake capturer and a memory clipboard.

```sh
go test -p 1 ./examples/snipping-tool/...
```

## Layout

- `main.go` parses the flags and opens the window or serves the preview.
- `snip/` owns the capturer seam, the crop, the save and copy paths, and the page.

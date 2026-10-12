# Theme studio

An editor beside a live preview. The textarea on the left holds a theme
stylesheet, and the gallery on the right paints from the custom properties
that stylesheet sets. Below the editor a token strip lists every
`--name: value` pair it finds, with a colour swatch when the value parses as a
colour. The studio chrome itself uses literal colours, so a theme repaints the
preview pane only.

## Running

From the repo root:

```
go run ./examples/theme-studio
```

That opens the studio in a desktop window. Pass `-web` to serve it instead on
127.0.0.1:8138, and `-addr` to listen elsewhere.

## Controls

- **Apply** sets the editor sheet as the page theme and writes the same bytes
  to the theme file.
- **Copy** puts the editor text on the OS clipboard.
- **Reset** reads the theme file back from disk into the editor and applies it.
- **Light**, **Dark**, **Amber**, and **Paper** load a built-in sheet into the
  editor and apply it.

## The live loop

Apply calls `Page.SetTheme` and saves the sheet to the theme file. The page
watches that file through `Config.ThemeFile`, so saving the same path in an
external editor also lands live, without going through the Apply button. The
page reads the file at startup too, so a theme saved while the app was closed
comes back on the next run.

## The theme file

The path is `os.UserConfigDir()/ownframe/theme-studio/theme.css`, which on
Linux is `~/.config/ownframe/theme-studio/theme.css`. The studio creates the
directory and seeds the file with the default sheet on first run.

## DevTools

F12 or Ctrl+Shift+I opens the devtools dock on the right edge, with Elements,
Frame, and Ops tabs. The dock is window chrome, so it never appears in the
page picture. Pressing F12 again closes it.

## Tests

From the repo root:

```
go test -p 1 ./examples/theme-studio/...
```

The tests are headless: they build the studio with `XDG_CONFIG_HOME` pointed
at a temp directory and never open a window or touch the desktop clipboard.

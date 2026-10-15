# File paths, printable output, command execution, and local data handling

## File loading and template paths

Location: `internal/page`

`Config.File` and `Config.ThemeFile` are read with `os.ReadFile`. Relative paths resolve against the process working directory, symlinks are followed, and neither path nor size is limited. Hot reload rereads the watched file when size or modification time changes and parses the new content as a template or CSS theme. `Serve` performs that reload before each page and frame response.

### Vulnerabilities

- Severity: Medium if an app derives paths from user input
  - `Config.File` and `Config.ThemeFile` accept arbitrary filesystem paths.
- Severity: Medium
  - Hot reload can replace trusted source or theme content whenever the watched file changes. A local process able to modify that file can alter rendered page content.
- Severity: Low
  - Stat and read use separate calls, so a file replacement can be observed between the two.
- Severity: Low
  - There is no maximum source size, so a giant watched file can exhaust memory.

### Valid scenario

An app sets `Config{File: "views/login.html"}` in trusted boot code.

### Invalid scenario

An app builds `Config.File` from a query parameter or database string. The caller can read and render arbitrary readable files.

### Recommended fixes

1. Document that path inputs are trusted application configuration.
2. Add helper validation guidance that resolves paths with `filepath.Clean` and checks a trusted root.
3. Reject symlink escapes for apps that need sandboxed files.
4. Enforce a reasonable source size limit.
5. Consider disabling hot reload in production apps.

## Crash reports and PDF output

Location: `internal/crash`, `internal/page/print.go`

Crash reports use directory mode `0755` and file mode `0644`, so other local users may read them. They can contains stack paths, panic text, and page titles. `Page.SavePDF` creates any caller supplied path without normalization.

### Vulnerabilities

- Severity: Medium
  - World readable crash reports.
- Severity: Low
  - Crash directory selection accepts any writable location.
- Severity: Low
  - `SavePDF` can create files in caller chosen locations.

### Valid scenario

An app with a private user data directory handles reports safely. `SavePDF` is used with a fixed application path.

### Invalid scenario

A shared machine lets another user read `.txt` crash reports under the default user config directory. An app passes user input to `SavePDF` and overwrites or creates a file.

### Recommended fixes

1. Create crash directories with mode `0700` and files with mode `0600`.
2. Validate or normalize crash directory input.
3. Document that `SavePDF` is a trusted path API or add an application side path check.

## OS command execution

Location: `internal/print`, `internal/filepick`

The helpers execute OS commands through `exec.CommandContext` argv values. Paths are not passed through shell parsers, but two interpolation points were fixed during this audit.

### Fixed during audit

- `internal/print/commands.go` previously escaped AppleScript quotes manually. A path with a newline could break out of the script string. The path is now passed as an `osascript` argument and read with `item 1 of argv`.
- `examples/snipping-tool/snip/capture_win.go` interpolated `path` into a PowerShell script. It now escapes single quotes by doubling them.

### Remaining notes

- `internal/print/commands.go` and `internal/filepick/pick_wsl_script.go` double single quotes before using values in PowerShell single quoted strings. Backticks and newlines do not gain meaning there.
- Helper binaries such as `lp`, `xdg-open`, `zenity`, and `kdialog` are found through `PATH`. A hostile environment can execute alternate binaries. This is expected for native helper tools.
- Paths may begin with `--` and be treated as options by a helper command. Fixed prefixes are currently used, but adding `--` separators would be a useful hardening step.
- Linux clipboard validation restricts `/tmp/.X11-unix/X<n>` access to numeric display values.

### Valid scenario

Print commands and file dialogs receive normal generated or dialog selected paths.

### Invalid scenario

A path such as `--rm` or a file name containing a shell or script metacharacter reaches a command that was built from shell text rather than argv. The two fixed interpolation points no longer behave that way.

## Drag and drop and clipboard

Location: `internal/window/drop.go`, `internal/page/form_focus*.go`, `internal/page/bind.go`

Drop data includes `host.Drop.Path`, which contains the absolute host path from the OS event. That value is not normalized and should be treated as sensitive. `Drop.Read` remains valid only during the handler callback and must not be called from another goroutine later. Clipboard reads happen only from Paste chords or context menu actions. Helper code strips line breaks before calling `Page.Paste`.

### Vulnerabilities

- Severity: Medium
  - `Drop.Path` can disclose absolute host paths to application code and, if logged or rendered, to users.
- Severity: Medium
  - `Page.Paste` writes bound struct fields through reflection. Only app controlled bindings are affected.
- Severity: Low
  - `Focus` can focus controls marked `tabindex="-1"` when invoked with an explicit ID. This mirrors common browser behavior and is not a privilege issue.
- Severity: Low
  - Devtools can copy operation text to the clipboard while the dock is open.

### Valid scenario

An app uses `Drop.Read` inside its drop handler and copies bytes into owned memory.

### Invalid scenario

An app stores `Drop.Path` and later renders it. A local or remote change in dropped files can then disclose filesystem layout.

### Recommended fixes

1. Document that `Drop.Path` is sensitive and desktop specific.
2. Copy dropped bytes inside the handler because `Read` is frame scoped.
3. Keep devtools disabled in shipped binaries.
4. Avoid logging or displaying raw path metadata.

## Network boundary for IPC

Location: `internal/ipc`

IPC is in process only. It has no connection, authentication, or namespacing boundary. The name of a channel is a capability only by obscurity. Any goroutine in one process can listen or replace a handler. `Serve` does not expose IPC.

### Vulnerabilities

- Severity: Medium
  - A panicking `Handle` propagates through `Request`. `Send` recovers panic.
- Severity: Medium
  - No timeout or cancellation enforcement beyond caller owned contexts. A slow handler blocks callers.
- Severity: Medium
  - Registration has no caps, which allows unbounded memory use from repeated channel registration.
- Severity: Low
  - Empty channel and channel name validation is minimal. Oversized channel names are accepted.

### Valid scenario

Two trusted packages in one process register handlers and exchange values.

### Invalid scenario

An untrusted library or loaded code registers on a well known channel such as `demo.double` and changes results returned to an app consumer.

### Recommended fixes

1. Recover panics in `Request` and return a documented error.
2. Document IPC as a trusted local convenience API, not an isolation mechanism.
3. Document that handlers must be non blocking.
4. Add registration limits or expose a registration budget if third party code can call the API.
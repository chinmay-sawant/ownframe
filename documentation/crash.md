# Crash reports

`Run` stores every failure in a report file and returns an error naming that file: a panic goes through the recover, and any startup or window error goes through the same writer with the error as the reason. `BindMobile` recovers a panic on its own goroutine. `Serve` does not. The process does not panic again, and the file is not uploaded. In a browser build the recover still runs and returns an error, but no file is written, because Go's wasm file operations return an error there.

```go
ownframe.SetCrashDir(dir)
path, err := ownframe.Report("Hello", "disk full")
```

`Report` is the same writer, for a caller that already has a reason and does not want to panic. `SetCrashDir` points both `Report` and a recovered panic at `dir`. It returns an error when `dir` is empty or unsafe, and stores a cleaned path.

The default folder is `ownframe/crashes` under `os.UserConfigDir`.
If that directory cannot be resolved, the folder is `ownframe-crashes` under
`os.TempDir`. `Write` creates or reuses that folder with mode `0700` and
writes each report with mode `0600`. `SetCrashDir` rejects empty or unsafe
paths before replacing a previously configured directory.

The file name starts with a UTC timestamp `20060102-150405`. A second write in that same second gets a numeric suffix. The body contains the title, the reason, `runtime.Version()`, and `debug.Stack()`.

A nil page passed to `Serve` or `BindMobile` still returns bare `ErrNilPage` from the existing prepare check. `Run` wraps it in the report error above, which keeps `ErrNilPage` for `errors.Is`. The recover does not call `Title` on a nil page.

A panic inside the Ebiten update loop is caught only when it unwinds through `Run` on that same goroutine.

The [crash example](../examples/crash) sets the folder to `crashes/`, writes a manual report from one button, and panics from another.

## Trust boundary

Crash reports are local diagnostics. They can contain stack traces, panic text,
page titles, and source paths. The default directory uses mode 0700 and files
use mode 0600, so other local users cannot read them. `Serve` does not recover
panics at all. `SetCrashDir` is a trusted path API and accepts any directory
the process can write.

Recommended protections:

- On shared machines, call `SetCrashDir` with a user private directory and
  prefer directory mode 0700 and file mode 0600.
- Do not pass titles or reasons derived from untrusted input if they could
  carry secrets.
- Keep crash directories off world readable or network mounts.
- Treat reports as possibly containing source details and avoid surfacing them
  to end users unless filtered.

See the security audit in
[plans/v0.0.2/security](../plans/v0.0.2/security/00-overview-threat-model-and-priority.md).

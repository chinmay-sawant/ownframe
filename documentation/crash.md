# Crash reports

`Run` and `BindMobile` recover a panic on their own goroutine. `Serve` does not. The recover writes one UTF-8 text file and returns an error. The process does not panic again, and the file is not uploaded.

```go
gpui.SetCrashDir(dir)
path, err := gpui.Report("Hello", "disk full")
```

`Report` is the same writer, for a caller that already has a reason and does not want to panic. `SetCrashDir` points both `Report` and a recovered panic at `dir`. An empty dir restores the default.

The default folder is `go-gpui/crashes` under `os.UserConfigDir`. If that directory cannot be resolved, the folder is `go-gpui-crashes` under `os.TempDir`.

The file name starts with a UTC timestamp `20060102-150405`. A second write in that same second gets a numeric suffix. The body contains the title, the reason, `runtime.Version()`, and `debug.Stack()`.

A nil page passed to `Run`, `Serve`, or `BindMobile` still returns `ErrNilPage` from the existing prepare check. The recover does not call `Title` on a nil page.

A panic inside the Ebiten update loop is caught only when it unwinds through `Run` on that same goroutine.

The [crash example](../examples/crash) sets the folder to `crashes/`, writes a manual report from one button, and panics from another.

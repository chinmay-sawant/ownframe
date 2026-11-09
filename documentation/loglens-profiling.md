# Profile Loglens on Windows

Use a fresh process for each repetition. Keep the executable, content viewport,
phase durations, warmup, and DevTools setting in the capture metadata. The
capture script records the executable's SHA-256 hash and launch arguments.
Keep the source revision and any uncommitted patches beside the binary.

The script controls the mouse and resizes the window. Run it in an approved
Windows desktop session and leave that session alone until capture finishes.
It stops if Loglens loses foreground focus. It does not attach to, stop, rename,
or replace an existing Loglens process. Review the script before running it.

From the OwnFrame checkout, build the named example in WSL:

```sh
GOOS=windows GOARCH=amd64 go build -p 1 -tags ownframe_debug \
  -o temp/loglens.exe ./examples/builder/loglens
go version -m temp/loglens.exe > temp/loglens-build-info.txt
```

The local `go.work` also includes `../blinkless` while these engine fixes are
unpublished. Both checkouts must be present. This workspace entry must be removed
when OwnFrame's module requirement can point to the published engine changes;
there is no new `replace` in either module's `go.mod`.

In Windows PowerShell, pass the Windows paths to the script and executable:

```powershell
.\scripts\windows-loglens-dump.ps1 -Exe C:\path\to\loglens.exe
.\scripts\windows-loglens-dump.ps1 -Exe C:\path\to\loglens.exe -DevTools
```

Each command defaults to three repetitions with ten seconds per phase and three
seconds of warmup. The phases are idle, sidebar hover, scroll, resize, navigation,
and idle after navigation. The script launches a new process for each repetition
and closes only that process. It saves partial-run failure metadata on error.

The content viewport defaults to 1440 by 900 CSS pixels. The DevTools run adds
340 pixels to the window width so the application keeps the same content width.
A viewport mismatch at startup fails capture. Resize alternates the width by
160 pixels; the other phases restore the initial dimensions and Overview screen.

Each phase saves process samples and debug state, a CPU profile, cumulative
allocation profiles before and after, and heap profiles before and after with
`gc=1`. The forced collections happen outside the measured interaction interval.
They are profiling operations, not an application memory-management policy.
The collector never sets a Go memory limit or calls `FreeOSMemory`.

Analyze the timestamped folder in WSL:

```sh
python3 scripts/loglens-report.py /mnt/c/Users/Acer/Downloads/loglens-dumps/TIMESTAMP \
  --out temp/loglens-report.json
```

The report rejects failed captures, reset counters, mixed configurations, and
phases with no evidence of their intended action. It reports each repetition
and medians across repetitions. Compare the same phases and settings across
binaries. Compare DevTools settings separately, keeping the content viewport fixed.

Inspect live heap ownership and allocations over one phase:

```sh
go tool pprof -inuse_space -top CAPTURE/run-1/navigation/heap-after.pb.gz
go tool pprof -alloc_space -top \
  -base CAPTURE/run-1/navigation/allocs-before.pb.gz \
  CAPTURE/run-1/navigation/allocs-after.pb.gz
```

`go tool pprof` reports allocation sites. It does not identify the pointer that
keeps an object alive. Collection tests and cache ownership checks establish
that separately.

## Read the measurements

| Measurement | Meaning |
| --- | --- |
| **Allocated bytes per second** | Difference in `runtime.total_alloc` divided by the runtime samples' elapsed time. |
| **Presents per second** | Difference in `frame.presents` divided by the debug snapshots' elapsed time. |
| **Allocated bytes per present** | Allocation delta divided by present delta. The two counters have independently sampled boundaries. |
| **Last sampled redraw allocation** | `stats.AllocFrame`; this is one sampled redraw, not every presented frame. |
| **Rolling p99** | Host Update and Draw work over the last 120 frames. This is not a phase-wide percentile, GPU completion time, or input latency. |
| **Go-managed memory** | `runtime.sys - runtime.heap_released`. This excludes native graphics allocations. |
| **Working set and private bytes** | Separate Windows process measurements. Neither is the live Go heap. |

`FullLayouts` counts full placement attempts. `HoverPatches` counts successful
scoped hover paths, including changes that matched no selector. `HoverFallbacks`
counts eligible retained hover changes that used full placement.
`HoverPatchErrors` counts refused or failed scoped attempts, including retries.
Parses and cascades alone do not establish whether a hover used scoped placement.

Keep profiler overhead in mind. State requests and CPU profiling consume resources.
A high FPS reading does not exclude a slow interaction between snapshots. Windows
process totals cannot establish a graphics-driver or native-stack breakdown.

## Local regression checks

```sh
go test -p 1 ./examples/builder/loglens \
  -run 'TestButtonHoverUsesScopedLayout|TestNavigationRetainedHeapPlateaus' -count=1 -v
go test -p 1 ./internal/page -run TestHoverSelectorsMatchFullLayout -count=1
go test -p 1 ./examples/builder/loglens -run '^$' \
  -bench 'BenchmarkLoglensResize|BenchmarkLoglensPointerHover' \
  -benchmem -benchtime=30x -count=3
python3 -m unittest discover -s scripts -p test_loglens_report.py
```

The navigation test disables motion to isolate document ownership, warms three
cycles, runs twelve more cycles, and checks that obsolete displays are collectible
and post-GC growth stays within 8 MiB. It does not measure native graphics memory.
Run `make graphics-check` for the supported Linux pixel checks. A passing Linux
run does not establish Windows input routing or hardware-driver behavior.

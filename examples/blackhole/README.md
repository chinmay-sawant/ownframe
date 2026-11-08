# Gargantua · Ownframe Black Hole

A general-relativistic Schwarzschild black hole and its accretion disk rendered in pure Go using Kage fragment shaders on the GPU, composited under Ownframe's HTML/CSS interface.

Every pixel's light ray is traced backward through curved spacetime around the black hole:
- **Gravitational Lensing**: Light bending allows the far side of the accretion disk to appear simultaneously above and below the event horizon shadow.
- **Relativistic Doppler Beaming**: Material rotating toward the observer blazes brighter and shifts blue, while receding material dims and reddens.
- **Einstein Rings & Sky Warping**: Background starfields and galactic dust smear into arcs around the photon sphere.
- **Three-Pass Post-Processing**: Relativistic scene pass $\to$ quarter-resolution Vogel spiral bloom extraction $\to$ chromatic lens dispersion, film grain, and vignetting.

---

## Architecture & Integration

Each frame is raymarched offscreen and composited through the normal page pipeline:
1. **Offscreen raymarch (`sceneRenderer.tick` in `game.go`)**:
   - Registered via `page.SetTick`; `main.go` serves it with `ownframe.Run` / `ownframe.Serve`.
   - Runs the three embedded Kage shaders (`blackhole.kage`, `bloom.kage`, `post.kage`) into an offscreen scene buffer scaled by dynamic resolution scaling (DRS).
   - JPEG-encodes the frame (quality 85), pushes it with `page.SetImage("blackhole-scene", ...)`, and calls `page.Redraw`.
2. **Ownframe HTML/CSS page**:
   - Black page shows the frame in `<div id="scene">` (background-image); the HUD (`hud-root`) composites on top through the standard shell paint path.
   - Real-time telemetry: FPS, resolution scale, distance ($r_s$), and orbit state.
   - Interactive control chips (`Reset Camera`, `Pause Orbit`, `Auto Quality`) wired through `ownframe.Handlers.Click` and fine-grained `PatchElement`/`PatchElements`.
3. **Camera & Physics (`camera.go`)**:
   - Zero-allocation vector arithmetic and orthonormal basis calculations.
   - Spherical coordinates with drag-to-orbit, momentum inertia, and exponential zoom.

Direct drawing to the screen cannot work here: anything the shaders paint before the page HUD is buried under the shell's opaque background fills and `BlendCopy` tiles on the next repaint, so the scene travels through the page image pipeline instead and the HUD is composited on top by the normal shell path.

---

## Running the Application

### 1. Desktop Native Window
```sh
go run ./examples/blackhole
```

### 2. Browser Preview Mode (`-web`)
Serves the rendered canvas and interactive shell over HTTP:
```sh
go run ./examples/blackhole -web
```
Open `http://localhost:8155` in your browser.
`-web` is a HUD + static-scene preview: without a desktop window there is no GPU loop, so the scene image stays static while the HUD telemetry stays live. For the live raymarched animation, run the desktop window instead.

### 3. Headless Performance Monitoring (`-bench`)
Runs the full headless performance measurement suite without opening an OS window:
```sh
go run ./examples/blackhole -bench
```

### 4. Go Unit Tests & Microbenchmarks
```sh
go test -v ./examples/blackhole
go test -run '^$' -bench . -benchmem ./examples/blackhole
```

---

## Headless Benchmark Results (i7-13700HX)

| Benchmark | Latency | Derived FPS | Memory | Allocations | Description |
|---|---|---|---|---|---|
| `BenchmarkCamera_Step` | **11.6 ns/op** | ~86.2 M fps | **0 B/op** | **0 allocs/op** | Camera orbit stepping & momentum decay |
| `BenchmarkCamera_Axes` | **79.5 ns/op** | ~12.5 M fps | **0 B/op** | **0 allocs/op** | Trigonometric orthonormal basis calculation |
| `BenchmarkAdaptive_Quality` | **65.5 ns/op** | ~15.2 M fps | **0 B/op** | **0 allocs/op** | Dynamic resolution scaling pacing loop |
| `BenchmarkPage_Testing_Tick` | **288.4 ns/op** | ~3.4 M fps | **0 B/op** | **0 allocs/op** | Ownframe testing host tick frame |
| `BenchmarkHeapGrowth` | **301.4 ns/op** | — | **0 B/op** | **0 allocs/op** | Steady-state tick memory leak guard |
| `BenchmarkHUD_Patch` | **259.7 µs/op** | ~3,850 fps | 439 KB/op | 492 allocs/op | Fine-grained element patching of 4 HUD metrics |

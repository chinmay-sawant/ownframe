package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
)

// profileDirPerm keeps profiling output out of group and other access.
const profileDirPerm = 0o750

func run(mode, out string) error {
	switch mode {
	case "initial", "cached", "data", "resize", "windowed":
	default:
		return fmt.Errorf("unknown mode %q", mode)
	}
	if err := os.MkdirAll(out, profileDirPerm); err != nil {
		return err
	}
	p, v, err := fixture(mode == "windowed")
	if err != nil {
		return err
	}
	if mode != "initial" {
		if err = redraw(p); err != nil {
			return err
		}
	}
	// The mode switch above allow-lists the profile file names.
	cpu, err := os.Create(filepath.Join(out, mode+"-cpu.pprof")) //nolint:gosec // reason: mode is allow-listed by the switch above, so only the out dir varies.
	if err != nil {
		return err
	}
	if err = pprof.StartCPUProfile(cpu); err != nil {
		_ = cpu.Close()
		return err
	}
	doc, np, err := measure(mode, p, v)
	pprof.StopCPUProfile()
	cerr := cpu.Close()
	if err != nil {
		return err
	}
	if cerr != nil {
		return cerr
	}
	p = np
	runtime.GC()
	heap, err := os.Create(filepath.Join(out, mode+"-heap.pprof")) //nolint:gosec // reason: mode is allow-listed by the switch above, so only the out dir varies.
	if err != nil {
		return err
	}
	werr := pprof.WriteHeapProfile(heap)
	cerr = heap.Close()
	runtime.KeepAlive(p)
	if werr != nil {
		return werr
	}
	if cerr != nil {
		return cerr
	}
	return writeDump(out, mode, doc, p)
}

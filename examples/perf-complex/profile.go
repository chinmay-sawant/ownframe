package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
)

func run(mode, out string) error {
	switch mode {
	case "initial", "cached", "data", "resize", "windowed":
	default:
		return fmt.Errorf("unknown mode %q", mode)
	}
	if err := os.MkdirAll(out, 0755); err != nil {
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
	cpu, err := os.Create(filepath.Join(out, mode+"-cpu.pprof"))
	if err != nil {
		return err
	}
	if err = pprof.StartCPUProfile(cpu); err != nil {
		cpu.Close()
		return err
	}
	doc, np, err := measure(mode, p, v)
	pprof.StopCPUProfile()
	cpu.Close()
	if err != nil {
		return err
	}
	p = np
	runtime.GC()
	heap, err := os.Create(filepath.Join(out, mode+"-heap.pprof"))
	if err != nil {
		return err
	}
	err = pprof.WriteHeapProfile(heap)
	heap.Close()
	runtime.KeepAlive(p)
	if err != nil {
		return err
	}
	return writeDump(out, mode, doc, p)
}

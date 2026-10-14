// Package dump measures departure board scenarios headless.
package dump

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"

	"github.com/chinmay-sawant/ownframe/examples/departure-board/board"
)

// RunFlag parses the -scenario list and calls Run.
func RunFlag(list, out, cpuPath string) error {
	fields := strings.FieldsFunc(list, func(r rune) bool { return r == ',' || r == ' ' })
	return Run(fields, out, cpuPath)
}

// Run measures scenarios (all when empty), writes <out>/<name>.json, profiles
// to cpuPath when set, and prints a row per scenario.
func Run(scenarios []string, out, cpuPath string) error {
	if len(scenarios) == 0 {
		scenarios = Scenarios()
	}
	if err := os.MkdirAll(out, 0755); err != nil {
		return err
	}
	if cpuPath != "" {
		f, err := os.Create(cpuPath)
		if err != nil {
			return err
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			return err
		}
		defer pprof.StopCPUProfile()
	}
	printHeader()
	for _, name := range scenarios {
		app, err := board.New()
		if err != nil {
			return err
		}
		rep, err := Measure(context.Background(), app, name)
		if err != nil {
			return err
		}
		body, err := json.MarshalIndent(rep, "", "  ")
		if err != nil {
			return err
		}
		path := filepath.Join(out, name+".json")
		if err := os.WriteFile(path, append(body, '\n'), 0644); err != nil {
			return err
		}
		printRow(rep)
	}
	return nil
}
